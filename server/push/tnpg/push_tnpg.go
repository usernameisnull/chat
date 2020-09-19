// Package tnpg implements push notification plugin for Tinode Push Gateway.
package tnpg

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/tinode/chat/server/push"
	"github.com/tinode/chat/server/push/fcm"
	"github.com/tinode/chat/server/store"
)

const (
	baseTargetAddress = "https://pushgw.tinode.co/"
	pushBatchSize     = 100
	subBatchSize      = 1000
	bufferSize        = 1024
)

var handler Handler

// Handler represents state of TNPG push client.
type Handler struct {
	input   chan *push.Receipt
	channel chan *push.ChannelReq
	stop    chan bool
	pushUrl string
	subUrl  string
}

type configType struct {
	Enabled   bool   `json:"enabled"`
	OrgName   string `json:"org"`
	AuthToken string `json:"token"`
}

// subUnsubReq is a request to subscribe/unsubscribe device IDs to channel (FCM topic).
type subUnsubReq struct {
	Channel string   `json:"channel"`
	Devices []string `json:"devices"`
	Unsub   bool     `json:"unsub"`
}

type tnpgResponse struct {
	// Push message response only.
	MessageID    string `json:"msg_id,omitempty"`
	ErrorMessage string `json:"errmsg,omitempty"`
	// Channel sub/unsub response only.
	Index int `json:"index,omitempty"`
	// Both push and sub/unsub response.
	ErrorCode string `json:"errcode,omitempty"`
}

type batchResponse struct {
	// Number of successfully sent messages.
	SuccessCount int `json:"sent_count"`
	// Number of failures.
	FailureCount int `json:"fail_count"`
	// Error code and message if the entire batch failed.
	FatalCode    string `json:"errcode,omitempty"`
	FatalMessage string `json:"errmsg,omitempty"`
	// Individual reponses in the same order as messages. Could be nil if the entire batch failed.
	Responses []*tnpgResponse `json:"resp,omitempty"`

	// Local values
	httpCode   int
	httpStatus string
}

// Error codes copied from https://github.com/firebase/firebase-admin-go/blob/master/messaging/messaging.go
const (
	internalError                  = "internal-error"
	invalidAPNSCredentials         = "invalid-apns-credentials"
	invalidArgument                = "invalid-argument"
	messageRateExceeded            = "message-rate-exceeded"
	mismatchedCredential           = "mismatched-credential"
	quotaExceeded                  = "quota-exceeded"
	registrationTokenNotRegistered = "registration-token-not-registered"
	senderIDMismatch               = "sender-id-mismatch"
	serverUnavailable              = "server-unavailable"
	thirdPartyAuthError            = "third-party-auth-error"
	tooManyTopics                  = "too-many-topics"
	unavailableError               = "unavailable-error"
	unknownError                   = "unknown-error"
	unregisteredError              = "unregistered-error"
)

// Init initializes the handler
func (Handler) Init(jsonconf string) error {
	var config configType
	if err := json.Unmarshal([]byte(jsonconf), &config); err != nil {
		return errors.New("failed to parse config: " + err.Error())
	}

	if !config.Enabled {
		return nil
	}

	if config.OrgName == "" {
		return errors.New("push.tnpg.org not specified.")
	}

	handler.pushUrl = baseTargetAddress + "push/" + config.OrgName
	handler.subUrl = baseTargetAddress + "sub/" + config.OrgName
	handler.input = make(chan *push.Receipt, bufferSize)
	handler.channel = make(chan *push.ChannelReq, bufferSize)
	handler.stop = make(chan bool, 1)

	go func() {
		for {
			select {
			case rcpt := <-handler.input:
				go sendPushes(rcpt, &config)
			case sub := <-handler.channel:
				go processSubscription(sub, &config)
			case <-handler.stop:
				return
			}
		}
	}()

	return nil
}

func postMessage(endpoint string, body interface{}, config *configType) (*batchResponse, error) {
	buf := new(bytes.Buffer)
	gzw := gzip.NewWriter(buf)
	err := json.NewEncoder(gzw).Encode(body)
	gzw.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+config.AuthToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var batch batchResponse
	var reader io.ReadCloser
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		reader, err = gzip.NewReader(resp.Body)
		if err == nil {
			defer reader.Close()
		}
	} else {
		reader = resp.Body
	}

	if err == nil {
		err = json.NewDecoder(reader).Decode(&batch)
	}
	resp.Body.Close()

	if err != nil {
		// Just log the error, but don't report it to caller. The push succeeded.
		log.Println("tnpg failed to decode response", err)
	}

	batch.httpCode = resp.StatusCode
	batch.httpStatus = resp.Status

	return &batch, nil
}

func sendPushes(rcpt *push.Receipt, config *configType) {
	messages := fcm.PrepareNotifications(rcpt, nil)

	n := len(messages)
	for i := 0; i < n; i += pushBatchSize {
		upper := i + pushBatchSize
		if upper > n {
			upper = n
		}
		var payloads []interface{}
		for j := i; j < upper; j++ {
			payloads = append(payloads, messages[j].Message)
		}
		resp, err := postMessage(handler.pushUrl, payloads, config)
		if err != nil {
			log.Println("tnpg push request failed:", err)
			break
		}
		if resp.httpCode >= 300 {
			log.Println("tnpg push rejected:", resp.httpStatus)
			break
		}
		if resp.FatalCode != "" {
			log.Println("tnpg push failed:", resp.FatalMessage)
			break
		}
		// Check for expired tokens and other errors.
		handlePushResponse(resp, messages[i:upper])
	}
}

func processSubscription(req *push.ChannelReq, config *configType) {
	su := subUnsubReq{
		Devices: fcm.DevicesForUser(req.Uid),
		Channel: req.Channel,
		Unsub:   req.Unsub,
	}
	if len(su.Devices) == 0 {
		return
	}
	if len(su.Devices) > subBatchSize {
		// It's extremely unlikely for a single user to have this many devices.
		su.Devices = su.Devices[0:subBatchSize]
		log.Println("tnpg: user", req.Uid.UserId(), "has more than", subBatchSize, "devices")
	}

	resp, err := postMessage(handler.subUrl, &su, config)
	if err != nil {
		log.Println("tnpg channel sub request failed:", err)
		return
	}
	if resp.httpCode >= 300 {
		log.Println("tnpg channel sub rejected:", resp.httpStatus)
		return
	}
	if resp.FatalCode != "" {
		log.Println("tnpg channel sub failed:", resp.FatalMessage)
		return
	}
	// Check for expired tokens and other errors.
	handleSubResponse(resp, req, su.Devices)
}

func handlePushResponse(batch *batchResponse, messages []fcm.MessageData) {
	if batch.FailureCount <= 0 {
		return
	}

	for i, resp := range batch.Responses {
		switch resp.ErrorCode {
		case "": // no error
		case messageRateExceeded, quotaExceeded, serverUnavailable, unavailableError, internalError, unknownError:
			// Transient errors. Stop sending this batch.
			log.Println("tnpg: transient failure", resp.ErrorMessage)
			return
		case mismatchedCredential, invalidArgument, senderIDMismatch, thirdPartyAuthError, invalidAPNSCredentials:
			// Config errors
			log.Println("tnpg: invalid config", resp.ErrorMessage)
			return
		case registrationTokenNotRegistered, unregisteredError:
			// Token is no longer valid.
			log.Println("tnpg: invalid token", resp.ErrorMessage)
			if err := store.Devices.Delete(messages[i].Uid, messages[i].DeviceId); err != nil {
				log.Println("tnpg: failed to delete invalid token", err)
			}
		default:
			log.Println("tnpg: unrecognized error", resp.ErrorMessage)
		}
	}
}

func handleSubResponse(batch *batchResponse, req *push.ChannelReq, devices []string) {
	if batch.FailureCount <= 0 {
		return
	}

	for _, resp := range batch.Responses {
		// FCM documentation sucks. There is no list of possible errors so no action can be taken but logging.
		log.Println("fcm sub/unsub error", resp.ErrorCode, req.Uid, devices[resp.Index])
	}
}

// IsReady checks if the handler is initialized.
func (Handler) IsReady() bool {
	return handler.input != nil
}

// Push returns a channel that the server will use to send messages to.
// If the adapter blocks, the message will be dropped.
func (Handler) Push() chan<- *push.Receipt {
	return handler.input
}

// Push returns a channel that the server will use to send messages to.
// If the adapter blocks, the message will be dropped.
func (Handler) Channel() chan<- *push.ChannelReq {
	return handler.channel
}

// Stop terminates the handler's worker and stops sending pushes.
func (Handler) Stop() {
	handler.stop <- true
}

func init() {
	push.Register("tnpg", &handler)
}

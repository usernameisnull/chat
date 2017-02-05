/******************************************************************************
 *
 *  Description :
 *    An isolated communication channel (chat room, 1:1 conversation) for
 *    usualy multiple users. There is no communication across topics.
 *
 *****************************************************************************/

package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/tinode/chat/server/push"
	"github.com/tinode/chat/server/store"
	"github.com/tinode/chat/server/store/types"
)

var UA_TIMER_DELAY = time.Second * 5

// Topic: an isolated communication channel
type Topic struct {
	// Еxpanded/unique name of the topic.
	name string
	// For single-user topics session-specific topic name, such as 'me', otherwise the same as 'name'.
	original string

	// Topic category
	cat types.TopicCat

	// TODO(gene): currenctly unused
	// If isProxy == true, the actual topic is hosted by another cluster member.
	// The topic should:
	// 1. forward all messages to master
	// 2. route replies from the master to sessions.
	// 3. disconnect sessions at master's request.
	// 4. shut down the topic at master's request.
	isProxy bool

	// Time when the topic was first created
	created time.Time
	// Time when the topic was last updated
	updated time.Time

	// Server-side ID of the last data message
	lastId int
	// If messages were hard-deleted, the ID of the last deleted meassage
	clearId int

	// Last published userAgent ('me' topic only)
	userAgent string

	// User ID of the topic owner/creator. Could be zero.
	owner types.Uid

	// Default access mode
	accessAuth types.AccessMode
	accessAnon types.AccessMode

	// Topic's public data
	public interface{}

	// Topic's per-subscriber data
	perUser map[types.Uid]perUserData
	// User's contact list (not nil for 'me' topic only).
	// The map keys are UserIds for P2P topics and grpXXX for group topics.
	perSubs map[string]perSubsData

	// Sessions attached to this topic
	sessions map[*Session]bool

	// Inbound {data} and {pres} messages from sessions or other topics, already converted to SCM. Buffered = 256
	broadcast chan *ServerComMessage

	// Channel for receiving {get}/{set} requests, buffered = 32
	meta chan *metaReq

	// Subscribe requests from sessions, buffered = 32
	reg chan *sessionJoin

	// Unsubscribe requests from sessions, buffered = 32
	unreg chan *sessionLeave

	// Track the most active sessions to report User Agent changes. Buffered = 32
	uaChange chan string

	// Channel to terminate topic  -- either the topic is deleted or system is being shut down. Buffered = 1.
	exit chan *shutDown
}

// perUserData holds topic's cache of per-subscriber data
type perUserData struct {
	// Timestamps when the subscription was created and updated
	created time.Time
	updated time.Time

	online int

	// Last t.lastId reported by user through {pres} as received or read
	recvId int
	readId int
	// Greatest ID of a soft-deleted message
	clearId int

	private interface{}
	// cleared     time.Time // time, when the topic was cleared by the user
	modeWant  types.AccessMode
	modeGiven types.AccessMode
	// P2p only:
	public interface{}
}

// perSubsData holds user's (on 'me' topic) cache of subscription data
type perSubsData struct {
	online bool
	// Uid of the other user for P2P topics, otherwise 0
	with types.Uid
}

// Session wants to leave the topic
type sessionLeave struct {
	// Session which initiated the request
	sess *Session
	// Leave and unsubscribe
	unsub bool
	// Topic to report success of failure on
	topic string
	// ID of originating request, if any
	reqId string
}

type shutDown struct {
	// Channel to report back completion of topic shutdown. Could be nil
	done chan<- bool
	// Topic is being deleted as opposite to total system shutdown
	del bool
}

type pushReceipt struct {
	rcpt   *push.Receipt
	uidMap map[types.Uid]int
}

func (t *Topic) run(hub *Hub) {

	log.Printf("Topic started: '%s'", t.name)

	keepAlive := TOPICTIMEOUT // TODO(gene): read keepalive value from the command line
	killTimer := time.NewTimer(time.Hour)
	killTimer.Stop()

	// 'me' only
	var uaTimer *time.Timer
	var currentUA string
	uaTimer = time.NewTimer(time.Minute)
	uaTimer.Stop()

	for {
		select {
		case sreg := <-t.reg:
			// Request to add a conection to this topic

			// The topic is alive, so stop the kill timer, if it's ticking. We don't want the topic to die
			// while processing the call
			killTimer.Stop()

			if err := t.handleSubscription(hub, sreg); err == nil {
				// give a broadcast channel to the connection (.read)
				// give channel to use when shutting down (.done)
				sreg.sess.subs[t.name] = &Subscription{
					broadcast: t.broadcast,
					done:      t.unreg,
					meta:      t.meta,
					uaChange:  t.uaChange}

				t.sessions[sreg.sess] = true

			} else if len(t.sessions) == 0 {
				// Failed to subscribe, the topic is still inactive
				killTimer.Reset(keepAlive)
			}
		case leave := <-t.unreg:
			// Remove connection from topic; session may continue to function

			if leave.unsub {
				// User wants to leave and unsubscribe.
				if err := t.replyLeaveUnsub(hub, leave.sess, leave.reqId, leave.topic); err != nil {
					log.Panicln(err)
					continue
				}

			} else {
				now := time.Now().UTC().Round(time.Millisecond)

				// Just leaving the topic without unsubscribing
				delete(t.sessions, leave.sess)

				pud := t.perUser[leave.sess.uid]
				pud.online--
				if t.cat == types.TopicCat_Me {
					mrs := t.mostRecentSession()
					if mrs == nil {
						// Last session
						mrs = leave.sess
					} else {
						// Change UA to the most recent live session and announce it. Don't block.
						select {
						case t.uaChange <- mrs.userAgent:
						default:
						}
					}
					// Update user's last online timestamp & user agent
					if err := store.Users.UpdateLastSeen(mrs.uid, mrs.userAgent, now); err != nil {
						log.Println(err)
					}
				} else if t.cat == types.TopicCat_Grp && pud.online == 0 {
					t.presPubChange(leave.sess.uid, "off")
				}

				t.perUser[leave.sess.uid] = pud

				if leave.reqId != "" {
					leave.sess.queueOut(NoErr(leave.reqId, leave.topic, now))
				}
			}

			// If there are no more subscriptions to this topic, start a kill timer
			if len(t.sessions) == 0 {
				killTimer.Reset(keepAlive)
			}

		case msg := <-t.broadcast:
			// Content message intended for broadcasting to recepients

			var pushRcpt *pushReceipt

			if msg.Data != nil {
				from := types.ParseUserId(msg.Data.From)

				// msg.sessFrom is not nil when the message originated at the client.
				// for internally generated messages, like invites, the akn is nil
				if msg.sessFrom != nil {
					userData := t.perUser[from]
					if userData.modeWant&userData.modeGiven&types.ModePub == 0 {
						msg.sessFrom.queueOut(ErrPermissionDenied(msg.id, t.original, msg.timestamp))
						continue
					}
				}

				if err := store.Messages.Save(&types.Message{
					ObjHeader: types.ObjHeader{CreatedAt: msg.Data.Timestamp},
					SeqId:     t.lastId + 1,
					Topic:     t.name,
					From:      from.String(),
					Head:      msg.Data.Head,
					Content:   msg.Data.Content}); err != nil {

					log.Printf("topic[%s].run: failed to save message: %v", t.name, err)
					msg.sessFrom.queueOut(ErrUnknown(msg.id, t.original, msg.timestamp))

					continue
				}

				t.lastId++
				msg.Data.SeqId = t.lastId

				if msg.id != "" {
					reply := NoErrAccepted(msg.id, t.original, msg.timestamp)
					reply.Ctrl.Params = map[string]int{"seq": t.lastId}
					msg.sessFrom.queueOut(reply)
				}

				pushRcpt = t.makePushReceipt(msg.Data)

				t.presPubMessageSent(t.lastId)

			} else if msg.Pres != nil {
				log.Printf("topic[%s].run: pres.src='%s' pres.with='%s' what='%s'", t.name, msg.Pres.Src, msg.Pres.With, msg.Pres.What)
				t.presProcReq(msg.Pres.Src, msg.Pres.With, (msg.Pres.What == "on"), msg.Pres.wantReply)
				if t.original != msg.Pres.Topic {
					// This is just a request for status, don't forward it to sessions
					continue
				}
			} else if msg.Info != nil {
				if msg.Info.SeqId > t.lastId {
					// Skip bogus read notification
					continue
				}

				if msg.Info.What == "read" || msg.Info.What == "recv" {
					uid := types.ParseUserId(msg.Info.From)
					pud := t.perUser[uid]
					var read, recv int
					if msg.Info.What == "read" {
						if msg.Info.SeqId > pud.readId {
							pud.readId = msg.Info.SeqId
							read = pud.readId
						} else {
							// No need to report stale or bogus read status
							continue
						}
					} else if msg.Info.What == "recv" {
						if msg.Info.SeqId > pud.recvId {
							pud.recvId = msg.Info.SeqId
							recv = pud.recvId
						} else {
							continue
						}
					}

					if pud.readId > pud.recvId {
						pud.recvId = pud.readId
						recv = pud.recvId
					}

					if err := store.Subs.Update(t.name, uid,
						map[string]interface{}{
							"RecvSeqId": pud.recvId,
							"ReadSeqId": pud.readId}); err != nil {

						log.Printf("topic[%s].run: failed to update SeqRead/Recv counter: %v", t.name, err)
						continue
					}

					t.presPubMessageCount(msg.sessSkip, 0, recv, read)

					t.perUser[uid] = pud
				}
			}

			// Broadcast the message. Only {data}, {pres}, {info} are broadcastable.
			// {meta} and {ctrl} are sent to the session only
			if msg.Data != nil || msg.Pres != nil || msg.Info != nil {

				var packet, _ = json.Marshal(msg)
				for sess := range t.sessions {
					if sess == msg.sessSkip {
						continue
					}

					select {
					case sess.send <- packet:
						// Update device map with the device ID which
						if pushRcpt != nil {
							i, ok := pushRcpt.uidMap[sess.uid]
							if ok {
								pushRcpt.rcpt.To[i].Delieved++
								if sess.deviceId != "" {
									pushRcpt.rcpt.To[i].Devices = append(pushRcpt.rcpt.To[i].Devices, sess.deviceId)
								}
							} else {
								log.Printf("data message sent to user who is not subscribed to topic")
							}
						}
					default:
						log.Printf("topic[%s].run: connection stuck, detaching", t.name)
						t.unreg <- &sessionLeave{sess: sess, unsub: false}
					}
				}
				if pushRcpt != nil {
					push.Push(pushRcpt.rcpt)
				}

			} else {
				// TODO(gene): remove this
				log.Panic("topic[%s].run: wrong message type for broadcasting", t.name)
			}

		case meta := <-t.meta:
			log.Printf("topic[%s].run: got meta message '%+#v' %x", t.name, meta, meta.what)

			// Request to get/set topic metadata
			if meta.pkt.Get != nil {
				// Get request
				if meta.what&constMsgMetaDesc != 0 {
					t.replyGetDesc(meta.sess, meta.pkt.Get.Id, false, meta.pkt.Get.Desc)
				}
				if meta.what&constMsgMetaSub != 0 {
					t.replyGetSub(meta.sess, meta.pkt.Get.Id, meta.pkt.Get.Sub)
				}
				if meta.what&constMsgMetaData != 0 {
					t.replyGetData(meta.sess, meta.pkt.Get.Id, meta.pkt.Get.Data)
				}
			} else if meta.pkt.Set != nil {
				// Set request
				if meta.what&constMsgMetaDesc != 0 {
					t.replySetDesc(meta.sess, meta.pkt.Set)
				}
				if meta.what&constMsgMetaSub != 0 {
					t.replySetSub(hub, meta.sess, meta.pkt.Set)
				}

			} else if meta.pkt.Del != nil {
				// Del request
				if meta.what == constMsgDelMsg {
					t.replyDelMsg(meta.sess, meta.pkt.Del)
				} else if meta.what == constMsgDelTopic {
					t.replyDelTopic(hub, meta.sess, meta.pkt.Del)
				}
			}
		case ua := <-t.uaChange:
			// process an update to user agent from one of the sessions
			currentUA = ua
			uaTimer.Reset(UA_TIMER_DELAY)

		case <-uaTimer.C:
			// Publish user agent changes after a delay
			t.presPubUAChange(currentUA)

		case <-killTimer.C:
			// Topic timeout
			log.Println("Topic timeout: ", t.name)
			hub.unreg <- &topicUnreg{topic: t.name}
			if t.cat == types.TopicCat_Me {
				uaTimer.Stop()
				t.presPubMeChange("off", currentUA)
			} else if t.cat == types.TopicCat_Grp {
				t.presPubTopicOnline("off")
			}
			return

		case sd := <-t.exit:
			// Handle two cases: topic is being deleted (del==true) or system shutdown (del==false, done!=nil).
			// FIXME(gene): save lastMessage value;
			if t.cat == types.TopicCat_Grp && sd.del {
				t.presPubTopicOnline("gone")
			}
			// Not publishing online/offline to deleted P2P topics

			// In case of a system shutdown don't bother with notifications.

			// Report completion back to sender, if 'done' is not nil.
			if sd.done != nil {
				sd.done <- true
			}
			return
		}
	}
}

// Session subscribed to a topic, created == true if topic was just created and {pres} needs to be announced
func (t *Topic) handleSubscription(h *Hub, sreg *sessionJoin) error {

	var getWhat = 0
	if sreg.pkt.Get != nil {
		getWhat = parseMsgClientMeta(sreg.pkt.Get.What)
	}

	if err := t.subCommonReply(h, sreg, (getWhat&constMsgMetaDesc != 0)); err != nil {
		return err
	}

	// New P2P topic is a special case. It requires an invite/notification for the second person
	if sreg.created && t.cat == types.TopicCat_P2P {

		log.Println("about to generate invite for ", t.name)
		for uid, pud := range t.perUser {
			if uid != sreg.sess.uid {
				if pud.modeWant&types.ModeBanned != 0 {
					break
				}

				var action types.InviteAction
				if pud.modeWant == types.ModeNone {
					action = types.InvJoin
				} else {
					action = types.InvInfo
				}
				log.Println("sending invite to ", uid.UserId())
				h.route <- t.makeInvite(uid, uid, sreg.sess.uid, pud.public, action,
					pud.modeWant, pud.modeGiven, nil)
				break
			}
		}
	}

	if sreg.loaded {
		// Load the list of contacts for presence notifications
		if t.cat == types.TopicCat_Me {
			if err := t.loadContacts(sreg.sess.uid); err != nil {
				log.Printf("hub: failed to load contacts for '%s'", t.name)
			}

			t.presPubMeChange("on", sreg.sess.userAgent)

		} else if t.cat == types.TopicCat_Grp {
			t.presPubTopicOnline("on")
		}
		// Not sending presence notifications for P2P topics
	}

	if getWhat&constMsgMetaSub != 0 {
		// Send get.sub response as a separate {meta} packet
		t.replyGetSub(sreg.sess, sreg.pkt.Id, sreg.pkt.Get.Sub)
	}

	if getWhat&constMsgMetaData != 0 {
		// Send get.data response as {data} packets
		t.replyGetData(sreg.sess, sreg.pkt.Id, sreg.pkt.Get.Data)
	}
	return nil
}

// subCommonReply generates a response to a subscription request
func (t *Topic) subCommonReply(h *Hub, sreg *sessionJoin, sendDesc bool) error {
	log.Println("subCommonReply ", t.name)

	var now time.Time
	// For newly created topics report topic creation time.
	if sreg.created {
		now = t.updated
	} else {
		now = time.Now().UTC().Round(time.Millisecond)
	}

	// The topic t is already initialized by the Hub

	var info, private interface{}
	var mode string

	if sreg.pkt.Set != nil {
		if sreg.pkt.Set.Sub != nil {
			if sreg.pkt.Set.Sub.User != "" {
				log.Println("subCommonReply: msg.Sub.Sub.User is ", sreg.pkt.Set.Sub.User)
				sreg.sess.queueOut(ErrMalformed(sreg.pkt.Id, t.original, now))
				return errors.New("user id must not be specified")
			}

			info = sreg.pkt.Set.Sub.Info
			mode = sreg.pkt.Set.Sub.Mode
		}

		if sreg.pkt.Set.Desc != nil && !isNullValue(sreg.pkt.Set.Desc.Private) {
			private = sreg.pkt.Set.Desc.Private
		}
	}
	if err := t.requestSub(h, sreg.sess, sreg.pkt.Id, mode, info, private, sreg.loaded); err != nil {
		log.Println("requestSub failed: ", err.Error())
		return err
	}

	pud := t.perUser[sreg.sess.uid]
	pud.online++
	if pud.online == 1 {
		if t.cat == types.TopicCat_Grp {
			// User just joined the topic, announce presence
			log.Printf("subCommonReply: user %s joined grp topic %s", sreg.sess.uid.UserId(), t.name)
			t.presPubChange(sreg.sess.uid, "on")
		}
	}

	t.perUser[sreg.sess.uid] = pud

	resp := NoErr(sreg.pkt.Id, t.original, now)
	// Report available access mode.
	resp.Ctrl.Params = map[string]string{"mode": (pud.modeGiven & pud.modeWant).String()}
	sreg.sess.queueOut(resp)

	if sendDesc {
		t.replyGetDesc(sreg.sess, sreg.pkt.Id, sreg.created, sreg.pkt.Get.Desc)
	}

	return nil
}

// User requests or updates a self-subscription to a topic
//	h 		- hub
//	sess 	- originating session
//  pktId 	- originating packet Id
//	want	- requested access mode
//	info 	- explanation info given by the requester
//	private	- private value to assign to the subscription
// Handle these cases:
// A. User is trying to subscribe for the first time (no subscription)
// B. User is already subscribed, just joining without changing anything
// C. User is responsing to an earlier invite (modeWant was "N" in subscription)
// D. User is already subscribed, changing modeWant
// E. User is accepting ownership transfer (requesting ownership transfer is not permitted)
func (t *Topic) requestSub(h *Hub, sess *Session, pktId string, want string, info,
	private interface{}, loaded bool) error {

	log.Println("topic.requestSub", t.name)

	now := time.Now().UTC().Round(time.Millisecond)

	// Parse the acess mode requested by the user
	var modeWant types.AccessMode
	var explicitWant bool
	if want != "" {
		log.Println("mode want explicit: ", want)
		modeWant.UnmarshalText([]byte(want))
		explicitWant = true
	}

	// If the user wants a self-ban, make sure it's the only change
	if modeWant.IsBanned() {
		modeWant = types.ModeBanned
	}

	// Vars for saving changes to access mode
	var updWant *types.AccessMode
	var updGiven *types.AccessMode

	// Check if it's an attempt at a new subscription to the topic. If so, save it to database
	userData, existingSub := t.perUser[sess.uid]
	if !existingSub {
		// User requested default access mode.
		// modeWant could still be ModeNone if the owner wants to manually approve every request
		if modeWant == types.ModeNone {
			if explicitWant {
				// The operation is invalid - user requested to clear access to topic which makes no sense.
				sess.queueOut(ErrMalformed(pktId, t.original, now))
				return errors.New("attempt to clear topic access")
			}

			modeWant = t.accessAuth
		}

		userData = perUserData{
			private:   private,
			modeGiven: t.accessAuth,
			modeWant:  modeWant,
		}

		// If it's a re-subscription to a p2p topic, set public
		if t.cat == types.TopicCat_P2P {
			// t.perUser contains just one element - other user
			for uid2, _ := range t.perUser {
				if user2, err := store.Users.Get(uid2); err != nil {
					log.Println(err.Error())
					sess.queueOut(ErrUnknown(pktId, t.original, now))
					return err
				} else {
					userData.public = user2.Public
				}
				break
			}
		}

		// Add subscription to database
		sub := &types.Subscription{
			User:      sess.uid.String(),
			Topic:     t.name,
			ModeWant:  userData.modeWant,
			ModeGiven: userData.modeGiven,
			Private:   userData.private,
		}

		if err := store.Subs.Create(sub); err != nil {
			log.Println(err.Error())
			sess.queueOut(ErrUnknown(pktId, t.original, now))
			return err
		}

	} else {
		var ownerChange bool
		// Process update to existing subscription. It could be an incomplete subscription for a new topic.

		// If user did not request a new access mode, copy one from cache
		if !explicitWant && modeWant == types.ModeNone {
			modeWant = userData.modeWant
		}

		if userData.modeGiven.IsOwner() {
			// Check for possible ownership transfer. Handle the following cases:
			// 1. Owner joining the topic without any changes
			// 2. Owner changing own settings
			// 3. Acceptance or rejection of the ownership transfer

			// Make sure the current owner cannot unset the owner flag or ban himself
			if t.owner == sess.uid && (!modeWant.IsOwner() || modeWant.IsBanned()) {
				log.Println("requestSub: owner attempts to unset the owner flag")
				sess.queueOut(ErrMalformed(pktId, t.original, now))
				return errors.New("cannot unset ownership or ban the owner")
			}

			// Ownership transfer
			ownerChange = modeWant.IsOwner() && !userData.modeWant.IsOwner()

			// The owner should be able to grant himself any access permissions
			// If ownership transfer is rejected don't upgrade
			if modeWant.IsOwner() && !userData.modeGiven.Check(modeWant) {
				userData.modeGiven |= modeWant
				updGiven = &userData.modeGiven
			}
		} else if modeWant.IsOwner() {
			// Ownership transfer can only be initiated by the owner
			sess.queueOut(ErrMalformed(pktId, t.original, now))
			return errors.New("non-owner cannot request ownership transfer")
		} else if userData.modeGiven.IsManager() && modeWant.IsManager() {
			// The sharer should be able to grant any permissions except ownership
			if !userData.modeGiven.Check(modeWant & ^types.ModeBanned) {
				userData.modeGiven |= (modeWant & ^types.ModeBanned)
				updGiven = &userData.modeGiven
			}
		}

		// Check if t.accessAuth is different now than at the last attempt to subscribe
		if userData.modeGiven == types.ModeNone && t.accessAuth != types.ModeNone {
			userData.modeGiven = t.accessAuth
			updGiven = &userData.modeGiven
		}

		// This is a request to change access to whatever is currently available
		if modeWant == types.ModeNone {
			modeWant = userData.modeGiven
		}

		// Access actually changed
		if userData.modeWant != modeWant {
			userData.modeWant = modeWant
			updWant = &modeWant
		}

		// Save changes to DB
		if updWant != nil || updGiven != nil {
			update := map[string]interface{}{}
			// FIXME(gene): gorethink has a bug which causes ModeXYZ to be saved as a string, converting to int
			if updWant != nil {
				update["ModeWant"] = int(*updWant)
			}
			if updGiven != nil {
				update["ModeGiven"] = int(*updGiven)
			}
			if err := store.Subs.Update(t.name, sess.uid, update); err != nil {
				sess.queueOut(ErrUnknown(pktId, t.original, now))
				return err
			}
			//log.Printf("requestSub: topic %s updated SUB: %+#v", t.name, update)
		}

		// No transactions in RethinkDB, but two owners are better than none
		if ownerChange {
			//log.Printf("requestSub: topic %s owner change", t.name)

			userData = t.perUser[t.owner]
			userData.modeGiven = (userData.modeGiven & ^types.ModeOwner)
			userData.modeWant = (userData.modeWant & ^types.ModeOwner)
			if err := store.Subs.Update(t.name, t.owner,
				// FIXME(gene): gorethink has a bug which causes ModeXYZ to be saved as a string, converting to int
				map[string]interface{}{
					"ModeWant":  int(userData.modeWant),
					"ModeGiven": int(userData.modeGiven)}); err != nil {
				return err
			}
			t.perUser[t.owner] = userData
			t.owner = sess.uid
		}
	}

	t.perUser[sess.uid] = userData

	// If the user is (self)banned from topic, no further action is needed
	if modeWant.IsBanned() {
		t.evictUser(sess.uid, false, sess)
		// FIXME(gene): need to send a reply to user
		return errors.New("self-banned access to topic")
	} else if userData.modeGiven.IsBanned() {
		sess.queueOut(ErrPermissionDenied(pktId, t.original, now))
		return errors.New("topic access denied")
	}

	// Don't report existing subscription (no value);
	// Don't report a newly loaded Grp topic because it's announced later
	if !existingSub && (t.cat == types.TopicCat_P2P || !loaded) {
		t.presTopicSubscribed(sess.uid, sess)
	} else if existingSub {
		log.Println("pres not published: existing sub")
	} else {
		log.Println("pres not published: topic just loaded")
	}

	// If requested access mode different from given:
	if !userData.modeGiven.Check(modeWant) {
		// Send req to approve to topic managers
		for uid, pud := range t.perUser {
			if pud.modeGiven&pud.modeWant&types.ModeShare != 0 {
				h.route <- t.makeInvite(uid, sess.uid, sess.uid, userData.public,
					types.InvAppr, modeWant, userData.modeGiven, info)
			}
		}
		// Send info to requester
		h.route <- t.makeInvite(sess.uid, sess.uid, sess.uid, nil,
			types.InvInfo, modeWant, userData.modeGiven, t.public)
	}

	return nil
}

// approveSub processes a request to initiate an invite or approve a subscription request from another user:
// Handle these cases:
// A. Manager is inviting another user for the first time (no prior subscription)
// B. Manager is re-inviting another user (adjusting modeGiven, modeWant is still "N")
// C. Manager is changing modeGiven for another user, modeWant != "N"
func (t *Topic) approveSub(h *Hub, sess *Session, target types.Uid, set *MsgClientSet) error {
	now := time.Now().UTC().Round(time.Millisecond)

	// Check if requester actually has permission to manage sharing
	if userData, ok := t.perUser[sess.uid]; !ok || !userData.modeGiven.IsManager() || !userData.modeWant.IsManager() {
		sess.queueOut(ErrPermissionDenied(set.Id, t.original, now))
		return errors.New("topic access denied")
	}

	// Parse the access mode granted
	var modeGiven types.AccessMode
	if set.Sub.Mode != "" {
		modeGiven.UnmarshalText([]byte(set.Sub.Mode))
	}

	// If the user is banned from topic, make sute it's the only change
	if modeGiven.IsBanned() {
		modeGiven = types.ModeBanned
	}

	// Make sure no one but the owner can do an ownership transfer
	if modeGiven.IsOwner() && t.owner != sess.uid {
		sess.queueOut(ErrPermissionDenied(set.Id, t.original, now))
		return errors.New("attempt to transfer ownership by non-owner")
	}

	var givenBefore types.AccessMode
	// Check if it's a new invite. If so, save it to database as a subscription.
	// Saved subscription does not mean the user is allowed to post/read
	userData, ok := t.perUser[target]
	if !ok {
		if modeGiven == types.ModeNone {
			if t.accessAuth != types.ModeNone {
				// Request to use default access mode for the new subscriptions.
				modeGiven = t.accessAuth
			} else {
				sess.queueOut(ErrMalformed(set.Id, t.original, now))
				return errors.New("cannot invite without giving any access rights")
			}
		}

		// Add subscription to database
		sub := &types.Subscription{
			User:      target.String(),
			Topic:     t.name,
			ModeWant:  types.ModeNone,
			ModeGiven: modeGiven,
		}

		if err := store.Subs.Create(sub); err != nil {
			sess.queueOut(ErrUnknown(set.Id, t.original, now))
			return err
		}

		userData = perUserData{
			modeGiven: sub.ModeGiven,
			modeWant:  sub.ModeWant,
			private:   nil,
		}

		t.perUser[target] = userData

	} else {
		// Action on an existing subscription (re-invite or confirm/decline)
		givenBefore = userData.modeGiven

		// Request to re-send invite without changing the access mode
		if modeGiven == types.ModeNone {
			modeGiven = userData.modeGiven
		} else if modeGiven != userData.modeGiven {
			userData.modeGiven = modeGiven

			// Save changed value to database
			if err := store.Subs.Update(t.name, sess.uid,
				map[string]interface{}{"ModeGiven": modeGiven}); err != nil {
				return err
			}

			t.perUser[target] = userData
		}
	}

	// The user does not want to be bothered, no further action is needed
	if userData.modeWant.IsBanned() {
		sess.queueOut(ErrPermissionDenied(set.Id, t.original, now))
		return errors.New("topic access denied")
	}

	// Handle the following cases:
	// * a ban of the user, modeGive.IsBanned = true (if user is banned no need to invite anyone)
	// * regular invite: modeWant = "N", modeGiven > 0
	// * access rights update: old modeGiven != new modeGiven
	if !modeGiven.IsBanned() {
		if userData.modeWant == types.ModeNone {
			// (re-)Send the invite to target
			h.route <- t.makeInvite(target, target, sess.uid, nil, types.InvJoin, userData.modeWant, modeGiven,
				set.Sub.Info)
		} else if givenBefore != modeGiven {
			// Inform target that the access has changed
			h.route <- t.makeInvite(target, target, sess.uid, nil, types.InvInfo, userData.modeWant, modeGiven,
				set.Sub.Info)
		}
	}

	// Has anything actually changed?
	if givenBefore != modeGiven {
		// inform requester of the change made
		h.route <- t.makeInvite(sess.uid, target, sess.uid, nil, types.InvInfo, userData.modeWant, modeGiven,
			map[string]string{"before": givenBefore.String()})
	}

	return nil
}

// replyGetDesc is a response to a get.desc request on a topic, sent to just the session as a {meta} packet
func (t *Topic) replyGetDesc(sess *Session, id string, created bool, opts *MsgGetOpts) error {
	now := time.Now().UTC().Round(time.Millisecond)

	// Check if user requested modified data
	updatedOnly := (opts == nil || opts.IfModifiedSince == nil || opts.IfModifiedSince.Before(t.updated))

	desc := &MsgTopicDesc{CreatedAt: &t.created}
	if !t.updated.IsZero() {
		desc.UpdatedAt = &t.updated
	}

	pud, full := t.perUser[sess.uid]

	if updatedOnly {
		if t.public != nil {
			desc.Public = t.public
		} else if full {
			// p2p topic
			desc.Public = pud.public
		}
	}

	if t.cat == types.TopicCat_P2P {
		for uid, _ := range t.perUser {
			if uid.Compare(sess.uid) != 0 {
				desc.With = uid.UserId()
				break
			}
		}
	}

	// Request may come from a subscriber (full == true) or a stranger.
	// Give subscriber a fuller description than to a stranger
	if full {
		if pud.modeGiven&pud.modeWant&types.ModeShare != 0 {
			desc.DefaultAcs = &MsgDefaultAcsMode{
				Auth: t.accessAuth.String(),
				Anon: t.accessAnon.String()}
		}

		desc.Acs = &MsgAccessMode{
			Want:  pud.modeWant.String(),
			Given: pud.modeGiven.String(),
			Mode:  (pud.modeGiven & pud.modeWant).String()}

		if updatedOnly {
			desc.Private = pud.private
		}

		desc.SeqId = t.lastId
		// Make sure reported values are sane:
		// t.clearId <= pud.clearId <= t.readId <= t.recvId <= t.lastId
		desc.ClearId = max(pud.clearId, t.clearId)
		desc.ReadSeqId = max(pud.readId, desc.ClearId)
		desc.RecvSeqId = max(pud.recvId, pud.readId)

		// When a topic is first created, its name may change. Report the new name
		if created {
			desc.Name = t.name
		}
	}

	sess.queueOut(&ServerComMessage{
		Meta: &MsgServerMeta{
			Id:        id,
			Topic:     t.original,
			Desc:      desc,
			Timestamp: &now}})

	return nil
}

// replySetDesc updates topic metadata, saves it to DB,
// replies to the caller as {ctrl} message, generates {pres} update if necessary
func (t *Topic) replySetDesc(sess *Session, set *MsgClientSet) error {
	now := time.Now().UTC().Round(time.Millisecond)

	assignAccess := func(upd map[string]interface{}, mode *MsgDefaultAcsMode) error {
		if auth, anon, err := parseTopicAccess(mode, types.ModeInvalid, types.ModeInvalid); err != nil {
			return err
		} else if auth&types.ModeOwner != 0 || anon&types.ModeOwner != 0 {
			return errors.New("default 'O' access is not permitted")
		} else {
			access := make(map[string]interface{})
			if auth != types.ModeInvalid {
				access["Auth"] = auth
			}
			if anon != types.ModeInvalid {
				access["Anon"] = anon
			}
			if len(access) > 0 {
				upd["Access"] = access
			}
		}
		return nil
	}

	assignPubPriv := func(upd map[string]interface{}, what string, p interface{}) (changed bool) {
		if isNullValue(p) {
			if upd[what] != nil {
				upd[what] = nil
				changed = true
			}
		} else {
			upd[what] = p
			changed = true
		}
		return
	}

	updateCached := func(upd map[string]interface{}) {
		if tmp, ok := upd["Access"]; ok {
			access := tmp.(map[string]interface{})
			if auth, ok := access["Auth"]; ok {
				t.accessAuth = auth.(types.AccessMode)
			}
			if anon, ok := access["Anon"]; ok {
				t.accessAnon = anon.(types.AccessMode)
			}
		}
		if public, ok := upd["Public"]; ok {
			t.public = public
		}
	}

	var err error
	var sendPres bool

	user := make(map[string]interface{})
	topic := make(map[string]interface{})
	sub := make(map[string]interface{})
	if set.Desc != nil {
		if t.cat == types.TopicCat_Me {
			// Update current user
			if set.Desc.DefaultAcs != nil {
				err = assignAccess(user, set.Desc.DefaultAcs)
			}
			if set.Desc.Public != nil {
				sendPres = assignPubPriv(user, "Public", set.Desc.Public)
			}
		} else if t.cat == types.TopicCat_Fnd {
			// User's own tags are sent as fnd.public. Assign them to user.Tags
			if set.Desc.Public != nil {
				if src, ok := set.Desc.Public.([]string); ok && len(src) > 0 {
					tags := make([]string, 0, len(src))
					if filterTags(tags, src) > 0 {
						// No need to send presence update
						assignPubPriv(user, "Tags", tags)
						t.public = tags
					}
				}
			}
		} else if t.cat == types.TopicCat_Grp && t.owner == sess.uid {
			// Update current topic
			if set.Desc.DefaultAcs != nil {
				err = assignAccess(topic, set.Desc.DefaultAcs)
			}
			if set.Desc.Public != nil {
				sendPres = assignPubPriv(topic, "Public", set.Desc.Public)
			}
		}

		if err != nil {
			sess.queueOut(ErrMalformed(set.Id, set.Topic, now))
			return err
		}

		if set.Desc.Private != nil {
			assignPubPriv(sub, "Private", set.Desc.Private)
		}
	}

	var change int
	if len(user) > 0 {
		err = store.Users.Update(sess.uid, user)
		change++
	}
	if err == nil && len(topic) > 0 {
		err = store.Topics.Update(t.name, topic)
		change++
	}
	if err == nil && len(sub) > 0 {
		err = store.Subs.Update(t.name, sess.uid, sub)
		change++
	}

	if err != nil {
		sess.queueOut(ErrUnknown(set.Id, set.Topic, now))
		return err
	} else if change == 0 {
		sess.queueOut(ErrMalformed(set.Id, set.Topic, now))
		return errors.New("set generated no update to DB")
	}

	// Update values cached in the topic object
	if private, ok := sub["Private"]; ok {
		pud := t.perUser[sess.uid]
		pud.private = private
		t.perUser[sess.uid] = pud
	}
	if t.cat == types.TopicCat_Me {
		updateCached(user)
	} else if t.cat == types.TopicCat_Grp && t.owner == sess.uid {
		updateCached(topic)
	}

	if sendPres {
		t.presPubChange(sess.uid, "upd")
	}

	sess.queueOut(NoErr(set.Id, set.Topic, now))

	// TODO(gene) send pres update

	return nil
}

// replyGetSub is a response to a get.sub request on a topic - load a list of subscriptions/subscribers,
// send it just to the session as a {meta} packet
func (t *Topic) replyGetSub(sess *Session, id string, opts *MsgGetOpts) error {
	now := time.Now().UTC().Round(time.Millisecond)

	var subs []types.Subscription
	var err error

	if t.cat == types.TopicCat_Me {
		// Fetch user's subscriptions, with Topic.Public denormalized into subscription
		subs, err = store.Users.GetTopics(sess.uid)
	} else if t.cat == types.TopicCat_Fnd {
		// Given a query provided in .private, fetch user's contacts
		if query, ok := t.perUser[sess.uid].private.([]interface{}); ok {
			if query != nil && len(query) > 0 {
				subs, err = store.Users.FindSubs(sess.uid, query)
			}
		}
	} else {
		// FIXME(gene): don't load subs from DB, use perUserData - it already contains subscriptions.
		subs, err = store.Topics.GetUsers(t.name)
	}

	if err != nil {
		sess.queueOut(ErrUnknown(id, t.original, now))
		return err
	}

	var ifModified time.Time
	var limit int
	if opts != nil {
		if opts.IfModifiedSince != nil {
			ifModified = *opts.IfModifiedSince
		}
		limit = opts.Limit
	}

	if limit <= 0 {
		limit = 1024
	}

	meta := &MsgServerMeta{Id: id, Topic: t.original, Timestamp: &now}
	if subs != nil && len(subs) > 0 {
		meta.Sub = make([]MsgTopicSub, 0, len(subs))
		for idx, sub := range subs {
			if idx == limit {
				break
			}
			// Check if the requester has provided a cut off date for updated ts of pub & priv fields.
			sendPubPriv := ifModified.IsZero() || sub.UpdatedAt.After(ifModified)

			var mts MsgTopicSub
			uid := types.ParseUid(sub.User)
			var clearId int
			if t.cat == types.TopicCat_Me {
				// Reporting user's subscriptions to other topics
				mts.Topic = sub.Topic
				mts.SeqId = sub.GetSeqId()
				// Report whatever is the greatest - soft - or hard- deleted id
				clearId = max(sub.GetHardClearId(), sub.ClearId)
				mts.ClearId = clearId
				mts.With = sub.GetWith()
				if mts.With != "" {
					mts.Online = t.perSubs[mts.With].online
				} else {
					mts.Online = t.perSubs[sub.Topic].online
				}

				mts.UpdatedAt = &sub.UpdatedAt
				lastSeen := sub.GetLastSeen()
				if !lastSeen.IsZero() {
					mts.LastSeen = &MsgLastSeenInfo{
						When:      &lastSeen,
						UserAgent: sub.GetUserAgent()}
				}
			} else {
				// Reporting subscribers to a group or a p2p topic
				mts.User = uid.UserId()
				clearId = max(t.clearId, sub.ClearId)
				if uid == sess.uid {
					// Report deleted messages for own subscriptions only
					mts.ClearId = clearId
				}
				if t.cat == types.TopicCat_Grp {
					pud := t.perUser[uid]
					mts.Online = pud.online > 0
				}
			}

			// Ensure sanity or ReadId and RecvId:
			mts.ReadSeqId = max(clearId, sub.ReadSeqId)
			mts.RecvSeqId = max(clearId, sub.RecvSeqId)
			mts.AcsMode = (sub.ModeGiven & sub.ModeWant).String()
			// Returning public and private only if they have changed since ifModified
			if sendPubPriv {
				mts.Public = sub.GetPublic()
				// Reporting private only if it's user's own supscription or
				// a synthetic 'private' in 'find' topic where it's a list of tags matched on.
				if uid == sess.uid || t.cat == types.TopicCat_Fnd {
					mts.Private = sub.Private
				}
			}

			meta.Sub = append(meta.Sub, mts)
		}
	}

	sess.queueOut(&ServerComMessage{Meta: meta})

	return nil
}

// replySetSub is a response to new subscription request or an update to a subscription {set.sub}:
// update topic metadata cache, save/update subs, reply to the caller as {ctrl} message, generate an invite
func (t *Topic) replySetSub(h *Hub, sess *Session, set *MsgClientSet) error {
	now := time.Now().UTC().Round(time.Millisecond)

	var uid types.Uid
	if uid = types.ParseUserId(set.Sub.User); uid.IsZero() && set.Sub.User != "" {
		// Invalid user ID
		sess.queueOut(ErrMalformed(set.Id, set.Topic, now))
		return errors.New("invalid user id")
	}

	// if set.User is not set, request is for the current user
	if uid.IsZero() {
		uid = sess.uid
	}

	if uid == sess.uid {
		// Request to modify current subscription
		return t.requestSub(h, sess, set.Id, set.Sub.Mode, set.Sub.Info, nil, false)
	} else {
		// Request to approve a subscription
		return t.approveSub(h, sess, uid, set)
	}
}

// replyGetData is a response to a get.data request - load a list of stored messages, send them to session as {data}
// response goes to a single session rather than all sessions in a topic
func (t *Topic) replyGetData(sess *Session, id string, req *MsgBrowseOpts) error {
	now := time.Now().UTC().Round(time.Millisecond)

	opts := msgOpts2storeOpts(req, t.perUser[sess.uid].clearId)

	messages, err := store.Messages.GetAll(t.name, opts)
	if err != nil {
		log.Println("topic: error loading topics ", err)
		sess.queueOut(ErrUnknown(id, t.original, now))
		return err
	}

	// Push the list of messages to the client as {data}.
	// Messages are sent in reverse order than fetched from DB to make it easier for
	// clients to process.
	if messages != nil {
		for i := len(messages) - 1; i >= 0; i-- {
			mm := messages[i]
			from := types.ParseUid(mm.From)
			msg := &ServerComMessage{Data: &MsgServerData{
				Topic:     t.original,
				SeqId:     mm.SeqId,
				From:      from.UserId(),
				Timestamp: mm.CreatedAt,
				Content:   mm.Content}}
			sess.queueOut(msg)
		}
	}

	return nil
}

// replyDelMsg deletes (soft or hard) messages in response to del.msg packet.
func (t *Topic) replyDelMsg(sess *Session, del *MsgClientDel) error {
	now := time.Now().UTC().Round(time.Millisecond)

	if del.Before > t.lastId || (del.Before <= 0 && del.List == nil) {
		sess.queueOut(ErrMalformed(del.Id, t.original, now))
		return errors.New("invalid del.msg parameter 'before'")
	}

	pud := t.perUser[sess.uid]
	if pud.modeGiven&pud.modeWant&types.ModeDelete == 0 {
		// User cannot hard-delete messages, silently switching to soft-deleting
		del.Hard = false
	}

	var err error
	if del.Before > 0 {
		// Make sure user has not deleted the messages already
		if (del.Before <= t.clearId) || (!del.Hard && del.Before <= pud.clearId) {
			sess.queueOut(InfoNoAction(del.Id, t.original, now))
			return nil
		}

		if t.cat == types.TopicCat_Me {
			err = store.Messages.Delete("", sess.uid, del.Hard, del.Before)
		} else {
			err = store.Messages.Delete(t.name, sess.uid, del.Hard, del.Before)
		}
	}

	if err != nil {
		sess.queueOut(ErrUnknown(del.Id, t.original, now))
		return err
	}

	if del.Hard {
		t.lastId = t.lastId
		t.clearId = del.Before

		// Broadcast the change
		t.presPubMessageDel(sess, t.clearId)
	} else {
		pud.clearId = del.Before
		if pud.readId < pud.clearId {
			pud.readId = pud.clearId
		}
		if pud.recvId < pud.readId {
			pud.recvId = pud.readId
		}
		t.perUser[sess.uid] = pud

		// Notify user's other sessions
		t.presPubMessageCount(sess, pud.clearId, 0, 0)
	}

	sess.queueOut(NoErr(del.Id, t.original, now))

	return nil
}

// 1. Checks if the requester is the owner. If so:
// 1.2 Evict all sessions
// 1.3 Ask hub to unregister self
// 1.4 Exit the run() loop
// 2. If requester is not the owner:
// 2.1 If this is a p2p topic:
// 2.1.1 Check if the other sbscription still exists, if so, treat request as {leave unreg=true}
// 2.1.2 If the other subscription does not exist, delete topic
// 2.2 If this is not a p2p topic, treat it as {leave unreg=true}
func (t *Topic) replyDelTopic(h *Hub, sess *Session, del *MsgClientDel) error {
	if t.owner != sess.uid {
		// Cases 2.1.1 and 2.2
		if t.cat != types.TopicCat_P2P || len(t.perUser) > 1 {
			log.Println("delTopic: not owner, just unsubscribing")
			return t.replyLeaveUnsub(h, sess, del.Id, t.original)
		}
	}

	log.Println("delTopic: owner or last p2p subscription, evicting all sessions")
	t.evictAll()

	log.Println("delTopic: owner, requesting topic removal from hub")
	h.unreg <- &topicUnreg{
		topic: t.name,
		sess:  sess,
		msg:   del,
		del:   true}

	return nil
}

func (t *Topic) replyLeaveUnsub(h *Hub, sess *Session, id, topic string) error {
	now := time.Now().UTC().Round(time.Millisecond)

	// Delete user's subscription from the database
	if err := store.Subs.Delete(t.name, sess.uid); err != nil {
		if id != "" {
			sess.queueOut(ErrUnknown(id, topic, now))
		}

		return err
	}

	// Evict all user's sessions and clear cached data
	t.evictUser(sess.uid, true, sess)

	if id != "" {
		sess.queueOut(NoErr(id, topic, now))
	}

	return nil
}

// Create a data message with an invite:
// notify - user whio will receive the message
// target - user who access rights are being changed
// from  - user who sent the request
// act - what's being done - request or an approval or a request
// mpodeWant, modeGiven - requested or granted access permissions
func (t *Topic) makeInvite(notify, target, from types.Uid, public interface{}, act types.InviteAction, modeWant,
	modeGiven types.AccessMode, info interface{}) *ServerComMessage {

	// FIXME(gene): this is a workaround for gorethink's broken way of marshalling json.
	// The data message will be saved to DB.
	inv, err := json.Marshal(MsgInvitation{
		Topic:  t.name,
		User:   target.UserId(),
		Public: public,
		Action: act.String(),
		Acs: MsgAccessMode{
			Want:  modeWant.String(),
			Given: modeGiven.String(),
			Mode:  (modeWant & modeGiven).String()},
		Info: info})
	if err != nil {
		log.Println(err)
	}
	converted := map[string]interface{}{}
	err = json.Unmarshal(inv, &converted)
	if err != nil {
		log.Println(err)
	}
	// endof workaround

	msg := &ServerComMessage{Data: &MsgServerData{
		Topic:     "me",
		From:      from.UserId(),
		Timestamp: time.Now().UTC().Round(time.Millisecond),
		Content:   converted}, rcptto: notify.UserId()}
	log.Printf("Invite generated: %#+v", msg.Data)
	return msg
}

// evictUser evicts given user's sessions from the topic and clears user's cached data, if requested
func (t *Topic) evictUser(uid types.Uid, unsub bool, ignore *Session) {
	now := time.Now().UTC().Round(time.Millisecond)
	note := NoErrEvicted("", t.original, now)

	// First notify topic subscribers that the user has left the topic
	if t.cat == types.TopicCat_Grp {
		log.Println("del: announcing GRP")
		if unsub {
			t.presPubChange(uid, "unsub")
			t.presTopicGone(uid)
		} else {
			t.presPubChange(uid, "off")
		}
	} else if t.cat == types.TopicCat_P2P && unsub {
		log.Println("del: announcing P2P")
		t.presTopicGone(uid)
	} else {
		log.Println("del: not announcing", t.cat, unsub)
	}

	// Second - detach user from topic
	if unsub {
		// Delete per-user data
		delete(t.perUser, uid)
	} else {
		// Clear online status
		pud := t.perUser[uid]
		pud.online = 0
		t.perUser[uid] = pud
	}

	// Detach all user's sessions
	for sess, _ := range t.sessions {
		if sess.uid == uid {
			delete(t.sessions, sess)
			sess.detach <- t.name
			if sess != ignore {
				sess.queueOut(note)
			}
		}
	}
}

func (t *Topic) makePushReceipt(data *MsgServerData) *pushReceipt {
	idx := make(map[types.Uid]int, len(t.perUser))
	receipt := push.Receipt{
		To: make([]push.PushTo, len(t.perUser)),
		Payload: push.Payload{
			Topic:     data.Topic,
			From:      data.From,
			Timestamp: data.Timestamp,
			SeqId:     data.SeqId,
			Content:   data.Content}}

	i := 0
	for uid, _ := range t.perUser {
		receipt.To[i].User = uid
		idx[uid] = i
		i++
	}

	return &pushReceipt{rcpt: &receipt, uidMap: idx}
}

// evictAll disconnects all sessions from the topic
func (t *Topic) evictAll() {
	note := NoErrEvicted("", t.original, time.Now().UTC().Round(time.Millisecond))
	for sess, _ := range t.sessions {
		delete(t.sessions, sess)
		sess.detach <- t.name
		sess.queueOut(note)
	}
}

func (t *Topic) mostRecentSession() *Session {
	var sess *Session
	var latest time.Time
	for s, _ := range t.sessions {
		if s.lastAction.After(latest) {
			sess = s
			latest = s.lastAction
		}
	}
	return sess
}

// Takes get.data parameters and ClearID, returns database query parameters
func msgOpts2storeOpts(req *MsgBrowseOpts, clearId int) *types.BrowseOpt {
	var opts *types.BrowseOpt
	if req != nil || clearId > 0 {
		opts = &types.BrowseOpt{}
		if req != nil {
			opts.Limit = req.Limit
			opts.Since = req.Since
			opts.Before = req.Before
		}
		if clearId > opts.Since {
			// ClearId deletes mesages upto and including the value itself. Since shows message starting
			// with the value itself, thus must add 1 to make sure the last deleted message is not shown.
			opts.Since = clearId + 1
		}
	}
	return opts
}

func isNullValue(i interface{}) bool {
	// Del control character
	const CLEAR_VALUE = "\u2421"
	if str, ok := i.(string); ok {
		return str == CLEAR_VALUE
	}
	return false
}

func topicCat(name string) types.TopicCat {
	return types.GetTopicCat(name)
}

// Generate random string as a name of the group topic
func genTopicName() string {
	return "grp" + store.GetUidString()
}

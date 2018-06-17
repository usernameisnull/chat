// Package adapter contains the interfaces to be implemented by the database adapter
package adapter

import (
	"time"

	"github.com/tinode/chat/server/auth"
	t "github.com/tinode/chat/server/store/types"
)

// Adapter is the interface that must be implemented by a database
// adapter. The current schema supports a single connection by database type.
type Adapter interface {
	// General

	// Open and configure the adapter
	Open(config string) error
	// Close the adapter
	Close() error
	// IsOpen checks if the adapter is ready for use
	IsOpen() bool
	// CheckDbVersion checks if the actual database version matches adapter version.
	CheckDbVersion() error
	// GetName returns the name of the adapter
	GetName() string

	// CreateDb creates the database optionally dropping an existing database first.
	CreateDb(reset bool) error

	// User management

	// UserCreate creates user record
	UserCreate(usr *t.User) error
	// UserGet returns record for a given user ID
	UserGet(id t.Uid) (*t.User, error)
	// UserGetAll returns user records for a given list of user IDs
	UserGetAll(ids ...t.Uid) ([]t.User, error)
	// UserDelete deletes user record
	UserDelete(id t.Uid, soft bool) error
	// UserUpdate updates user record
	UserUpdate(uid t.Uid, update map[string]interface{}) error

	// Credential management

	// CredAdd adds credential record.
	CredAdd(cred *t.Credential) error
	// CredGet returns credential record.
	CredGet(uid t.Uid, method string) ([]*t.Credential, error)
	// CredIsConfirmed returns true if the given credential has been verified, false otherwise.
	CredIsConfirmed(uid t.Uid, metod string) (bool, error)
	// CredDel deletes given credential.
	CredDel(uid t.Uid, method string) error
	// CredConfirm marks given credential as validated.
	CredConfirm(uid t.Uid, method string) error
	// CredFail increments count of failed validation attepmts for the given credentials.
	CredFail(uid t.Uid, method string) error

	// Authentication management for the basic authentication scheme

	AuthGetUniqueRecord(unique string) (t.Uid, auth.Level, []byte, time.Time, error)
	AuthGetRecord(user t.Uid, scheme string) (string, auth.Level, []byte, time.Time, error)
	AuthAddRecord(user t.Uid, scheme, unique string, authLvl auth.Level, secret []byte, expires time.Time) (bool, error)
	AuthDelRecord(user t.Uid, unique string) error
	AuthDelAllRecords(uid t.Uid) (int, error)
	AuthUpdRecord(user t.Uid, scheme, unique string, authLvl auth.Level, secret []byte, expires time.Time) (bool, error)

	// Topic management

	// TopicCreate creates a topic
	TopicCreate(topic *t.Topic) error
	// TopicCreateP2P creates a p2p topic
	TopicCreateP2P(initiator, invited *t.Subscription) error
	// TopicGet loads a single topic by name, if it exists. If the topic does not exist the call returns (nil, nil)
	TopicGet(topic string) (*t.Topic, error)
	// TopicsForUser loads subscriptions for a given user. Reads public value.
	TopicsForUser(uid t.Uid, keepDeleted bool, opts *t.QueryOpt) ([]t.Subscription, error)
	// UsersForTopic loads users' subscriptions for a given topic. Public is loaded.
	UsersForTopic(topic string, keepDeleted bool, opts *t.QueryOpt) ([]t.Subscription, error)
	// TopicShare creates topc subscriptions
	TopicShare(subs []*t.Subscription) (int, error)
	// TopicDelete deletes topic, subscription, messages
	TopicDelete(topic string) error
	// TopicUpdateOnMessage increments Topic's or User's SeqId value and updates TouchedAt timestamp.
	TopicUpdateOnMessage(topic string, msg *t.Message) error
	// TopicUpdate updates topic record.
	TopicUpdate(topic string, update map[string]interface{}) error

	// Topic subscriptions

	// SubscriptionGet reads a subscription of a user to a topic
	SubscriptionGet(topic string, user t.Uid) (*t.Subscription, error)
	// SubsForUser gets a list of topics of interest for a given user. Does NOT load Public value.
	SubsForUser(user t.Uid, keepDeleted bool, opts *t.QueryOpt) ([]t.Subscription, error)
	// SubsForTopic gets a list of subscriptions to a given topic.. Does NOT load Public value.
	SubsForTopic(topic string, keepDeleted bool, opts *t.QueryOpt) ([]t.Subscription, error)
	// SubsUpdate updates pasrt of a subscription object. Pass nil for fields which don't need to be updated
	SubsUpdate(topic string, user t.Uid, update map[string]interface{}) error
	// SubsDelete deletes a single subscription
	SubsDelete(topic string, user t.Uid) error
	// SubsDelForTopic soft-deletes all subscriptions to the given topic
	SubsDelForTopic(topic string) error
	// SubsDelForUser soft-deletes all subscriptions of the given user
	SubsDelForUser(user t.Uid) error

	// Search

	// FindUsers searches for new contacts given a list of tags
	FindUsers(user t.Uid, req, opt []string) ([]t.Subscription, error)
	// FindTopics searches for group topics given a list of tags
	FindTopics(req, opt []string) ([]t.Subscription, error)

	// Messages

	// MessageSave saves message to database
	MessageSave(msg *t.Message) error
	// MessageGetAll returns messages matching the query
	MessageGetAll(topic string, forUser t.Uid, opts *t.QueryOpt) ([]t.Message, error)
	// MessageDeleteList marks messages as deleted.
	// Soft- or Hard- is defined by forUser value: forUSer.IsZero == true is hard.
	MessageDeleteList(topic string, toDel *t.DelMessage) error
	// MessageGetDeleted returns a list of deleted message Ids.
	MessageGetDeleted(topic string, forUser t.Uid, opts *t.QueryOpt) ([]t.DelMessage, error)

	// Devices (for push notifications)

	// DeviceUpsert creates or updates a device record
	DeviceUpsert(uid t.Uid, dev *t.DeviceDef) error
	// DeviceGetAll returns all devices for a given set of users
	DeviceGetAll(uid ...t.Uid) (map[t.Uid][]t.DeviceDef, int, error)
	// DeviceDelete deletes a device record
	DeviceDelete(uid t.Uid, deviceID string) error

	// File upload records. The files are stored outside of the database.

	// FileStartUpload initializes a file upload
	FileStartUpload(fd *t.FileDef) error
	// FileFinishUpload markes file upload as completed, successfully or otherwise.
	FileFinishUpload(fid string, status int, size int64) (*t.FileDef, error)
	// FileGet fetches a record of a specific file
	FileGet(fid string) (*t.FileDef, error)
	// FilesGetAll returns all file records for a given query.
	FilesGetAll(opts *t.QueryOpt, unusedOnly bool) ([]t.FileDef, error)
	// FileLink markes files as attachments in a specific message incrementing usage counter.
	FileLink(fid []string, topic string, seqid int) error
	// FileUnlink decrements usage counter of file records by query.
	FileUnlink(opts *t.QueryOpt) error
	// Delete file records by query. If unusedOnly is true, delete only records where useCount is zero.
	FileDelete(opts *t.QueryOpt, unusedOnly bool) error
}

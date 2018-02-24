// Package adapter contains the interfaces to be implemented by the database adapter
package adapter

import (
	"time"

	t "github.com/tinode/chat/server/store/types"
)

// Adapter is the interface that must be implemented by a database
// adapter. The current schema supports a single connection by database type.
type Adapter interface {
	Open(config string) error
	Close() error
	IsOpen() bool
	CheckDbVersion() error
	GetName() string

	CreateDb(reset bool) error

	// User management
	UserCreate(usr *t.User) error
	UserGet(id t.Uid) (*t.User, error)
	UserGetAll(ids ...t.Uid) ([]t.User, error)
	UserDelete(id t.Uid, soft bool) error
	UserUpdateLastSeen(uid t.Uid, userAgent string, when time.Time) error
	//UserUpdateStatus(uid t.Uid, status interface{}) error
	UserUpdate(uid t.Uid, update map[string]interface{}) error

	// Authentication management
	GetAuthRecord(unique string) (t.Uid, int, []byte, time.Time, error)
	AddAuthRecord(user t.Uid, authLvl int, unique string, secret []byte, expires time.Time) (bool, error)
	DelAuthRecord(unique string) (int, error)
	DelAllAuthRecords(uid t.Uid) (int, error)
	UpdAuthRecord(unique string, authLvl int, secret []byte, expires time.Time) (int, error)

	// Topic/contact management

	// TopicCreate creates a topic
	TopicCreate(topic *t.Topic) error
	// TopicCreateP2P creates a p2p topic
	TopicCreateP2P(initiator, invited *t.Subscription) error
	// TopicGet loads a single topic by name, if it exists. If the topic does not exist the call returns (nil, nil)
	TopicGet(topic string) (*t.Topic, error)
	// TopicsForUser loads subscriptions for a given user. Reads public value.
	TopicsForUser(uid t.Uid, keepDeleted bool) ([]t.Subscription, error)
	// UsersForTopic loads users' subscriptions for a given topic
	UsersForTopic(topic string, keepDeleted bool) ([]t.Subscription, error)
	TopicShare(subs []*t.Subscription) (int, error)
	TopicDelete(topic string) error
	// Increment Topic's or User's SeqId value
	TopicUpdateOnMessage(topic string, msg *t.Message) error
	TopicUpdate(topic string, update map[string]interface{}) error

	// SubscriptionGet rads a subscription of a user to a topic
	SubscriptionGet(topic string, user t.Uid) (*t.Subscription, error)
	// SubsForUser gets a list of topics of interest for a given user. Does NOT read public value.
	SubsForUser(user t.Uid, keepDeleted bool) ([]t.Subscription, error)
	// SubsForTopic gets a list of subscriptions to a given topic
	SubsForTopic(topic string, keepDeleted bool) ([]t.Subscription, error)
	// SubsUpdate updates pasrt of a subscription object. Pass nil for fields which don't need to be updated
	SubsUpdate(topic string, user t.Uid, update map[string]interface{}) error
	// SubsDelete deletes a single subscription
	SubsDelete(topic string, user t.Uid) error
	// SubsDelForTopic deletes all subscriptions to the given topic
	SubsDelForTopic(topic string) error

	// FindUsers searches for new contacts given a list of tags
	FindUsers(user t.Uid, tags []string) ([]t.Subscription, error)
	// FindTopics searches for group topics given a list of tags
	FindTopics(tags []string) ([]t.Subscription, error)
	UserTagsUpdate(user t.Uid, unique, tags t.StringSlice) error
	TopicTagsUpdate(topic string, unique, tags t.StringSlice) error

	// Messages
	MessageSave(msg *t.Message) error
	MessageGetAll(topic string, forUser t.Uid, opts *t.BrowseOpt) ([]t.Message, error)
	// Mark messages as deleted. Soft- or Hard- is defined by forUser value: forUSer.IsZero == true is hard.
	MessageDeleteList(topic string, toDel *t.DelMessage) error
	// Get a list of deleted message Ids
	MessageGetDeleted(topic string, forUser t.Uid, opts *t.BrowseOpt) ([]t.DelMessage, error)

	// Devices (for push notifications)
	DeviceUpsert(uid t.Uid, dev *t.DeviceDef) error
	DeviceGetAll(uid ...t.Uid) (map[t.Uid][]t.DeviceDef, int, error)
	DeviceDelete(uid t.Uid, deviceID string) error
}

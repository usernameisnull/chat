package types

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strings"
	"time"
)

// Uid is a database-specific record id, suitable to be used as a primary key.
type Uid uint64

var ZeroUid Uid = 0

const (
	uid_BASE64_UNPADDED = 11
	uid_BASE64_PADDED   = 12

	p2p_BASE64_UNPADDED = 22
	p2p_BASE64_PADDED   = 24
)

func (uid Uid) IsZero() bool {
	return uid == 0
}

func (uid Uid) Compare(u2 Uid) int {
	if uid < u2 {
		return -1
	} else if uid > u2 {
		return 1
	}
	return 0
}

func (uid *Uid) MarshalBinary() ([]byte, error) {
	dst := make([]byte, 8)
	binary.LittleEndian.PutUint64(dst, uint64(*uid))
	return dst, nil
}

func (uid *Uid) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return errors.New("Uid.UnmarshalBinary: invalid length")
	}
	*uid = Uid(binary.LittleEndian.Uint64(b))
	return nil
}

func (uid *Uid) UnmarshalText(src []byte) error {
	if len(src) != uid_BASE64_UNPADDED {
		return errors.New("Uid.UnmarshalText: invalid length")
	}
	dec := make([]byte, base64.URLEncoding.DecodedLen(uid_BASE64_PADDED))
	for len(src) < uid_BASE64_PADDED {
		src = append(src, '=')
	}
	count, err := base64.URLEncoding.Decode(dec, src)
	if count < 8 {
		if err != nil {
			return errors.New("Uid.UnmarshalText: failed to decode " + err.Error())
		}
		return errors.New("Uid.UnmarshalText: failed to decode")
	}
	*uid = Uid(binary.LittleEndian.Uint64(dec))
	return nil
}

func (uid *Uid) MarshalText() ([]byte, error) {
	if *uid == 0 {
		return []byte{}, nil
	}
	src := make([]byte, 8)
	dst := make([]byte, base64.URLEncoding.EncodedLen(8))
	binary.LittleEndian.PutUint64(src, uint64(*uid))
	base64.URLEncoding.Encode(dst, src)
	return dst[0:uid_BASE64_UNPADDED], nil
}

func (uid *Uid) MarshalJSON() ([]byte, error) {
	dst, _ := uid.MarshalText()
	return append(append([]byte{'"'}, dst...), '"'), nil
}

func (uid *Uid) UnmarshalJSON(b []byte) error {
	size := len(b)
	if size != (uid_BASE64_UNPADDED + 2) {
		return errors.New("Uid.UnmarshalJSON: invalid length")
	} else if b[0] != '"' || b[size-1] != '"' {
		return errors.New("Uid.UnmarshalJSON: unrecognized")
	}
	return uid.UnmarshalText(b[1 : size-1])
}

func (uid Uid) String() string {
	buf, _ := uid.MarshalText()
	return string(buf)
}

func ParseUid(s string) Uid {
	var uid Uid
	uid.UnmarshalText([]byte(s))
	return uid
}

func (uid Uid) UserId() string {
	return uid.PrefixId("usr")
}

func (uid Uid) FndName() string {
	return uid.PrefixId("fnd")
}

func (uid Uid) PrefixId(prefix string) string {
	if uid.IsZero() {
		return ""
	}
	return prefix + uid.String()
}

// ParseUserId parses user ID of the form "usrXXXXXX"
func ParseUserId(s string) Uid {
	var uid Uid
	if strings.HasPrefix(s, "usr") {
		(&uid).UnmarshalText([]byte(s)[3:])
	}
	return uid
}

// Given two UIDs generate a P2P topic name
func (uid Uid) P2PName(u2 Uid) string {
	if !uid.IsZero() && !u2.IsZero() {
		b1, _ := uid.MarshalBinary()
		b2, _ := u2.MarshalBinary()

		if uid < u2 {
			b1 = append(b1, b2...)
		} else if uid > u2 {
			b1 = append(b2, b1...)
		} else {
			// Explicitly disable P2P with self
			return ""
		}

		return "p2p" + base64.URLEncoding.EncodeToString(b1)[:p2p_BASE64_UNPADDED]
	}

	return ""
}

// ParseP2P extracts uids from the name of a p2p topic
func ParseP2P(p2p string) (uid1, uid2 Uid, err error) {
	if strings.HasPrefix(p2p, "p2p") {
		src := []byte(p2p)[3:]
		if len(src) != p2p_BASE64_UNPADDED {
			err = errors.New("ParseP2P: invalid length")
			return
		}
		dec := make([]byte, base64.URLEncoding.DecodedLen(p2p_BASE64_PADDED))
		for len(src) < p2p_BASE64_PADDED {
			src = append(src, '=')
		}
		var count int
		count, err = base64.URLEncoding.Decode(dec, src)
		if count < 16 {
			if err != nil {
				err = errors.New("ParseP2P: failed to decode " + err.Error())
			}
			err = errors.New("ParseP2P: invalid decoded length")
			return
		}
		uid1 = Uid(binary.LittleEndian.Uint64(dec))
		uid2 = Uid(binary.LittleEndian.Uint64(dec[8:]))
	} else {
		err = errors.New("ParseP2P: missing or invalid prefix")
	}
	return
}

// Header shared by all stored objects
type ObjHeader struct {
	Id        string // using string to get around rethinkdb's problems with unit64
	id        Uid
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (h *ObjHeader) Uid() Uid {
	if h.id.IsZero() && h.Id != "" {
		h.id.UnmarshalText([]byte(h.Id))
	}
	return h.id
}

func (h *ObjHeader) SetUid(uid Uid) {
	h.id = uid
	h.Id = uid.String()
}

func TimeNow() time.Time {
	return time.Now().UTC().Round(time.Millisecond)
}

// InitTimes initializes time.Time variables in the header to current time
func (h *ObjHeader) InitTimes() {
	if h.CreatedAt.IsZero() {
		h.CreatedAt = TimeNow()
	}
	h.UpdatedAt = h.CreatedAt
	h.DeletedAt = nil
}

// InitTimes initializes time.Time variables in the header to current time
func (h *ObjHeader) MergeTimes(h2 *ObjHeader) {
	// Set the creation time to the earliest value
	if h.CreatedAt.IsZero() || (!h2.CreatedAt.IsZero() && h2.CreatedAt.Before(h.CreatedAt)) {
		h.CreatedAt = h2.CreatedAt
	}
	// Set the update time to the latest value
	if h.UpdatedAt.Before(h2.UpdatedAt) {
		h.UpdatedAt = h2.UpdatedAt
	}
	// Set deleted time to the latest value
	if h2.DeletedAt != nil && (h.DeletedAt == nil || h.DeletedAt.Before(*h2.DeletedAt)) {
		h.DeletedAt = h2.DeletedAt
	}
}

// Stored user
type User struct {
	ObjHeader
	// Currently unused: Unconfirmed, Active, etc.
	State int

	Access DefaultAccess // Default access to user

	// Values for 'me' topic:
	// Server-issued sequence ID for messages in 'me'
	SeqId int
	// If messages were hard-deleted in the topic, id of the last deleted message
	ClearId int
	// Last time when the user joined 'me' topic, by User Agent
	LastSeen time.Time
	// User agent provided when accessing the topic last time
	UserAgent string

	Public interface{}

	// Unique indexed tags (email, phone) for finding this user. Stored on the
	// 'users' as well as indexed in 'tagunique'
	Tags []string

	Devices map[string]*DeviceDef
}

const max_devices = 8

type AccessMode uint

// User access to topic
const (
	ModeSub    AccessMode = 1 << iota // user can Read, i.e. {sub} (R)
	ModePub                           // user can Write, i.e. {pub} (W)
	ModePres                          // user can receive presence updates (P)
	ModeShare                         // user can invite other people to join (S)
	ModeDelete                        // user can hard-delete messages (D), only owner can completely delete
	ModeOwner                         // user is the owner (O) - full access
	ModeBanned                        // user has no access, requests to share/gain access/{sub} are ignored (X)

	ModeNone AccessMode = 0 // No access, requests to gain access are processed normally (N)
	// Read & write
	ModePubSub AccessMode = ModeSub | ModePub
	// normal user's access to a topic
	ModePublic AccessMode = ModeSub | ModePub | ModePres
	// self-subscription to !me - user can only read and delete incoming invites
	ModeSelf AccessMode = ModeSub | ModeDelete | ModePres
	// owner's subscription to a generic topic
	ModeFull AccessMode = ModeSub | ModePub | ModePres | ModeShare | ModeDelete | ModeOwner
	// manager of the topic - everything but being the owner
	ModeManager AccessMode = ModeSub | ModePub | ModePres | ModeShare | ModeDelete
	// Default P2P access mode
	ModeP2P AccessMode = ModeSub | ModePub | ModePres | ModeDelete

	// Invalid mode to indicate an error
	ModeInvalid AccessMode = 0x100000
)

func (m AccessMode) MarshalText() ([]byte, error) {

	// Need to distinguish between "not set" and "no access"
	if m == 0 {
		return []byte{'N'}, nil
	}

	if m == ModeInvalid {
		return nil, errors.New("AccessMode invalid")
	}

	// Banned mode superseeds all other modes
	if m&ModeBanned != 0 {
		return []byte{'X'}, nil
	}

	var res = []byte{}
	var modes = []byte{'R', 'W', 'P', 'S', 'D', 'O'}
	for i, chr := range modes {
		if (m & (1 << uint(i))) != 0 {
			res = append(res, chr)
		}
	}
	return res, nil
}

func (m *AccessMode) UnmarshalText(b []byte) error {
	var m0 AccessMode

	for i := 0; i < len(b); i++ {
		switch b[i] {
		case 'R', 'r':
			m0 |= ModeSub
		case 'W', 'w':
			m0 |= ModePub
		case 'S', 's':
			m0 |= ModeShare
		case 'D', 'd':
			m0 |= ModeDelete
		case 'P', 'p':
			m0 |= ModePres
		case 'O', 'o':
			m0 |= ModeOwner
		case 'X', 'x':
			m0 |= ModeBanned
		case 'N', 'n':
			m0 = 0 // N means explicitly no access, all other bits cleared
			break
		default:
			return errors.New("AccessMode: invalid character '" + string(b[i]) + "'")
		}
	}

	if m0&ModeBanned != 0 {
		m0 = ModeBanned // clear all other bits
	}

	*m = m0
	return nil
}

func (m AccessMode) String() string {
	res, err := m.MarshalText()
	if err != nil {
		return ""
	}
	return string(res)
}

func (m AccessMode) MarshalJSON() ([]byte, error) {
	res, err := m.MarshalText()
	if err != nil {
		return nil, err
	}

	res = append([]byte{'"'}, res...)
	return append(res, '"'), nil
}

func (m *AccessMode) UnmarshalJSON(b []byte) error {
	if b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("syntax error")
	}

	return m.UnmarshalText(b[1 : len(b)-1])
}

// Check if grant mode allows all that was requested in want mode
func (grant AccessMode) Check(want AccessMode) bool {
	return grant&want == want
}

// Check if banned
func (a AccessMode) IsBanned() bool {
	return a&ModeBanned != 0
}

// Check if owner
func (a AccessMode) IsOwner() bool {
	return a&ModeOwner != 0
}

// Check if owner or sharer
func (a AccessMode) IsManager() bool {
	return a.IsOwner() || (a&ModeShare != 0)
}

// Check if allowed to publish
func (a AccessMode) CanPub() bool {
	return a&ModePub != 0
}

// Relationship between users & topics, stored in database as Subscription
type TopicAccess struct {
	User  string
	Topic string
	Want  AccessMode
	Given AccessMode
}

// Subscription to a topic
type Subscription struct {
	ObjHeader
	User  string // User who has relationship with the topic
	Topic string // Topic subscribed to

	State int // Subscription state, currently unused

	// Values persisted through subscription deletion
	ClearId   int // User soft-deleted messages equal or lower to this seq ID
	RecvSeqId int // Last SeqId reported by user as received by at least one of his sessions
	ReadSeqId int // Last SeqID reported read by the user

	//
	ModeWant  AccessMode  // Access applied for
	ModeGiven AccessMode  // Granted access
	Private   interface{} // User's private data associated with the subscription to topic

	// Deserialized ephemeral values
	public      interface{} // Deserialized public value from topic or user (depends on context)
	with        string      // p2p topics only: id of the other user
	seqId       int         // deserialized SeqID from user or topic
	hardClearId int         // Id of the last hard-deleted message deserialized from user or topic
	lastSeen    time.Time   // timestamp when the user was last online
	userAgent   string      // user agent string of the last online access
}

// SetPublic assigns to public, otherwise not accessible from outside the package
func (s *Subscription) SetPublic(pub interface{}) {
	s.public = pub
}

func (s *Subscription) GetPublic() interface{} {
	return s.public
}

func (s *Subscription) SetWith(with string) {
	s.with = with
}

func (s *Subscription) GetWith() string {
	return s.with
}

func (s *Subscription) GetSeqId() int {
	return s.seqId
}

func (s *Subscription) SetSeqId(id int) {
	s.seqId = id
}

func (s *Subscription) GetHardClearId() int {
	return s.hardClearId
}

func (s *Subscription) SetHardClearId(id int) {
	s.hardClearId = id
}

func (s *Subscription) GetLastSeen() time.Time {
	return s.lastSeen
}

func (s *Subscription) GetUserAgent() string {
	return s.userAgent
}

func (s *Subscription) SetLastSeenAndUA(when time.Time, ua string) {
	s.lastSeen = when
	s.userAgent = ua
}

// Result of a search for connections
type Contact struct {
	Id       string
	MatchOn  []string
	Access   DefaultAccess
	LastSeen time.Time
	Public   interface{}
}

type perUserData struct {
	//owner   bool
	private interface{}
	want    AccessMode
	given   AccessMode
}

// Topic stored in database
type Topic struct {
	ObjHeader
	State int

	// Name  string -- topic name is stored in Id

	// Use bearer token or use ACL
	UseBt bool

	// Default access to topic
	Access DefaultAccess

	// Server-issued sequential ID
	SeqId int
	// If messages were deleted, id of the last deleted message
	ClearId int

	Public interface{}

	// Deserialized ephemeral params
	owner   Uid                  // first assigned owner
	perUser map[Uid]*perUserData // deserialized from Subscription
}

type DefaultAccess struct {
	Auth AccessMode
	Anon AccessMode
}

//func (t *Topic) GetAccessList() []TopicAccess {
//	return t.users
//}

func (t *Topic) GiveAccess(uid Uid, want AccessMode, given AccessMode) {
	if t.perUser == nil {
		t.perUser = make(map[Uid]*perUserData, 1)
	}

	pud := t.perUser[uid]
	if pud == nil {
		pud = &perUserData{}
	}

	pud.want = want
	pud.given = given

	t.perUser[uid] = pud
	if want&given&ModeOwner != 0 && t.owner.IsZero() {
		t.owner = uid
	}
}

func (t *Topic) SetPrivate(uid Uid, private interface{}) {
	if t.perUser == nil {
		t.perUser = make(map[Uid]*perUserData, 1)
	}
	pud := t.perUser[uid]
	if pud == nil {
		pud = &perUserData{}
	}
	pud.private = private
	t.perUser[uid] = pud
}

func (t *Topic) GetOwner() Uid {
	return t.owner
}

func (t *Topic) GetPrivate(uid Uid) (private interface{}) {
	if t.perUser == nil {
		return
	}
	pud := t.perUser[uid]
	if pud == nil {
		return
	}
	private = pud.private
	return
}

func (t *Topic) GetAccess(uid Uid) (mode AccessMode) {
	if t.perUser == nil {
		return
	}
	pud := t.perUser[uid]
	if pud == nil {
		return
	}
	mode = pud.given & pud.want
	return
}

// Stored {data} message
type Message struct {
	ObjHeader
	SeqId   int
	Topic   string
	From    string // UID as string of the user who sent the message, could be empty
	Head    map[string]string
	Content interface{}
}

// Invites

type InviteAction int

const (
	InvJoin InviteAction = iota // an invitation to subscribe
	InvAppr                     // a request to aprove a subscription
	InvInfo                     // info only (request approved or subscribed by a third party), no action required
)

func (a InviteAction) String() string {
	switch a {
	case InvJoin:
		return "join"
	case InvAppr:
		return "appr"
	case InvInfo:
		return "info"
	}
	return ""
}

type BrowseOpt struct {
	Since  int
	Before int
	Limit  uint
}

type TopicCat int

const (
	TopicCat_Me TopicCat = iota
	TopicCat_Fnd
	TopicCat_P2P
	TopicCat_Grp
)

func GetTopicCat(name string) TopicCat {
	switch name[:3] {
	case "usr":
		return TopicCat_Me
	case "p2p":
		return TopicCat_P2P
	case "grp":
		return TopicCat_Grp
	case "fnd":
		return TopicCat_Fnd
	default:
		panic("invalid topic type" + name)
	}
}

// Data provided by connected device. Used primarily for
// push notifications
type DeviceDef struct {
	// Device registration ID
	DeviceId string
	// Device platform (iOS, Android, Web)
	Platform string
	// Last logged in
	LastSeen time.Time
	// Device language, ISO code
	Lang string
}

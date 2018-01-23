package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/tinode/chat/server/auth"
	"github.com/tinode/chat/server/store"
	"github.com/tinode/chat/server/store/types"
)

// TokenAuth is a singleton instance of the authenticator.
type TokenAuth struct {
	tokenHmacSalt     []byte
	tokenTimeout      time.Duration
	tokenSerialNumber int
}

// Token composition: [8:UID][4:expires][2:authLevel][2:serial-number][32:signature] == 48 bytes
// Token markers
const (
	tokenUIDStart = 0
	tokenUIDEnd   = 8

	tokenExpiresStart = 8
	tokenExpiresEnd   = 12

	tokenAuthLvlStart = 12
	tokenAuthLvlEnd   = 14

	tokenSerialStart = 14
	tokenSerialEnd   = 16

	tokenSignatureStart = 16

	tokenLengthDecoded = 48

	tokenMinHmacLength = 32
)

// Init initializes the authenticator: parses the config and sets salt, serial number and lifetime.
func (ta *TokenAuth) Init(jsonconf string) error {
	if ta.tokenHmacSalt != nil {
		return errors.New("auth_token: already initialized")
	}

	type configType struct {
		// Key for signing tokens
		Key []byte `json:"key"`
		// Datatabase or other serial number, to invalidate all issued tokens at once.
		SerialNum int `json:"serial_num"`
		// Token expiration time
		ExpireIn int `json:"expire_in"`
	}
	var config configType
	if err := json.Unmarshal([]byte(jsonconf), &config); err != nil {
		return errors.New("auth_token: failed to parse config: " + err.Error() + "(" + jsonconf + ")")
	}

	if config.Key == nil || len(config.Key) < tokenMinHmacLength {
		return errors.New("auth_token: the key is missing or too short")
	}
	if config.ExpireIn <= 0 {
		return errors.New("auth_token: invalid expiration value")
	}

	ta.tokenHmacSalt = config.Key
	ta.tokenTimeout = time.Duration(config.ExpireIn) * time.Second

	ta.tokenSerialNumber = config.SerialNum

	return nil
}

// AddRecord is not supprted, will produce an error.
func (TokenAuth) AddRecord(uid types.Uid, secret []byte, lifetime time.Duration) (int, auth.AuthErr) {
	return auth.LevelNone, auth.NewErr(auth.ErrUnsupported, errors.New("token auth: AddRecord is not supported"))
}

// UpdateRecord is not supported, will produce an error.
func (TokenAuth) UpdateRecord(uid types.Uid, secret []byte, lifetime time.Duration) auth.AuthErr {
	return auth.NewErr(auth.ErrUnsupported, errors.New("token auth: UpdateRecord is not supported"))
}

// Authenticate checks validity of provided token.
func (ta *TokenAuth) Authenticate(token []byte) (types.Uid, int, time.Time, auth.AuthErr) {
	// [8:UID][4:expires][2:authLevel][2:serial-number][32:signature] == 48 bytes

	if len(token) < tokenLengthDecoded {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, errors.New("token auth: invalid length"))
	}

	var uid types.Uid
	if err := uid.UnmarshalBinary(token[tokenUIDStart:tokenUIDEnd]); err != nil {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, err)
	}
	var authLvl int
	if authLvl = int(binary.LittleEndian.Uint16(token[tokenAuthLvlStart:tokenAuthLvlEnd])); authLvl < 0 || authLvl > auth.LevelRoot {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, errors.New("token auth: invalid auth level"))
	}

	if snum := int(binary.LittleEndian.Uint16(token[tokenSerialStart:tokenSerialEnd])); snum != ta.tokenSerialNumber {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, errors.New("token auth: serial number does not match"))
	}

	hasher := hmac.New(sha256.New, ta.tokenHmacSalt)
	hasher.Write(token[:tokenSignatureStart])
	if !hmac.Equal(token[tokenSignatureStart:], hasher.Sum(nil)) {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrFailed, errors.New("token auth: invalid signature"))
	}

	expires := time.Unix(int64(binary.LittleEndian.Uint32(token[tokenExpiresStart:tokenExpiresEnd])), 0).UTC()
	if expires.Before(time.Now().Add(1 * time.Second)) {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrExpired, errors.New("token auth: expired token"))
	}

	return uid, authLvl, expires, auth.NewErr(auth.NoErr, nil)
}

// GenSecret generates a new token.
func (ta *TokenAuth) GenSecret(uid types.Uid, authLvl int, lifetime time.Duration) ([]byte, time.Time, auth.AuthErr) {
	// [8:UID][4:expires][2:authLevel][2:serial-number][32:signature] == 48 bytes

	buf := new(bytes.Buffer)
	uidbits, _ := uid.MarshalBinary()
	binary.Write(buf, binary.LittleEndian, uidbits)
	if lifetime == 0 {
		lifetime = ta.tokenTimeout
	} else if lifetime < 0 {
		return nil, time.Time{}, auth.NewErr(auth.ErrExpired, errors.New("token auth: negative lifetime"))
	}
	expires := time.Now().Add(lifetime).UTC().Round(time.Millisecond)
	binary.Write(buf, binary.LittleEndian, uint32(expires.Unix()))
	binary.Write(buf, binary.LittleEndian, uint16(authLvl))
	binary.Write(buf, binary.LittleEndian, uint16(ta.tokenSerialNumber))

	hasher := hmac.New(sha256.New, ta.tokenHmacSalt)
	hasher.Write(buf.Bytes())
	binary.Write(buf, binary.LittleEndian, hasher.Sum(nil))

	return buf.Bytes(), expires, auth.NewErr(auth.NoErr, nil)
}

// IsUnique is not supported, will produce an error.
func (TokenAuth) IsUnique(token []byte) (bool, auth.AuthErr) {
	return false, auth.NewErr(auth.ErrUnsupported, errors.New("auth token: IsUnique is not supported"))
}

func init() {
	store.RegisterAuthScheme("token", &TokenAuth{})
}

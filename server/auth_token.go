package main

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

// TokenAuth is a placeholder singleton instance of the authenticator.
type TokenAuth struct{}

// Token composition: [8:UID][4:expires][2:authLevel][2:serial-number][32:signature] == 48 bytes
// Token markers
const (
	UID_START = 0
	UID_END   = 8

	EXPIRES_START = 8
	EXPIRES_END   = 12

	AUTHLVL_START = 12
	AUTHLVL_END   = 14

	SERIAL_START = 14
	SERIAL_END   = 16

	SIGN_START = 16
)

const (
	tokenLengthDecoded = 48
	tokenMinHmacLength = 32
)

var tokenHmacSalt []byte
var tokenTimeout time.Duration
var tokenSerialNumber int

// Init initializes the authenticator: parses the config and sets salt, serial number and lifetime.
func (TokenAuth) Init(jsonconf string) error {
	if tokenHmacSalt != nil {
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

	tokenHmacSalt = config.Key
	tokenTimeout = time.Duration(config.ExpireIn) * time.Second

	tokenSerialNumber = config.SerialNum

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
func (TokenAuth) Authenticate(token []byte) (types.Uid, int, time.Time, auth.AuthErr) {
	// [8:UID][4:expires][2:authLevel][2:serial-number][32:signature] == 48 bytes

	if len(token) < tokenLengthDecoded {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, errors.New("token auth: invalid length"))
	}

	var uid types.Uid
	if err := uid.UnmarshalBinary(token[UID_START:UID_END]); err != nil {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, err)
	}
	var authLvl int
	if authLvl = int(binary.LittleEndian.Uint16(token[AUTHLVL_START:AUTHLVL_END])); authLvl < 0 || authLvl > auth.LevelRoot {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, errors.New("token auth: invalid auth level"))
	}

	if snum := int(binary.LittleEndian.Uint16(token[SERIAL_START:SERIAL_END])); snum != tokenSerialNumber {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrMalformed, errors.New("token auth: serial number does not match"))
	}

	hasher := hmac.New(sha256.New, tokenHmacSalt)
	hasher.Write(token[:SIGN_START])
	if !hmac.Equal(token[SIGN_START:], hasher.Sum(nil)) {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrFailed, errors.New("token auth: invalid signature"))
	}

	expires := time.Unix(int64(binary.LittleEndian.Uint32(token[EXPIRES_START:EXPIRES_END])), 0).UTC()
	if expires.Before(time.Now().Add(1 * time.Second)) {
		return types.ZeroUid, auth.LevelNone, time.Time{},
			auth.NewErr(auth.ErrExpired, errors.New("token auth: expired token"))
	}

	return uid, authLvl, expires, auth.NewErr(auth.NoErr, nil)
}

// GenSecret generates a new token.
func (TokenAuth) GenSecret(uid types.Uid, authLvl int, lifetime time.Duration) ([]byte, time.Time, auth.AuthErr) {
	// [8:UID][4:expires][2:authLevel][2:serial-number][32:signature] == 48 bytes

	buf := new(bytes.Buffer)
	uidbits, _ := uid.MarshalBinary()
	binary.Write(buf, binary.LittleEndian, uidbits)
	if lifetime == 0 {
		lifetime = tokenTimeout
	} else if lifetime < 0 {
		return nil, time.Time{}, auth.NewErr(auth.ErrExpired, errors.New("token auth: negative lifetime"))
	}
	expires := time.Now().Add(lifetime).UTC().Round(time.Millisecond)
	binary.Write(buf, binary.LittleEndian, uint32(expires.Unix()))
	binary.Write(buf, binary.LittleEndian, uint16(authLvl))
	binary.Write(buf, binary.LittleEndian, uint16(tokenSerialNumber))

	hasher := hmac.New(sha256.New, tokenHmacSalt)
	hasher.Write(buf.Bytes())
	binary.Write(buf, binary.LittleEndian, hasher.Sum(nil))

	return buf.Bytes(), expires, auth.NewErr(auth.NoErr, nil)
}

// IsUnique is not supported, will produce an error.
func (TokenAuth) IsUnique(token []byte) (bool, auth.AuthErr) {
	return false, auth.NewErr(auth.ErrUnsupported, errors.New("auth token: IsUnique is not supported"))
}

func init() {
	var auth TokenAuth
	store.RegisterAuthScheme("token", auth)
}

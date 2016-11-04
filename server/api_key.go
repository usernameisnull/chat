/******************************************************************************
 *
 *  Description :
 *
 *  Authentication
 *
 *****************************************************************************/

package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"log"
)

// 32 random bytes to be used for signing auth tokens
// FIXME(gene): move it to the database (make it unique per-application)
// base64 = T713/rYYgW7g4m3vG6zGRh7+FM1t0T8j13koXScOAj4=
/*
var hmac_salt = []byte{
	0x4f, 0xbd, 0x77, 0xfe, 0xb6, 0x18, 0x81, 0x6e,
	0xe0, 0xe2, 0x6d, 0xef, 0x1b, 0xac, 0xc6, 0x46,
	0x1e, 0xfe, 0x14, 0xcd, 0x6d, 0xd1, 0x3f, 0x23,
	0xd7, 0x79, 0x28, 0x5d, 0x27, 0x0e, 0x02, 0x3e}
*/

// Singned AppID. Composition:
//   [1:algorithm version][4:appid][2:key sequence][1:isRoot][16:signature] = 24 bytes
// convertible to base64 without padding
// All integers are little-endian

const (
	APIKEY_VERSION = 1
	// APIKEY_APPID is deprecated and will be removed in the future
	APIKEY_APPID     = 4
	APIKEY_SEQUENCE  = 2
	APIKEY_WHO       = 1
	APIKEY_SIGNATURE = 16
	APIKEY_LENGTH    = APIKEY_VERSION + APIKEY_APPID + APIKEY_SEQUENCE + APIKEY_WHO + APIKEY_SIGNATURE
)

// Client signature validation
//   key: client's secret key
// Returns application id, key type
func checkApiKey(apikey string) (isValid, isRoot bool) {

	if declen := base64.URLEncoding.DecodedLen(len(apikey)); declen != APIKEY_LENGTH {
		return
	}

	data, err := base64.URLEncoding.DecodeString(apikey)
	if err != nil {
		log.Println("failed to decode.base64 appid ", err)
		return
	}
	if data[0] != 1 {
		log.Println("unknown appid signature algorithm ", data[0])
		return
	}

	hasher := hmac.New(md5.New, globals.apiKeySalt)
	hasher.Write(data[:APIKEY_VERSION+APIKEY_APPID+APIKEY_SEQUENCE+APIKEY_WHO])
	check := hasher.Sum(nil)
	if !bytes.Equal(data[APIKEY_VERSION+APIKEY_APPID+APIKEY_SEQUENCE+APIKEY_WHO:], check) {
		log.Println("invalid apikey signature")
		return
	}

	isRoot = (data[APIKEY_VERSION+APIKEY_APPID+APIKEY_SEQUENCE] == 1)

	isValid = true

	return
}

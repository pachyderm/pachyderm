package random

import (
	"crypto/rand"
	"encoding/base64"
)

// String returns a cryptographically random, URL safe string with length
// at least n
//
// TODO(msteffen): move away from UUIDv4 towards this (current implementation of
// UUIDv4 produces UUIDs via CSPRNG, but the UUIDv4 spec doesn't guarantee that
// behavior, and we shouldn't assume it going forward)
func String(n int) string {
	var numBytes int
	for n >= base64.RawURLEncoding.EncodedLen(numBytes) {
		numBytes++
	}
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		panic("could not generate cryptographically secure random string!")
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

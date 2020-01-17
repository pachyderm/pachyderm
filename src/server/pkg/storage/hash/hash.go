package hash

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

// New creates a new hasher.
// (bryce) make a decision on which hash to go with.
func New() hash.Hash {
	return sha512.New()
}

// Sum computes a hash sum for a set of bytes.
func Sum(data []byte) []byte {
	sum := sha512.Sum512(data)
	return sum[:]
}

// EncodeHash encodes a hash into a string representation.
func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

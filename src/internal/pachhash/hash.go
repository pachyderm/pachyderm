package pachhash

import (
	"encoding/hex"
	"hash"

	"golang.org/x/crypto/blake2b"
)

// The size of an Output in bytes
const OutputSize = 32

type Output = [32]byte

// New creates a new hasher.
func New() hash.Hash {
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	return h
}

// Sum computes a hash sum for a set of bytes.
func Sum(data []byte) Output {
	return blake2b.Sum256(data)
}

// EncodeHash encodes a hash into a string representation.
func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

package pachhash

import (
	"encoding/hex"
	"hash"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/crypto/blake2b"
)

// OutputSize is the size of an Output in bytes
const OutputSize = 32

// Output is the output of the hash function.
// Sum returns an Output
type Output = [OutputSize]byte

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

// ParseHex parses a hex string into output.
func ParseHex(x []byte) (*Output, error) {
	o := Output{}
	n, err := hex.Decode(o[:], x)
	if err != nil {
		return nil, err
	}
	if n < OutputSize {
		return nil, errors.Errorf("hex string too short to be Output")
	}
	return &o, nil
}

// EncodeHash encodes a hash into a string representation.
func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

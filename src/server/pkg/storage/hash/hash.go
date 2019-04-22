package hash

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

func New() hash.Hash {
	return sha512.New()
}

func Sum(data []byte) []byte {
	sum := sha512.Sum512(data)
	return sum[:]
}

func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

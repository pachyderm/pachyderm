package server

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func NewHash() hash.Hash {
	return sha512.New()
}

func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func getBlock(hash hash.Hash) *pfs.Block {
	return &pfs.Block{
		Hash: base64.URLEncoding.EncodeToString(hash.Sum(nil)),
	}
}

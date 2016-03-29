package server

import (
	"crypto/sha512"
	"encoding/base64"
	"hash"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func newHash() hash.Hash {
	return sha512.New()
}

func getBlock(hash hash.Hash) *pfs.Block {
	return &pfs.Block{
		Hash: base64.URLEncoding.EncodeToString(hash.Sum(nil)),
	}
}

package server

import (
	"crypto/sha512"
	"encoding/base64"
	"hash"

	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
)

func newHash() hash.Hash {
	return sha512.New()
}

func getBlock(hash hash.Hash) *pfsserver.Block {
	return &pfsserver.Block{
		Hash: base64.URLEncoding.EncodeToString(hash.Sum(nil)),
	}
}

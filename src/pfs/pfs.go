package pfs

import (
	"encoding/hex"
	"hash"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
)

const (
	// ChunkSize is the size of file chunks when resumable upload is used
	ChunkSize = int64(512 * 1024 * 1024) // 512 MB

	// default system repo types
	UserRepoType = "user"
	MetaRepoType = "meta"
	SpecRepoType = "spec"
)

// NewHash returns a hash that PFS uses internally to compute checksums.
func NewHash() hash.Hash {
	return pachhash.New()
}

// EncodeHash encodes a hash into a readable format.
func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

// DecodeHash decodes a hash into bytes.
func DecodeHash(hash string) ([]byte, error) {
	return hex.DecodeString(hash)
}

func (r *Repo) String() string {
	if r.Type == UserRepoType {
		return r.Name
	}
	return r.Name + "." + r.Type
}

func (r *Repo) NewBranch(name string) *Branch {
	return &Branch{
		Repo: proto.Clone(r).(*Repo),
		Name: name,
	}
}

func (r *Repo) NewCommit(branch, id string) *Commit {
	return &Commit{
		ID:     id,
		Branch: r.NewBranch(branch),
	}
}

func (c *Commit) NewFile(path string) *File {
	return &File{
		Commit: proto.Clone(c).(*Commit),
		Path:   path,
	}
}

func (c *Commit) String() string {
	return c.Branch.String() + "=" + c.ID
}

func (b *Branch) NewCommit(id string) *Commit {
	return &Commit{
		Branch: proto.Clone(b).(*Branch),
		ID:     id,
	}
}

func (b *Branch) String() string {
	return b.Repo.String() + "@" + b.Name
}

package pfs

import (
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
)

const (
	// ChunkSize is the size of file chunks when resumable upload is used
	ChunkSize = int64(512 * 1024 * 1024) // 512 MB
	// EmptyStr is included in the description of output commits from failed jobs.
	EmptyStr = "(empty)"

	// default system repo types
	UserRepoType  = "user"
	MetaRepoType  = "meta"
	BuildRepoType = "build"
	SpecRepoType  = "spec"
)

// FullID prints repoName@CommitID
func (c *Commit) FullID() string {
	return fmt.Sprintf("%s@%s", c.Branch.Repo.Name, c.ID)
}

func (c *Commit) Job() *Job {
	return &Job{ID: c.ID}
}

func (ji *JobInfo) Commit(jobCommitInfo *JobCommitInfo) *Commit {
	return &Commit{
		Branch: proto.Clone(jobCommitInfo.Branch).(*Branch),
		ID:     ji.Job.ID,
	}
}

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

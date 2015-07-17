package drive

import (
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/satori/go.uuid"
)

const (
	InitialCommitID = "scratch"
)

var (
	ReservedCommitIDs = map[string]bool{
		InitialCommitID: true,
	}
)

type Driver interface {
	Init() error
	InitRepository(repository *pfs.Repository, shard map[int]bool) error
	GetFile(path *pfs.Path, shard int) (io.ReadCloser, error)
	MakeDirectory(path *pfs.Path, shards map[int]bool) error
	PutFile(path *pfs.Path, shard int, reader io.Reader) error
	ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error)
	GetParent(commit *pfs.Commit, shard int) (*pfs.Commit, error)
	Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error)
	Commit(commit *pfs.Commit, shards map[int]bool) error
	PullDiff(commit *pfs.Commit, shard int) (io.Reader, error)
	PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error
	GetCommitInfo(commit *pfs.Commit, shard int) (*pfs.CommitInfo, error)
}

func NewCommitID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

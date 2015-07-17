package drive

import (
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/satori/go.uuid"
)

const (
	SystemRootCommitID = "__root__"
	InitialCommitID    = "scratch"
)

var (
	ReservedCommitIDs = map[string]bool{
		SystemRootCommitID: true,
	}
)

type Driver interface {
	Init() error
	InitRepository(repository *pfs.Repository, shard map[int]bool) error
	GetFile(path *pfs.Path, shard int) (io.ReadCloser, error)
	MakeDirectory(path *pfs.Path, shard int) error
	PutFile(path *pfs.Path, shard int, reader io.Reader) error
	ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error)
	GetParent(commit *pfs.Commit) (*pfs.Commit, error)
	Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error)
	Commit(commit *pfs.Commit, shards map[int]bool) error
	PullDiff(commit *pfs.Commit, shard int) (io.Reader, error)
	PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error
	GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error)
}

func NewCommitID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

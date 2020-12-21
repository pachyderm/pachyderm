package server

import (
	"context"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

type commitStore interface {
	AddFileset(ctx context.Context, commit *pfs.Commit, filesetID string) error
	GetFileset(ctx context.Context, commit *pfs.Commit) (filesetID string, err error)
}

// keySpaceCommitStore is a temporary implementation using the existing keyspace for filesets.
type keySpaceCommitStore struct {
	storage *fileset.Storage
}

func newKeySpaceCommitStore(s *fileset.Storage) *keySpaceCommitStore {
	return &keySpaceCommitStore{
		storage: s,
	}
}

func (s *keySpaceCommitStore) AddFileset(ctx context.Context, commit *pfs.Commit, filesetID string) error {
	n := time.Now().UnixNano()
	subFilesetStr := fileset.SubFileSetStr(n)
	subFilesetPath := path.Join(commit.Repo.Name, commit.ID, subFilesetStr)
	return s.storage.Copy(ctx, path.Join(tmpRepo, filesetID), subFilesetPath, 0)
}

func (s *keySpaceCommitStore) GetFileset(ctx context.Context, commit *pfs.Commit) (string, error) {
	filesetID := uuid.NewWithoutDashes()
	if err := s.storage.Copy(ctx, commitPath(commit), path.Join(tmpRepo, filesetID), defaultTTL); err != nil {
		return "", err
	}
	return filesetID, nil
}

// TODO
// type postgresCommitStore struct {
// }

package server

import (
	"context"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/track"
)

type commitStore interface {
	AddFileset(ctx context.Context, commit *pfs.Commit, filesetID fileset.ID) error
	GetFileset(ctx context.Context, commit *pfs.Commit) (filesetID *fileset.ID, err error)
	UpdateFileset(ctx context.Context, commit *pfs.Commit, fn func(x fileset.ID) (*fileset.ID, error)) error
	DropFilesets(ctx context.Context, commit *pfs.Commit) error
}

type memCommitStore struct {
	s *fileset.Storage

	mu       sync.Mutex
	staging  map[string][]fileset.ID
	finished map[string]fileset.ID
}

func newMemCommitStore(s *fileset.Storage) *memCommitStore {
	return &memCommitStore{
		s:        s,
		staging:  make(map[string][]fileset.ID),
		finished: make(map[string]fileset.ID),
	}
}

func (s *memCommitStore) AddFileset(ctx context.Context, commit *pfs.Commit, filesetID fileset.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := commitKey(commit)
	if _, exists := s.finished[key]; exists {
		return errors.Errorf("commit is finished")
	}
	id, err := s.s.Clone(ctx, filesetID, track.NoTTL)
	if err != nil {
		return err
	}
	ids := s.staging[key]
	ids = append(ids, *id)
	s.staging[key] = ids
	return nil
}

func (s *memCommitStore) GetFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := commitKey(commit)
	if id, exists := s.finished[key]; exists {
		return s.s.Clone(ctx, id, defaultTTL)
	}
	// return nil, errors.Errorf("commit is not finished")
	return s.s.Compose(ctx, s.staging[key], defaultTTL)
}

func (s *memCommitStore) UpdateFileset(ctx context.Context, commit *pfs.Commit, fn func(fileset.ID) (*fileset.ID, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := commitKey(commit)
	x, exists := s.finished[key]
	if !exists {
		id, err := s.s.Compose(ctx, s.staging[key], defaultTTL)
		if err != nil {
			return err
		}
		x = *id
	}
	y, err := fn(x)
	if err != nil {
		return err
	}
	id, err := s.s.Clone(ctx, *y, track.NoTTL)
	if err != nil {
		return err
	}
	s.finished[key] = *id
	return s.s.Drop(ctx, x)
}

func (s *memCommitStore) DropFilesets(ctx context.Context, commit *pfs.Commit) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := commitKey(commit)
	delete(s.finished, key)
	delete(s.staging, key)
	return nil
}

// TODO
// type postgresCommitStore struct {
// }

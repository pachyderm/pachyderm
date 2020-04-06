package gc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/prometheus/client_golang/prometheus"
)

// Server is the interface that the garbage collector service provides to clients
type Server interface {
	DeleteChunk(context.Context, string) error
	FlushDelete(context.Context, string) error
}

// Deleter is an interface that must be provided when creating the garbage
// collector server, it handles removing chunks from the backing object storage.
type Deleter interface {
	Delete(context.Context, []string) error
}

// wraps a sync.Cond as a one-time event trigger
type trigger struct {
	cond      *sync.Cond
	triggered bool
}

func newTrigger() *trigger {
	return &trigger{cond: sync.NewCond(&sync.Mutex{})}
}

func (t *trigger) withLock(cb func()) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()
	cb()
}

func (t *trigger) Trigger() {
	t.withLock(func() {
		t.triggered = true
		t.cond.Broadcast()
	})
}

func (t *trigger) Wait() {
	t.withLock(func() {
		for !t.triggered {
			t.cond.Wait()
		}
	})
}

type server struct {
	deleter  Deleter
	db       *gorm.DB
	mutex    sync.Mutex
	deleting map[string]*trigger
}

// NewServer constructs a garbage collector server:
//  * deleter - an object implementing the Deleter interface that will be
//      called for deleting chunks from object storage
//  * postgresHost, postgresPort - the host and port of the postgres instance
//      which is used for coordinating garbage collection reference counts with
//      the garbage collector clients
//  * registry (optional) - a Prometheus stats registry for tracking usage and
//    performance
func NewServer(deleter Deleter, postgresHost string, postgresPort uint16, registry prometheus.Registerer) (Server, error) {
	if registry != nil {
		initPrometheus(registry)
	}

	db, err := openDatabase(postgresHost, postgresPort)
	if err != nil {
		return nil, err
	}

	// TODO: determine reasonable values for this
	db.LogMode(false)
	db.DB().SetMaxOpenConns(3)
	db.DB().SetMaxIdleConns(2)
	return &server{
		deleter:  deleter,
		db:       db,
		deleting: make(map[string]*trigger),
	}, nil
}

func (s *server) markChunksDeleting(ctx context.Context, chunks []string) ([]string, error) {
	chunksDeleting := []chunkModel{}
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Set the deleting field for the passed in chunks, excluding
			// the chunks that had a reference added before the deleting process
			// began.
			return txn.Raw(`
				UPDATE chunks
				SET deleting = now()
				WHERE chunk IN (?)
				AND chunk NOT IN (
					SELECT DISTINCT chunk
					FROM refs
					WHERE chunk IN (?)
				)
				RETURNING chunk
			`, chunks, chunks).Scan(&chunksDeleting)
		},
	}
	if err := runTransaction(ctx, s.db, stmtFuncs, markChunksDeletingStats); err != nil {
		return nil, err
	}
	return convertChunks(chunksDeleting), nil
}

func (s *server) removeChunkRows(ctx context.Context, chunks []string) ([]string, error) {
	deletedChunks := []chunkModel{}
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Delete the chunks and references from the deleted chunks.
			// Return the chunks that should be transitively deleted (if any).
			return txn.Raw(`
				WITH deleted_chunks AS (
					DELETE FROM chunks
					WHERE chunk IN (?)
					RETURNING chunk
				), deleted_refs AS (
					DELETE FROM refs
					USING deleted_chunks
					WHERE refs.sourcetype = 'chunk'
					AND refs.source = deleted_chunks.chunk
					RETURNING refs.chunk
				)
				SELECT deleted_refs.chunk
				FROM deleted_refs
				JOIN refs
				ON deleted_refs.chunk = refs.chunk
				GROUP BY 1
				HAVING COUNT(*) = 1
		`, chunks).Scan(&deletedChunks)
		},
	}
	if err := runTransaction(ctx, s.db, stmtFuncs, removeChunkRowsStats); err != nil {
		return nil, err
	}
	return convertChunks(deletedChunks), nil
}

func (s *server) DeleteChunk(ctx context.Context, chunk string) (retErr error) {
	defer func(start time.Time) { applyRequestStats("DeleteChunk", retErr, start) }(time.Now())

	trigger := newTrigger()

	// Check if we have outstanding deletes for these chunks and save the trigger
	candidates := func() []string {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		result := []string{}
		if _, ok := s.deleting[chunk]; !ok {
			s.deleting[chunk] = trigger
			result = append(result, chunk)
		}
		return result
	}()

	if len(candidates) == 0 {
		return
	}

	// Spawn goroutine to do all the remote calls async
	// (bryce) need to think more about the retry strategy.
	go func() {
		// set the chunks as deleting
		var toDelete []string
		backoff.RetryNotify(func() error {
			var err error
			toDelete, err = s.markChunksDeleting(context.Background(), candidates)
			return err
		}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
			fmt.Println("markChunksDeleting: ", err)
			return nil
		})

		if len(toDelete) > 0 {
			// delete objects from object storage
			backoff.RetryNotify(func() error {
				defer func(start time.Time) { applyDeleteStats(retErr, start) }(time.Now())
				return s.deleter.Delete(context.Background(), toDelete)
			}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
				fmt.Println("delete: ", err)
				return nil
			})

			// delete the rows from the db
			transitiveDeletes := []string{}
			backoff.RetryNotify(func() error {
				var err error
				transitiveDeletes, err = s.removeChunkRows(context.Background(), toDelete)
				return err
			}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
				fmt.Println("removeChunkRows: ", err)
				return nil
			})

			// Pass transitive deletes to a new RPC
			for _, chunk := range transitiveDeletes {
				s.DeleteChunk(context.Background(), chunk)
			}
		}

		func() {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			for _, c := range candidates {
				delete(s.deleting, c)
			}
		}()

		trigger.Trigger()
	}()

	return nil
}

func (s *server) FlushDelete(ctx context.Context, chunk string) (retErr error) {
	defer func(start time.Time) { applyRequestStats("FlushDelete", retErr, start) }(time.Now())
	s.mutex.Lock()
	trigger := s.deleting[chunk]
	s.mutex.Unlock()
	if trigger != nil {
		trigger.Wait()
	}
	return nil
}

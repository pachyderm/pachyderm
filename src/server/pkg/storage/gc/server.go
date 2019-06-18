package gc

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/prometheus/client_golang/prometheus"
)

type Server interface {
	DeleteChunks(context.Context, []chunk.Chunk) error
	FlushDeletes(context.Context, []chunk.Chunk) error
}

type Deleter interface {
	Delete(context.Context, []chunk.Chunk) error
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

type serverImpl struct {
	deleter  Deleter
	db       *sql.DB
	mutex    sync.Mutex
	deleting map[string]*trigger
}

func MakeServer(deleter Deleter, host string, port uint16, registry prometheus.Registerer) (Server, error) {
	connStr := fmt.Sprintf("host=%s port=%d dbname=pgc user=pachyderm password=elephantastic sslmode=disable", host, port)
	connector, err := pq.NewConnector(connStr)
	if err != nil {
		return nil, err
	}

	if registry != nil {
		initPrometheus(registry)
	}

	// Opening a connection is done lazily - next DeleteChunks call will connect
	db := sql.OpenDB(connector)

	// TODO: determine reasonable values for this
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(2)
	return &serverImpl{
		deleter:  deleter,
		db:       db,
		deleting: make(map[string]*trigger),
	}, nil
}

func (si *serverImpl) markChunksDeleting(ctx context.Context, chunks []chunk.Chunk) ([]chunk.Chunk, error) {
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, fmt.Sprintf("'%s'", chunk.Hash))
	}
	sort.Strings(chunkIds)

	query := `
update chunks set 
  deleting = now()
where
  chunk in (` + strings.Join(chunkIds, ",") + `) and
  chunk not in (
		select distinct chunk from refs
    where chunk in (` + strings.Join(chunkIds, ",") + `)
	)
returning chunk
	`

	for {
		txn, err := si.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return nil, err
		}

		if err != nil {
			if err := txn.Rollback(); err != nil {
				return nil, err
			}
			if isRetriableError(err) {
				continue
			}
			return nil, err
		}

		cursor, err := txn.QueryContext(ctx, query)
		if err != nil {
			if err := txn.Rollback(); err != nil {
				return nil, err
			}
			if isRetriableError(err) {
				continue
			}
			return nil, err
		}

		// Flush returned chunks through the server
		chunks = readChunksFromCursor(cursor)
		cursor.Close()
		if err := cursor.Err(); err != nil {
			if err := txn.Rollback(); err != nil {
				return nil, err
			}
			if isRetriableError(err) {
				continue
			}
			return nil, err
		}

		if err := txn.Commit(); err != nil {
			if isRetriableError(err) {
				continue
			}
			return nil, err
		}
		break
	}
	return chunks, nil
}

func (si *serverImpl) removeChunkRows(ctx context.Context, chunks []chunk.Chunk) error {
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, fmt.Sprintf("'%s'", chunk.Hash))
	}
	sort.Strings(chunkIds)

	query := `
delete from chunks
where
  chunk in (` + strings.Join(chunkIds, ",") + `) and
	deleting is not null
	`

	_, err := si.db.ExecContext(ctx, query)
	return err
}

func retry(name string, maxAttempts int, fn func() error) {
	var err error
	for attempts := 0; attempts < maxAttempts; attempts++ {
		err = fn()
		if err == nil {
			return
		}
	}
	panic(fmt.Sprintf("'%s' failed %d times, latest error: %v", name, maxAttempts, err))
}

func (si *serverImpl) DeleteChunks(ctx context.Context, chunks []chunk.Chunk) (retErr error) {
	defer func(startTime time.Time) { applyRequestMetrics("DeleteChunks", retErr, startTime) }(time.Now())
	// Spawn goroutine to do this all async
	go func() {
		trigger := newTrigger()

		// Check if we have outstanding deletes for these chunks and save the trigger
		candidates := func() []chunk.Chunk {
			si.mutex.Lock()
			defer si.mutex.Unlock()

			result := []chunk.Chunk{}
			for _, c := range chunks {
				if _, ok := si.deleting[c.Hash]; !ok {
					si.deleting[c.Hash] = trigger
					result = append(result, c)
				}
			}
			return result
		}()

		if len(candidates) == 0 {
			return
		}

		// set the chunks as deleting
		toDelete := []chunk.Chunk{}
		retry("markChunksDeleting", 10, func() error {
			var err error
			toDelete, err = si.markChunksDeleting(context.Background(), candidates)
			return err
		})

		if len(toDelete) > 0 {
			// delete objects from object storage
			retry("deleter.Delete", 10, func() error {
				return si.deleter.Delete(context.Background(), toDelete)
			})

			// delete the rows from the db
			retry("removeChunkRows", 10, func() error {
				return si.removeChunkRows(context.Background(), toDelete)
			})
		}

		func() {
			si.mutex.Lock()
			defer si.mutex.Unlock()
			for _, chunk := range chunks {
				delete(si.deleting, chunk.Hash)
			}
		}()

		trigger.Trigger()
	}()

	return nil
}

func (si *serverImpl) FlushDeletes(ctx context.Context, chunks []chunk.Chunk) (retErr error) {
	defer func(startTime time.Time) { applyRequestMetrics("FlushDeletes", retErr, startTime) }(time.Now())

	triggers := func() []*trigger {
		si.mutex.Lock()
		defer si.mutex.Unlock()

		res := []*trigger{}
		for _, chunk := range chunks {
			trigger := si.deleting[chunk.Hash]
			if trigger != nil {
				res = append(res, trigger)
			}
		}
		return res
	}()

	for _, trigger := range triggers {
		trigger.Wait()
	}

	return nil
}

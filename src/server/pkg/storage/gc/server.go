package gc

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/prometheus/client_golang/prometheus"
)

// Server is the interface that the garbage collector service provides to clients
type Server interface {
	DeleteChunks(context.Context, []chunk.Chunk) error
	FlushDeletes(context.Context, []chunk.Chunk) error
}

// Deleter is an interface that must be provided when creating the garbage
// collector server, it handles removing chunks from the backing object storage.
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
	db       *gorm.DB
	mutex    sync.Mutex
	deleting map[string]*trigger
}

// MakeServer constructs a garbage collector server:
//  * deleter - an object implementing the Deleter interface that will be
//      called for deleting chunks from object storage
//  * postgresHost, postgresPort - the host and port of the postgres instance
//      which is used for coordinating garbage collection reference counts with
//      the garbage collector clients
//  * registry (optional) - a Prometheus stats registry for tracking usage and
//    performance
func MakeServer(
	deleter Deleter,
	postgresHost string,
	postgresPort uint16,
	registry prometheus.Registerer,
) (Server, error) {
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
	return &serverImpl{
		deleter:  deleter,
		db:       db,
		deleting: make(map[string]*trigger),
	}, nil
}

func (si *serverImpl) markChunksDeleting(ctx context.Context, chunks []chunk.Chunk) ([]chunk.Chunk, error) {
	chunkIDs := []string{}
	for _, chunk := range chunks {
		chunkIDs = append(chunkIDs, chunk.Hash)
	}
	sort.Strings(chunkIDs)

	result := []chunkModel{}
	statements := []stmtCallback{
		func(txn *gorm.DB) *gorm.DB {
			refQuery := txn.Debug().Table(refTable).Select("distinct chunk").Where("chunk in (?)", chunkIDs).QueryExpr()

			return txn.Raw(`
update chunks set deleting = now()
where chunk in (?) and chunk not in (?)
returning chunk`, chunkIDs, refQuery).Scan(&result)
		},
	}

	if err := runTransaction(ctx, si.db, statements, markChunksDeletingStats); err != nil {
		return nil, err
	}
	return convertChunks(result), nil
}

func (si *serverImpl) removeChunkRows(ctx context.Context, chunks []chunk.Chunk) ([]chunk.Chunk, error) {
	chunkIDs := []string{}
	for _, chunk := range chunks {
		chunkIDs = append(chunkIDs, chunk.Hash)
	}
	sort.Strings(chunkIDs)

	result := []chunkModel{}
	statements := []stmtCallback{
		func(txn *gorm.DB) *gorm.DB {

			return txn.Raw(`
with del_chunks as (
	delete from chunks
	where chunk in (?) and deleting is not null
	returning chunk
), del_refs as (
	delete from refs using del_chunks
	where refs.sourcetype = 'chunk' and refs.source = del_chunks.chunk
	returning refs.chunk
), deleted as (
 select chunk, count(*) from del_refs group by 1 order by 1
), before as (
 select chunk, count(*) as count from refs join deleted using (chunk) group by 1 order by 1
)

select chunk from before join deleted using (chunk) where before.count - deleted.count = 0
		`, chunkIDs).Scan(&result)
		},
	}

	if err := runTransaction(ctx, si.db, statements, removeChunkRowsStats); err != nil {
		return nil, err
	}
	return convertChunks(result), nil
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
	defer func(start time.Time) { applyRequestStats("DeleteChunks", retErr, start) }(time.Now())
	chunkIDs := []string{}
	for _, c := range chunks {
		chunkIDs = append(chunkIDs, c.Hash)
	}
	fmt.Printf("DeleteChunks: %v\n", strings.Join(chunkIDs, ", "))

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

	// Spawn goroutine to do all the remote calls async
	go func() {
		// set the chunks as deleting
		toDelete := []chunk.Chunk{}
		retry("markChunksDeleting", 10, func() (retErr error) {
			var err error
			toDelete, err = si.markChunksDeleting(context.Background(), candidates)
			return err
		})

		if len(toDelete) > 0 {
			// delete objects from object storage
			retry("deleter.Delete", 10, func() (retErr error) {
				defer func(start time.Time) { applyDeleteStats(retErr, start) }(time.Now())
				chunkIDs := []string{}
				for _, c := range toDelete {
					chunkIDs = append(chunkIDs, c.Hash)
				}
				fmt.Printf("deleting chunks: %v\n", strings.Join(chunkIDs, ", "))
				return si.deleter.Delete(context.Background(), toDelete)
			})

			// delete the rows from the db
			transitiveDeletes := []chunk.Chunk{}
			retry("removeChunkRows", 10, func() (retErr error) {
				var err error
				transitiveDeletes, err = si.removeChunkRows(context.Background(), toDelete)
				return err
			})

			// Pass transitive deletes to a new RPC
			if len(transitiveDeletes) > 0 {
				si.DeleteChunks(context.Background(), transitiveDeletes)
			}
		}

		func() {
			si.mutex.Lock()
			defer si.mutex.Unlock()
			for _, c := range candidates {
				delete(si.deleting, c.Hash)
			}
		}()

		trigger.Trigger()
	}()

	return nil
}

func (si *serverImpl) FlushDeletes(ctx context.Context, chunks []chunk.Chunk) (retErr error) {
	defer func(start time.Time) { applyRequestStats("FlushDeletes", retErr, start) }(time.Now())

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

	for _, t := range triggers {
		t.Wait()
	}

	return nil
}

package gc

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
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
	db       *gorm.DB
	mutex    sync.Mutex
	deleting map[string]*trigger
}

func MakeServer(deleter Deleter, host string, port uint16, registry prometheus.Registerer) (Server, error) {
	if registry != nil {
		initPrometheus(registry)
	}

	db, err := openDatabase(host, port)
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
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, chunk.Hash)
	}
	sort.Strings(chunkIds)

	result := []chunkModel{}
	statements := []stmtCallback{
		func(txn *gorm.DB) *gorm.DB {
			refQuery := txn.Debug().Table(refTable).Select("distinct chunk").Where("chunk in (?)", chunkIds).QueryExpr()

			return txn.Raw(`
update chunks set deleting = now()
where chunk in (?) and chunk not in (?)
returning chunk`, chunkIds, refQuery).Scan(&result)
		},
	}

	if err := runTransaction(si.db, ctx, statements, markChunksDeletingStats); err != nil {
		return nil, err
	}
	return convertChunks(result), nil
}

func (si *serverImpl) removeChunkRows(ctx context.Context, chunks []chunk.Chunk) ([]chunk.Chunk, error) {
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, chunk.Hash)
	}
	sort.Strings(chunkIds)

	// TODO: remove refs from these chunks to others, start a new delete process
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
), counts as (
  select chunk, count(*) - 1 as count from refs join del_refs using (chunk) group by 1 order by 1
)

select chunk from counts where count = 0`, chunkIds).Scan(&result)
		},
	}

	if err := runTransaction(si.db, ctx, statements, removeChunkRowsStats); err != nil {
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
		retry("markChunksDeleting", 10, func() (retErr error) {
			defer func(start time.Time) { applySqlStats("markChunksDeleting", retErr, start) }(time.Now())
			var err error
			toDelete, err = si.markChunksDeleting(context.Background(), candidates)
			return err
		})

		if len(toDelete) > 0 {
			// delete objects from object storage
			retry("deleter.Delete", 10, func() (retErr error) {
				defer func(start time.Time) { applyDeleteStats(retErr, start) }(time.Now())
				return si.deleter.Delete(context.Background(), toDelete)
			})

			// delete the rows from the db
			transitiveDeletes := []chunk.Chunk{}
			retry("removeChunkRows", 10, func() (retErr error) {
				defer func(start time.Time) { applySqlStats("removeChunkRows", retErr, start) }(time.Now())
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
			for _, chunk := range chunks {
				delete(si.deleting, chunk.Hash)
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

	for _, trigger := range triggers {
		trigger.Wait()
	}

	return nil
}

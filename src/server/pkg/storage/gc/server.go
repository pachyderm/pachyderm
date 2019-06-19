package gc

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
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
		chunkIds = append(chunkIds, fmt.Sprintf("'%s'", chunk.Hash))
	}
	sort.Strings(chunkIds)

	for {
		// TODO: wrap with retriable error check
		txn := si.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if txn.Error != nil {
			return nil, txn.Error
		}

		refQuery := txn.Table(refTable).Select("distinct chunk").Where("chunk in (?)", chunkIds).QueryExpr()

		txn.Table(chunkTable).Where("chunk in (?)", chunkIds).Where("chunk not in (?)", refQuery).Update("deleting", gorm.Expr("now()"))
		if txn.Error != nil {
			return nil, txn.Error
		}

		result := []chunkModel{}
		txn.Where("chunk in (?)", chunkIds).Find(&result)
		if txn.Error != nil {
			return nil, txn.Error
		}

		if err := txn.Commit().Error; err != nil {
			return nil, txn.Error
		}
		return convertChunks(result), nil
	}
}

func (si *serverImpl) removeChunkRows(ctx context.Context, chunks []chunk.Chunk) error {
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, fmt.Sprintf("'%s'", chunk.Hash))
	}
	sort.Strings(chunkIds)

	return si.db.Where("chunk in (?)", chunkIds).Where("deleting is not null").Delete(&chunkModel{}).Error
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
	defer func(start time.Time) { applyRequestMetrics("DeleteChunks", retErr, start) }(time.Now())
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
			defer func(start time.Time) { applySqlMetrics("markChunksDeleting", retErr, start) }(time.Now())
			var err error
			toDelete, err = si.markChunksDeleting(context.Background(), candidates)
			return err
		})

		if len(toDelete) > 0 {
			// delete objects from object storage
			retry("deleter.Delete", 10, func() (retErr error) {
				defer func(start time.Time) { applyDeleteMetrics(retErr, start) }(time.Now())
				return si.deleter.Delete(context.Background(), toDelete)
			})

			// delete the rows from the db
			retry("removeChunkRows", 10, func() (retErr error) {
				defer func(start time.Time) { applySqlMetrics("removeChunkRows", retErr, start) }(time.Now())
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
	defer func(start time.Time) { applyRequestMetrics("FlushDeletes", retErr, start) }(time.Now())

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

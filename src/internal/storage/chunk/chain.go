package chunk

import (
	"context"
	"math"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// TaskChain manages a chain of tasks that have a parallel and serial part.
// The parallel part should be executed by the user specified callback and the
// serial part should be executed within the callback passed into the user
// specified callback.
type TaskChain struct {
	eg       *errgroup.Group
	ctx      context.Context
	prevChan chan struct{}
	sem      *semaphore.Weighted
}

// NewTaskChain creates a new task chain.
func NewTaskChain(ctx context.Context, parallelism int64) *TaskChain {
	eg, errCtx := errgroup.WithContext(ctx)
	prevChan := make(chan struct{})
	close(prevChan)
	sem := semaphore.NewWeighted(math.MaxInt64)
	if parallelism > 0 {
		sem = semaphore.NewWeighted(parallelism)
	}
	return &TaskChain{
		eg:       eg,
		ctx:      errCtx,
		prevChan: prevChan,
		sem:      sem,
	}
}

// CreateTask creates a new task in the task chain.
func (c *TaskChain) CreateTask(cb func(context.Context, func(func() error) error) error) error {
	if err := c.sem.Acquire(c.ctx, 1); err != nil {
		return errors.EnsureStack(err)
	}
	scb := c.serialCallback()
	c.eg.Go(func() error {
		defer c.sem.Release(1)
		defer scb(func() error { return nil })
		return cb(c.ctx, scb)
	})
	return nil
}

// Wait waits on the currently executing tasks to finish.
func (c *TaskChain) Wait() error {
	select {
	case <-c.ctx.Done():
		return errors.EnsureStack(c.eg.Wait())
	case <-c.prevChan:
		return nil
	}
}

func (c *TaskChain) serialCallback() func(func() error) error {
	prevChan := c.prevChan
	nextChan := make(chan struct{})
	c.prevChan = nextChan
	var once sync.Once
	return func(cb func() error) error {
		defer once.Do(func() { close(nextChan) })
		select {
		case <-prevChan:
			return cb()
		case <-c.ctx.Done():
			return errors.EnsureStack(c.ctx.Err())
		}
	}
}

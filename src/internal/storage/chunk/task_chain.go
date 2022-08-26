package chunk

import (
	"context"

	"github.com/hashicorp/go-multierror"
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
func NewTaskChain(ctx context.Context, sem *semaphore.Weighted) *TaskChain {
	eg, errCtx := errgroup.WithContext(ctx)
	prevChan := make(chan struct{})
	close(prevChan)
	return &TaskChain{
		eg:       eg,
		ctx:      errCtx,
		prevChan: prevChan,
		sem:      sem,
	}
}

// TaskChainFunc is a function which is computed in parallel, and then returns
// a closure to be computed serially, or an error.
type TaskChainFunc = func(context.Context) (serCB func() error, err error)

// CreateTask creates a new task in the task chain.
func (c *TaskChain) CreateTask(cb TaskChainFunc) error {
	if err := c.sem.Acquire(c.ctx, 1); err != nil {
		err = multierror.Append(c.eg.Wait(), err)
		return errors.EnsureStack(err)
	}
	// get our place in line for the serial portion
	prevChan := c.prevChan
	nextChan := make(chan struct{})
	c.prevChan = nextChan
	// spawn a new goroutine for the parallel portion
	c.eg.Go(func() error {
		defer c.sem.Release(1)
		serCB, err := cb(c.ctx)
		if err != nil {
			return err
		}
		// Either:
		// - There has been an error returned to the errgroup, and the signal will come from the context
		// - There hasn't been an error from anything yet, and we need to wait our turn to do the serial callback
		select {
		case <-c.ctx.Done():
			return errors.EnsureStack(c.ctx.Err())
		case <-prevChan:
			if err := serCB(); err != nil {
				return err
			}
			close(nextChan)
			return nil
		}
	})
	return nil
}

// Wait waits on the currently executing tasks to finish.
func (c *TaskChain) Wait() error {
	return errors.EnsureStack(c.eg.Wait())
}

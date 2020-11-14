package chunk

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// TaskChain manages a chain of tasks that have a parallel and serial part.
// The parallel part should be executed by the user specified callback and the
// serial part should be executed within the callback passed into the user
// specified callback.
type TaskChain struct {
	eg       *errgroup.Group
	ctx      context.Context
	prevChan chan struct{}
}

// NewTaskChain creates a new task chain.
func NewTaskChain(ctx context.Context) *TaskChain {
	eg, errCtx := errgroup.WithContext(ctx)
	prevChan := make(chan struct{})
	close(prevChan)
	return &TaskChain{
		eg:       eg,
		ctx:      errCtx,
		prevChan: prevChan,
	}
}

// CreateTask creates a new task in the task chain.
func (c *TaskChain) CreateTask(cb func(context.Context, func(func() error) error) error) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}
	scb := c.serialCallback()
	c.eg.Go(func() error {
		return cb(c.ctx, scb)
	})
	return nil
}

// Wait waits on the currently executing tasks to finish.
func (c *TaskChain) Wait() error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-c.prevChan:
		return nil
	}
}

func (c *TaskChain) serialCallback() func(func() error) error {
	prevChan := c.prevChan
	nextChan := make(chan struct{})
	c.prevChan = nextChan
	return func(cb func() error) error {
		defer close(nextChan)
		select {
		case <-prevChan:
			return cb()
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	}
}

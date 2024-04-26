package cleanup

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

type CleanupFunc func(context.Context) error

type Cleaner struct {
	cbs   []CleanupFunc
	names []string
}

func (c *Cleaner) AddCleanupCtx(name string, f CleanupFunc) {
	if f == nil {
		return
	}
	c.cbs = append(c.cbs, f)
	c.names = append(c.names, name)
}

func (c *Cleaner) AddCleanup(name string, f func() error) {
	if f == nil {
		return
	}
	c.AddCleanupCtx(name, func(ctx context.Context) error { return f() })
}

func (c *Cleaner) Subsume(x *Cleaner) {
	for i := range x.cbs {
		c.AddCleanupCtx(x.names[i], x.cbs[i])
	}
}

func (c *Cleaner) Cleanup(ctx context.Context) error {
	if c == nil {
		return nil
	}
	var errs error
	for i := len(c.cbs) - 1; i >= 0; i-- {
		f := c.cbs[i]
		name := c.names[i]
		log.Debug(ctx, "clean up "+name)
		if err := f(ctx); err != nil {
			log.Debug(ctx, "failed to clean up "+name, zap.Error(err))
			errors.JoinInto(&errs, errors.Wrapf(err, "clean up %s", name))
			continue
		}
		log.Debug(ctx, "cleaned up "+name+" ok")
	}
	return errs
}

package ctx

import (
	"time"

	"golang.org/x/net/context"
)

type Context struct {
	c1     context.Context
	c2     context.Context
	err    error
	done   chan struct{}
	finish chan struct{}
}

func Merge(c1 context.Context, c2 context.Context) *Context {
	c := &Context{
		c1:     c1,
		c2:     c2,
		done:   make(chan struct{}),
		finish: make(chan struct{}),
	}
	go func() {
		select {
		case <-c1.Done():
			c.err = c1.Err()
			close(c.done)
		case <-c2.Done():
			c.err = c2.Err()
			close(c.done)
		case <-c.finish:
			return
		}
	}()
	return c
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	d1, ok1 := c.c1.Deadline()
	d2, ok2 := c.c2.Deadline()

	if !ok1 && !ok2 {
		return time.Unix(0, 0), false
	}

	if d1.Before(d2) {
		return d1, true
	}
	return d2, true
}

func (c *Context) Done() <-chan struct{} {
	return c.done
}

func (c *Context) Err() error {
	return c.err
}

func (c *Context) Value(key interface{}) interface{} {
	v1 := c.c1.Value(key)
	if v1 != nil {
		return v1
	}

	v2 := c.c2.Value(key)
	if v2 != nil {
		return v2
	}

	return nil
}

func (c *Context) Finish() {
	close(c.finish)
}

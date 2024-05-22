package errors

import (
	"testing"

	"github.com/pkg/errors"
)

var (
	a     = errors.New("a")
	b     = errors.New("b")
	run   = errors.New("run")
	close = errors.New("close")
)

func TestJoin(t *testing.T) {
	testData := []struct {
		err, target error
	}{
		{a, a},
		{b, b},
		{Join(a, b), a},
		{Join(a, b), b},
		{Join(a, a), a},
		{Join(b, b), b},
		{Join(a, Join(b)), a},
		{Join(a, Join(b)), b},
		{Join(nil, a, nil), a},
		{Join(nil, a, Join(nil, b, nil, b)), a},
		{Join(nil, a, Join(nil, b, nil, b)), b},
	}

	for _, test := range testData {
		if !errors.Is(test.err, test.target) {
			t.Errorf("expected Is(%v, %v) to be true", test.err, test.target)
		}
	}

	if err := Join(nil, nil, nil); err != nil {
		t.Errorf("Join(nil...) should be nil; got %v", err)
	}
}

func TestJoinInto(t *testing.T) {
	var err error
	JoinInto(&err, a)
	if !errors.Is(err, a) {
		t.Errorf("err %v should be a", err)
	}
	JoinInto(&err, b)
	if !errors.Is(err, a) {
		t.Errorf("err %v should be a", err)
	}
	if !errors.Is(err, b) {
		t.Errorf("err %v should be b", err)
	}

	err = b
	JoinInto(&err, a)
	if !errors.Is(err, a) {
		t.Errorf("err %v should be a", err)
	}
	if !errors.Is(err, b) {
		t.Errorf("err %v should be b", err)
	}
}

type closeRecorder struct {
	err    error
	closed bool
}

func (c *closeRecorder) Close() error {
	c.closed = true
	return c.err
}

func TestClose(t *testing.T) {
	x := new(closeRecorder)
	err := func() (retErr error) {
		defer Close(&retErr, x, "cleanup")
		if x.closed {
			t.Errorf("x closed too early")
		}
		return nil
	}()
	if !x.closed {
		t.Errorf("x unexpectedly not closed")
	}
	if err != nil {
		t.Errorf("expected nil error; got %v", err)
	}

	x = new(closeRecorder)
	err = func() (retErr error) {
		x.err = close
		defer Close(&retErr, x, "cleanup")
		return nil
	}()
	if !x.closed {
		t.Errorf("x unexpectedly not closed")
	}
	if !errors.Is(err, close) {
		t.Errorf("expected close error; got %v", err)
	}

	x = new(closeRecorder)
	err = func() (retErr error) {
		defer Close(&retErr, x, "cleanup")
		return run //nolint:wrapcheck
	}()
	if !x.closed {
		t.Errorf("x unexpectedly not closed")
	}
	if !errors.Is(err, run) {
		t.Errorf("expected run error; got %v", err)
	}

	x = new(closeRecorder)
	err = func() (retErr error) {
		x.err = close
		defer Close(&retErr, x, "cleanup")
		return run //nolint:wrapcheck
	}()
	if !x.closed {
		t.Errorf("x unexpectedly not closed")
	}
	if !errors.Is(err, run) && !errors.Is(err, close) {
		t.Errorf("expected run and close errors; got %v", err)
	}
}

func TestInvoke(t *testing.T) {
	x := new(closeRecorder)
	err := func() (retErr error) {
		x.err = close
		defer Invoke(&retErr, x.Close, "cleanup")
		if x.closed {
			t.Errorf("x closed too early")
		}
		return run //nolint:wrapcheck
	}()
	if !x.closed {
		t.Errorf("x unexpectedly not closed")
	}
	if !errors.Is(err, run) && !errors.Is(err, close) {
		t.Errorf("expected run and close errors; got %v", err)
	}
}

func TestInvoke1(t *testing.T) {
	x := new(closeRecorder)
	closer := func(close bool) error {
		if close {
			return x.Close()
		}
		return nil
	}
	err := func() (retErr error) {
		x.err = close
		defer Invoke1(&retErr, closer, true, "cleanup")
		Invoke1(&retErr, closer, false, "ignored")
		if x.closed {
			t.Errorf("x closed too early")
		}
		return run //nolint:wrapcheck
	}()
	if !x.closed {
		t.Errorf("x unexpectedly not closed")
	}
	if !errors.Is(err, run) && !errors.Is(err, close) {
		t.Errorf("expected run and close errors; got %v", err)
	}
}

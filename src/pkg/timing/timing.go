package timing

import (
	"sync/atomic"
	"time"
)

type Timer interface {
	Now() time.Time
}

type FakeTimer interface {
	Timer
	Add(deltaSeconds int64, deltaNanos int64)
	Set(unixSeconds int64, unixNanos int64)
}

func NewSystemTimer() Timer {
	return newSystemTimer()
}

func NewFakeTimer() FakeTimer {
	return newFakeTimer()
}

type systemTimer struct{}

func newSystemTimer() *systemTimer {
	return &systemTimer{}
}

func (s *systemTimer) Now() time.Time {
	return time.Now().UTC()
}

type fakeTimer struct {
	unixNanos int64
}

func newFakeTimer() *fakeTimer {
	return &fakeTimer{0}
}

func (f *fakeTimer) Now() time.Time {
	unixNanos := f.unixNanos
	return time.Unix(
		unixNanos/int64(time.Second),
		unixNanos%int64(time.Second),
	).UTC()
}

func (f *fakeTimer) Add(deltaSeconds int64, deltaNanos int64) {
	delta := (deltaSeconds * int64(time.Second)) + deltaNanos
	atomic.AddInt64(&f.unixNanos, delta)
}

func (f *fakeTimer) Set(unixSeconds int64, unixNanos int64) {
	unix := (unixSeconds * int64(time.Second)) + unixNanos
	atomic.StoreInt64(&f.unixNanos, unix)
}

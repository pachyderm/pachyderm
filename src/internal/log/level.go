package log

import (
	"sync/atomic"
	"time"

	"go.uber.org/zap/zapcore"
)

// resettableLevel is a log level (zapcore.LevelEnabler) that can be changed at runtime either
// forever (SetLevel) or for a set duration (SetLevelFor).  See also: zap.AtomicLevel.
//
// A resettableLevel must be initialized before use with NewResettableLevelAt.  A resettableLevel's
// methods take a value receiver; the contained pointers are the actual value of the object (and do
// not change as the level is changed).
type resettableLevel struct {
	orig, cur *atomic.Int32
	reset     *atomic.Pointer[time.Timer]
}

var _ zapcore.LevelEnabler = resettableLevel{}

// NewResettableLevelAt creates a new resettable level set to the provided level, and set to revert
// to that level if changed with SetLevelFor.
func NewResettableLevelAt(l zapcore.Level) resettableLevel {
	rl := resettableLevel{
		orig:  new(atomic.Int32),
		cur:   new(atomic.Int32),
		reset: new(atomic.Pointer[time.Timer]),
	}
	rl.orig.Store(int32(l))
	rl.cur.Store(int32(l))
	return rl
}

// Level returns the current level.
func (rl resettableLevel) Level() zapcore.Level {
	return zapcore.Level(rl.cur.Load())
}

// Enabled implements zapcore.LevelEnabler
func (rl resettableLevel) Enabled(l zapcore.Level) bool {
	return rl.Level().Enabled(l)
}

// String implements fmt.Stringer.
func (rl resettableLevel) String() string {
	return rl.Level().String()
}

// UnmarshalText implements encoding.TextUnmarshaler.  The unmarshaled level is Set() as the current
// level; overriding any "original" the level was initialized with.
func (rl resettableLevel) UnmarshalText(text []byte) error {
	var l zapcore.Level
	if err := l.UnmarshalText(text); err != nil {
		return err //nolint:wrapcheck
	}
	rl.SetLevel(l)
	return nil
}

// SetLevel sets the current and original level, and cancels any pending level revert.
func (rl resettableLevel) SetLevel(l zapcore.Level) {
	if t := rl.reset.Swap(nil); t != nil {
		t.Stop()
	}
	rl.orig.Store(int32(l))
	rl.cur.Store(int32(l))
}

// SetLevelFor sets the current level for a period of time, after which the level will revert to the
// stored original level.  notify, if non-nil, will run after the revert.
func (rl resettableLevel) SetLevelFor(l zapcore.Level, d time.Duration, notify func(from, to string)) {
	wasSet := make(chan struct{})
	t := rl.reset.Swap(time.AfterFunc(d, func() {
		<-wasSet
		to := rl.orig.Load()
		from := rl.cur.Swap(to)
		if notify != nil {
			notify(zapcore.Level(from).String(), zapcore.Level(to).String())
		}
	}))
	if t != nil {
		t.Stop()
	}
	rl.cur.Store(int32(l))
	close(wasSet) // Avoid reverting before we actually change the log level.
}

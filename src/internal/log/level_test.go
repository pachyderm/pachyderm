package log

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestResettableLevel_Basics(t *testing.T) {
	l := NewResettableLevelAt(zapcore.DebugLevel)
	if got, want := l.Level(), zapcore.DebugLevel; got != want {
		t.Errorf("initial level:\n  got: %v\n want: %v", got, want)
	}
	if err := l.UnmarshalText([]byte("INFO")); err != nil {
		t.Errorf("UnmarshalText(INFO): unexpected error: %v", err)
	}
	if got, want := l.String(), "info"; got != want {
		t.Errorf("String():\n  got: %v\n want: %v", got, want)
	}
	if err := l.UnmarshalText([]byte("blah")); err == nil {
		t.Error("UnmarshalText(blah): expected error")
	}
	if got, want := l.String(), "info"; got != want {
		t.Errorf("String(): after invalid unmarshal:\n  got: %v\n want: %v", got, want)
	}

	l.SetLevel(zapcore.ErrorLevel)
	if got, want := l.Level(), zapcore.ErrorLevel; got != want {
		t.Errorf("SetLevel(ErrorLevel):\n  got: %v\n want: %v", got, want)
	}

	l.SetLevelFor(zapcore.DebugLevel, time.Second, nil)
	if got, want := l.Level(), zapcore.DebugLevel; got != want {
		t.Errorf("SetLevelFor(DebugLevel, 1s, ...):\n  got: %v\n want: %v", got, want)
	}
}

func TestResettableLevel_Notify(t *testing.T) {
	l := NewResettableLevelAt(zapcore.ErrorLevel)

	ch := make(chan string)
	l.SetLevelFor(zapcore.DebugLevel, 100*time.Millisecond, func(from, to string) {
		ch <- fmt.Sprintf("%v -> %v", from, to)
	})
	if got, want := l.Level(), zapcore.DebugLevel; got != want {
		t.Errorf("SetLevelFor(DebugLevel, 100ms, ...):\n  got: %v\n want: %v", got, want)
	}
	var gotChange string
	select {
	case gotChange = <-ch:
	case <-time.After(200 * time.Millisecond):
		gotChange = "<timeout after 200ms>"
	}
	if got, want := gotChange, "debug -> error"; got != want {
		t.Errorf("notify callback:\n  got: %v\n want: %v", got, want)
	}
	if got, want := l.Level(), zapcore.ErrorLevel; got != want {
		t.Errorf("level after revert:\n  got: %v\n want: %v", got, want)
	}
}

func TestResettableLevel_Override(t *testing.T) {
	l := NewResettableLevelAt(zapcore.ErrorLevel)

	ch := make(chan string)
	l.SetLevelFor(zapcore.DebugLevel, 100*time.Millisecond, func(from, to string) {
		ch <- fmt.Sprintf("%v -> %v", from, to)
	})
	if got, want := l.Level(), zapcore.DebugLevel; got != want {
		t.Errorf("SetLevelFor(DebugLevel, 100ms, ...):\n  got: %v\n want: %v", got, want)
	}

	l.SetLevel(zapcore.InfoLevel)
	if got, want := l.Level(), zapcore.InfoLevel; got != want {
		t.Errorf("SetLevel(InfoLevel):\n  got: %v\n want: %v", got, want)
	}

	select {
	case <-ch:
		t.Fatal("notify function ran unexpectedly")
	case <-time.After(300 * time.Millisecond):
		// Expected.
	}
	if got, want := l.Level(), zapcore.InfoLevel; got != want {
		t.Errorf("level after revert:\n  got: %v\n want: %v", got, want)
	}
}

func TestResettableLevel_QuickRevert(t *testing.T) {
	l := NewResettableLevelAt(zapcore.ErrorLevel)

	ch := make(chan struct{})
	l.SetLevelFor(zapcore.DebugLevel, time.Nanosecond, func(from, to string) {
		close(ch)
	})
	// While writing this test, I added a sleep(10ms) right before the actual level change, and
	// observed this test passing.  Removing the synchronization channel similarly causes the
	// test to correctly fail.
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for revert")
	}
	if got, want := l.Level(), zapcore.ErrorLevel; got != want {
		t.Errorf("level after revert:\n  got: %v\n want: %v", got, want)
	}
}

func TestResettableLevel_ManyChanges(t *testing.T) {
	l := NewResettableLevelAt(zapcore.ErrorLevel)

	ch := make(chan string)
	var wasEverInfo atomic.Bool
	go func() {
		// This simulates a logger running concurrently with level changes and lets the race
		// detector find concurrency bugs.
		for {
			if l.Level() == zapcore.InfoLevel {
				wasEverInfo.Store(true)
				return
			}
		}
	}()
	for i := 0; i < 20; i++ {
		level := zapcore.InfoLevel
		if i%2 == 1 {
			level = zapcore.DebugLevel
		}
		l.SetLevelFor(level, 100*time.Millisecond, func(from, to string) {
			ch <- fmt.Sprintf("%v -> %v", from, to)
		})
		time.Sleep(time.Millisecond)
	}
	if got, want := l.Level(), zapcore.DebugLevel; got != want {
		t.Errorf("SetLevelFor(DebugLevel, 100ms, ...):\n  got: %v\n want: %v", got, want)
	}
	var gotChange string
	select {
	case gotChange = <-ch:
		close(ch) // make other senders panic
	case <-time.After(200 * time.Millisecond):
		gotChange = "<timeout after 200ms>"
	}
	if got, want := gotChange, "debug -> error"; got != want {
		t.Errorf("notify callback:\n  got: %v\n want: %v", got, want)
	}
	if got, want := wasEverInfo.Load(), true; got != want {
		t.Errorf("was ever at level info?\n  got: %v\n want: %v", got, want)
	}
	if got, want := l.Level(), zapcore.ErrorLevel; got != want {
		t.Errorf("level after revert:\n  got: %v\n want: %v", got, want)
	}
}

func TestResettableLevel_Copy(t *testing.T) {
	l := NewResettableLevelAt(zapcore.InfoLevel)
	l2 := l
	l.SetLevel(zapcore.DebugLevel)
	if got, want := l2.Level(), zapcore.DebugLevel; got != want {
		t.Errorf("copied level:\n  got: %v\n want: %v", got, want)
	}
}

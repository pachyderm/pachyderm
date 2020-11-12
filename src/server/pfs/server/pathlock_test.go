package server

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// TestPathLockSomeCollisions starts a bunch of threads for paths, some of which
// overlap and some of which don't. This is the main way that we expect PutFile
// to work.
func TestPathLockSomeCollisions(t *testing.T) {
	var (
		seed       = time.Now().UnixNano()
		r          = rand.New(rand.NewSource(seed))
		running    = make(map[string]bool) // paths held by currently running ops
		runningMu  sync.Mutex              // guards 'running'
		runningStr = func() []string {     // print 'running' as a list, not a map
			strs := make([]string, 0, len(running))
			for k := range running {
				strs = append(strs, k)
			}
			return strs
		}
		start, end   = true, false                  // syntax sugar for checkRunning
		maxRunning   = 0                            // count the max # of running ops
		checkRunning = func(op bool, path string) { // check 'running' for prefix collisions
			runningMu.Lock()
			defer runningMu.Unlock()
			if len(running) > maxRunning {
				maxRunning = len(running)
			}
			for p := range running {
				if op == end && path == p {
					continue // we expect path itself to be in running if we're removing it
				}
				// no prefix of 'path' is running (as this isn't the main test goro, use
				// Errorf instead of require.False so that execution isn't halted on
				// failure, preventing deadlock)
				if strings.HasPrefix(path, p) {
					t.Errorf("(seed: %d) path %q has prefix %s ∈ %v", seed, path, p, runningStr())
				}
				// 'path' is not a prefix of any running op
				if strings.HasPrefix(p, path) {
					t.Errorf("(seed: %d) path %q is a prefix of %s ∈ %v", seed, path, p, runningStr())
				}
			}
			if op == start {
				running[path] = true
			} else {
				delete(running, path)
			}
		}
		p  = newPathLock(nil) // the data structure under test
		eg errgroup.Group     // for running the subprocesses
	)
	// Iterate through [0,1000) in a random order, represented as binary strings.
	// This induces regular collisions (these strings aren't fixed-width, so small
	// numbers' representations are often prefixes of larger numbers), but also
	// allows for some parallelism.
	//
	// This test uses 'maxRunning' to check that some parallelism occurs, and the
	// prevention of collisions can be verified by removing 'p.start(path)' and
	// 'p.finish(path)' below, which causes the collision checks in 'checkRunning'
	// to fail.
	for _, i := range r.Perm(1000) {
		// small nums are prefixes of bigger nums. Iterate all nums but jump around
		// to induce collisions
		path := fmt.Sprintf("%b", i)
		require.NoError(t, p.start(path))
		eg.Go(func() error {
			defer func() { require.NoError(t, p.finish(path)) }()
			checkRunning(start, path)
			// force later prefixes to run ~first
			time.Sleep(time.Duration(100-i) * time.Millisecond)
			checkRunning(end, path)
			return nil
		})
	}
	require.NoError(t, eg.Wait())
	// want at least some concurrency in this test
	require.True(t, maxRunning > 1, maxRunning)
}

// TestPathLockAllCollisions is the strictest test, and starts several
// goroutines where every goroutine is given a path that extends the previous
// goroutines' paths. There should be no overlap, and later goroutines should
// wait on earlier goroutines
func TestPathLockAllCollisions(t *testing.T) {
	p := newPathLock(nil)
	// Lock every prefix of 'paths' and make sure that they're processed in order
	paths := strings.Repeat("a", 100)
	ch := make(chan string)
	// write all paths in a separate goro so the main goro can read from 'ch'
	go func() {
		var eg errgroup.Group
		for i := range paths {
			path := paths[:i]
			require.NoError(t, p.start(path))
			eg.Go(func() error {
				defer func() { require.NoError(t, p.finish(path)) }()
				// force later prefixes to run first
				time.Sleep(time.Duration(100-i) * time.Millisecond)
				ch <- path
				return nil
			})
		}
		require.NoError(t, eg.Wait())
		close(ch)
	}()

	var prev string
	for p := range ch {
		if prev != "" {
			require.True(t, len(prev) < len(p), "expected %q to be shorter than %q", p, prev)
		}
		prev = p
	}
}

// TestPathLockAllCollisionsRev is similar to TestPathLockAllCollisions, but it
// iterates over the colliding paths from longest to shortest, to ensure that
// pathlock handles that direction as well (blocking paths if they're a prefix
// of a running operation's path)
func TestPathLockAllCollisionsRev(t *testing.T) {
	p := newPathLock(nil)
	// Lock every prefix of 'paths' and make sure that they're processed in order
	paths := strings.Repeat("a", 100)
	ch := make(chan string)
	// write all paths in a separate goro so the main goro can read from 'ch'
	go func() {
		var eg errgroup.Group
		for i := range paths {
			path := paths[:len(paths)-1-i]
			require.NoError(t, p.start(path))
			eg.Go(func() error {
				defer func() { require.NoError(t, p.finish(path)) }()
				// force later prefixes to run first
				time.Sleep(time.Duration(100-i) * time.Millisecond)
				ch <- path
				return nil
			})
		}
		require.NoError(t, eg.Wait())
		close(ch)
	}()

	var prev string
	for p := range ch {
		if prev != "" {
			require.True(t, len(prev) > len(p), "expected %q to be shorter than %q", p, prev)
		}
		prev = p
	}
}

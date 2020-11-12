package server

import (
	gopath "path"
	"sort"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// pathlock is used to prevent two PutFile calls that affect the same path (e.g.
// Put("/foo") and Delete("/")) from racing. The rule is: before running any
// PutFile operation B, make sure there is no running PutFile operation A such
// that B.Path is a prefix of A.Path or vice-versa.
//
// Note that pathlock can only block one goro at a time. This works for PutFile
// (where requests are arriving in a stream, and the whole stream stream can be
// paused until conflicting paths are finished), but you could not use this to
// schedule N goroutines and have each independently call start() (i.e. many
// goros blocking at once). Not only would it be complicated to track which
// goros are blocked on what, this use-case would introduce complicated
// fairness/livelock issues that pathlock doesn't attempt to solve.
type pathlock struct {
	sync.Mutex
	sync.Cond

	// running contains the set of paths being written to by some in-progress
	// PutFile operation (we use a set of sorted strings rather than a map so that
	// we can quickly find prefix collisions)
	running []string

	// blocking is an in-progress PutFile operation that is blocking the next
	// PutFile operation in the stream. Before exiting, each putFile goro checks
	// if it's the blocking operation, and if so, signals to the blocked goro that
	// it may proceed
	blocking string

	// This is used to limit the total concurrency of putFile request. It's
	// included in pathlock as a convenience, as everywhere that putFile calls
	// pathlock.start(), it would also call putFileLimiter.Acquire() (and likewise
	// for pathlock.finish() and putFileLimiter.Release). This is unset for tests.
	putFileLimiter limit.ConcurrencyLimiter
}

func newPathLock(putFileLimiter limit.ConcurrencyLimiter) *pathlock {
	p := pathlock{
		putFileLimiter: putFileLimiter,
	}
	p.Cond.L = &p.Mutex
	return &p
}

func (p *pathlock) start(path string) (retErr error) {
	// add leading slash, so '/' collides w/ everything.
	// TODO(msteffen): would paths.go (in this package) or s/s/pkg/paths be
	// better? The latter doesn't exist in 1.11.x, and the former adds a leading
	// slash but doesn't clean the path.
	path = gopath.Clean("/" + path)
	defer func() {
		if retErr == nil && p.putFileLimiter != nil {
			p.putFileLimiter.Acquire()
		}
	}()
	p.Lock()
	defer p.Unlock()
	for { // Check block condition in loop--see sync.Cond documentation
		if blocking := func() string {
			// Find if there's any path P in 'running' such that 'path' is a prefix of
			// P
			//
			// minGe is the smallest entry in 'running' that is greater than or equal
			// to 'path'. If there is a P in running such that 'path' is a prefix of
			// P, then 'path' must be a prefix of 'running[minGe]'[1]
			minGe := sort.SearchStrings(p.running, path)
			if minGe < len(p.running) && strings.HasPrefix(p.running[minGe], path) {
				return p.running[minGe]
			}
			// Find if there's any path P in 'running' such that P is a prefix of
			// 'path'. Do this by checking if every prefix of 'path' is in 'running'.
			for i := len(path) - 1; i > 0 && minGe > 0; i-- {
				// Loop invariant: minGe is the smallest entry in running that is >=
				//   path[:i+1] (note loop starts at len-1).
				//   path[:i] < path[:i+1] --> path[:i] must be in [0,minGe)
				// Loop invariant: path[:i] and [0, minGe) are non-empty
				minGe = sort.SearchStrings(p.running[:minGe], path[:i])
				if minGe < len(p.running) && p.running[minGe] == path[:i] {
					return p.running[minGe]
				}
			}
			return ""
		}(); blocking != "" {
			if p.blocking != "" {
				return errors.Errorf("Internal error: PutFile operation on %q cannot block on %q as an operation is already waiting on %q",
					path, blocking, p.blocking)
			}
			p.blocking = blocking
			p.Wait() // yield p
		} else {
			break
		}
	}
	minGe := sort.SearchStrings(p.running, path)
	p.running = append(p.running, "")
	copy(p.running[minGe+1:], p.running[minGe:]) // OK even if minGe == len(running)
	p.running[minGe] = path
	return nil
}

func (p *pathlock) finish(path string) (retErr error) {
	// add leading slash, so '/' collides w/ everything (see start())
	path = gopath.Clean("/" + path)
	defer func() {
		if p.putFileLimiter != nil {
			// Call Release() even if finish() returns an error, as the semaphore is
			// currently locked by the caller.
			p.putFileLimiter.Release()
		}
	}()
	p.Lock()
	defer p.Unlock()
	// remove 'path' from 'running
	minGe := sort.SearchStrings(p.running, path)
	if minGe == len(p.running) || p.running[minGe] != path {
		return errors.Errorf("no path lock found for PutFile operation on %q", path)
	}
	copy(p.running[minGe:], p.running[minGe+1:]) // OK even if (minGe+1) == len(running)
	p.running = p.running[:len(p.running)-1]

	// Wake up any blocked threads
	if p.blocking == path {
		p.blocking = "" // reset block
		// Both Signal() and Broadcast() *should* be safe, b/c there should only
		// ever be one blocked goro, but I'm always a little worried about having
		// introduced some subtle deadlock, and because and start() wraps p.Wait()
		// in a loop anyway, Broadcast() seems like the safer option
		p.Broadcast()
	}
	return nil
}

// [1] Proof:
// By assumption, we know I = { i ∈ ℕ | path is a prefix of running[i] }
// is nonempty. Let i be the smallest element of I (well-ordering).
//
// Because path is a prefix of running[i], we know path <= running[i], therefore
// minGe <= i (by the definition of minGe). Then, because
//                   path <= running[minGe] <= running[i]
// path must be a prefix of running[minGe] as well.
//
// (if there were some j s.t.  j < len(path) && path[j] != running[minGe], then
// let j be minimal (w.o.) and we have either 1) path[j] < running[minGe][j]
// --> running[i][j] < running[minGe][j] --> running[i] < running[minGe], a
// contradition. Or 2) path[j] > running[minGe][j] --> path > running[minGe], a
// contradiction)

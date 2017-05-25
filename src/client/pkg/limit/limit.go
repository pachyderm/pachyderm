// Package limit provides primitives to limit concurrency.
//
// Note that this is not to be confused with rate-limiting.  With concurrency
// limiting (which is what this package does), you are limiting the number of
// operations that can be running at any given point in time.  With rate
// limiting, you are limiting the number of operations that can be fired
// within a given time window.
//
// For instance, even if you limit concurrency to 1, you can still have N
// requests per second where N is an arbitrarily large number, given that
// each request takes 1/N second to complete.
package limit

// ConcurrencyLimiter limits the number of concurrent operations
// If the ConcurrencyLimiter is initialized with a concurrency of 0, then
// all of the following functions will be no-ops, meaning that an arbitrary
// concurrency is allowed.
type ConcurrencyLimiter interface {
	// Acquire acquires the right to proceed.  It blocks if the concurrency
	// limit has been reached.
	Acquire()
	// Release signals that an operation has completed.
	Release()
	// Wait blocks until all operations that have called Acquire thus far
	// are completed.
	Wait()
}

// New returns a new ConcurrencyLimiter with the given limit
func New(concurrency int) ConcurrencyLimiter {
	if concurrency == 0 {
		return &noOpLimiter{}
	}
	return &concurrencyLimiter{make(chan struct{}, concurrency)}
}

type concurrencyLimiter struct {
	sem chan struct{}
}

func (c *concurrencyLimiter) Acquire() {
	c.sem <- struct{}{}
}

func (c *concurrencyLimiter) Release() {
	<-c.sem
}

func (c *concurrencyLimiter) Wait() {
	for i := 0; i < cap(c.sem); i++ {
		c.sem <- struct{}{}
	}
}

type noOpLimiter struct{}

func (n *noOpLimiter) Acquire() {}

func (n *noOpLimiter) Release() {}

func (n *noOpLimiter) Wait() {}

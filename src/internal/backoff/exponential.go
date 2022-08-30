package backoff

import (
	"math/rand"
	"time"
)

/*
ExponentialBackOff is a backoff implementation that increases the backoff
period for each retry attempt using a randomization function that grows exponentially.

NextBackOff() is calculated using the following formula:

 randomized interval =
     RetryInterval * (random value in range [1 - RandomizationFactor, 1 + RandomizationFactor])

In other words NextBackOff() will range between the randomization factor
percentage below and above the retry interval.

For example, given the following parameters:

 RetryInterval = 2
 RandomizationFactor = 0.5
 Multiplier = 2

the actual backoff period used in the next retry attempt will range between 1 and 3 seconds,
multiplied by the exponential, that is, between 2 and 6 seconds.

Note: MaxInterval caps the RetryInterval and not the randomized interval.

If the time elapsed since an ExponentialBackOff instance is created goes past the
MaxElapsedTime, then the method NextBackOff() starts returning backoff.Stop.

The elapsed time can be reset by calling Reset().

Example: Given the following default arguments, for 10 tries the sequence will be,
and assuming we go over the MaxElapsedTime on the 10th try:

 Request #  RetryInterval (seconds)  Randomized Interval (seconds)

  1          0.5                     [0.25,   0.75]
  2          0.75                    [0.375,  1.125]
  3          1.125                   [0.562,  1.687]
  4          1.687                   [0.8435, 2.53]
  5          2.53                    [1.265,  3.795]
  6          3.795                   [1.897,  5.692]
  7          5.692                   [2.846,  8.538]
  8          8.538                   [4.269, 12.807]
  9         12.807                   [6.403, 19.210]
 10         19.210                   backoff.Stop

Note: Implementation is not thread-safe.
*/
type ExponentialBackOff struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	// After MaxElapsedTime the ExponentialBackOff stops.
	// It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration
	Clock          Clock

	CurrentInterval time.Duration
	StartTime       time.Time
}

// Clock is an interface that returns current time for BackOff.
type Clock interface {
	Now() time.Time
}

// Default values for ExponentialBackOff.
const (
	DefaultInitialInterval     = 500 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 1.5
	DefaultMaxInterval         = 60 * time.Second
	DefaultMaxElapsedTime      = 15 * time.Minute
)

// withCanonicalRandomizationFactor is a utility function used by all
// NewXYZBackoff functions to clamp b.RandomizationFactor to either 0 or 1
func (b *ExponentialBackOff) withCanonicalRandomizationFactor() *ExponentialBackOff {
	if b.RandomizationFactor < 0 {
		b.RandomizationFactor = 0
	} else if b.RandomizationFactor > 1 {
		b.RandomizationFactor = 1
	}
	return b
}

// withReset is a utility function that calls 'b.Reset()' and then returns it,
// so that all NewXYZBackoff functions can reset their result and return it
// inline
func (b *ExponentialBackOff) withReset() *ExponentialBackOff {
	b.Reset()
	return b
}

// NewExponentialBackOff creates an instance of ExponentialBackOff using default values.
func NewExponentialBackOff() *ExponentialBackOff {
	b := &ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		MaxElapsedTime:      DefaultMaxElapsedTime,
		Clock:               SystemClock,
	}
	return b.withCanonicalRandomizationFactor().withReset()
}

// NewInfiniteBackOff creates an instance of ExponentialBackOff that never
// ends.
func NewInfiniteBackOff() *ExponentialBackOff {
	b := &ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         15 * time.Second,
		MaxElapsedTime:      0,
		Clock:               SystemClock,
	}
	return b.withCanonicalRandomizationFactor().withReset()
}

// NewTestingBackOff returns a backoff tuned towards waiting for a Pachyderm
// state change in a test
func NewTestingBackOff() *ExponentialBackOff {
	b := &ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         5 * time.Second,
		MaxElapsedTime:      60 * time.Second,
		Clock:               SystemClock,
	}
	return b.withCanonicalRandomizationFactor().withReset()
}

// New60sBackOff returns a backoff identical to New10sBackOff except with a
// longer MaxElapsedTime This may be more useful for watcher and controllers
// (e.g. the PPS master or the worker) than New10sBackOff, which is a length of
// time that makes more sense for the critical paths of slow RPCs (e.g. PutFile)
func New60sBackOff() *ExponentialBackOff {
	b := &ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         2 * time.Second,
		MaxElapsedTime:      60 * time.Second,
		Clock:               SystemClock,
	}
	return b.withCanonicalRandomizationFactor().withReset()
}

type systemClock struct{}

func (t systemClock) Now() time.Time {
	return time.Now()
}

// SystemClock implements Clock interface that uses time.Now().
var SystemClock = systemClock{}

// Reset the interval back to the initial retry interval and restarts the timer.
func (b *ExponentialBackOff) Reset() {
	b.CurrentInterval = b.InitialInterval
	b.StartTime = b.Clock.Now()
}

// NextBackOff calculates the next backoff interval using the formula:
// 	Randomized interval = RetryInterval +/- (RandomizationFactor * RetryInterval)
func (b *ExponentialBackOff) NextBackOff() time.Duration {
	// Make sure we have not gone over the maximum elapsed time.
	if b.MaxElapsedTime != 0 && b.GetElapsedTime() > b.MaxElapsedTime {
		return Stop
	}
	defer b.incrementCurrentInterval()
	return GetRandomValueFromInterval(b.RandomizationFactor, rand.Float64(), b.CurrentInterval)
}

// GetElapsedTime returns the elapsed time since an ExponentialBackOff instance
// is created and is reset when Reset() is called.
//
// The elapsed time is computed using time.Now().UnixNano().
func (b *ExponentialBackOff) GetElapsedTime() time.Duration {
	return b.Clock.Now().Sub(b.StartTime)
}

// Increments the current interval by multiplying it with the multiplier.
func (b *ExponentialBackOff) incrementCurrentInterval() {
	// Check for overflow, if overflow is detected set the current interval to the max interval.
	if float64(b.CurrentInterval) >= float64(b.MaxInterval)/b.Multiplier {
		b.CurrentInterval = b.MaxInterval
	} else {
		b.CurrentInterval = time.Duration(float64(b.CurrentInterval) * b.Multiplier)
	}
}

// Returns a random value from the following interval:
// 	[randomizationFactor * currentInterval, randomizationFactor * currentInterval].
func GetRandomValueFromInterval(randomizationFactor, random float64, currentInterval time.Duration) time.Duration {
	var delta = randomizationFactor * float64(currentInterval)
	var minInterval = float64(currentInterval) - delta
	var maxInterval = float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minInterval + (random * (maxInterval - minInterval + 1)))
}

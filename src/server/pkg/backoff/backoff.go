// Package backoff implements backoff algorithms for retrying operations.
//
// Use Retry function for retrying operations that may fail.
// If Retry does not meet your needs,
// copy/paste the function into your project and modify as you wish.
//
// There is also Ticker type similar to time.Ticker.
// You can use it if you need to work with channels.
//
// See Examples section below for usage examples.
package backoff

import "time"

// BackOff is a backoff policy for retrying an operation.
type BackOff interface {
	// NextBackOff returns the duration to wait before retrying the operation,
	// or backoff.Stop to indicate that no more retries should be made.
	//
	// Example usage:
	//
	// 	duration := backoff.NextBackOff();
	// 	if (duration == backoff.Stop) {
	// 		// Do not retry operation.
	// 	} else {
	// 		// Sleep for duration and retry operation.
	// 	}
	//
	NextBackOff() time.Duration

	// Reset to initial state.
	Reset()
}

// Stop indicates that no more retries should be made for use in NextBackOff().
const Stop time.Duration = -1

// ZeroBackOff is a fixed backoff policy whose backoff time is always zero,
// meaning that the operation is retried immediately without waiting, indefinitely.
type ZeroBackOff struct{}

// Reset ...
func (b *ZeroBackOff) Reset() {}

// NextBackOff ...
func (b *ZeroBackOff) NextBackOff() time.Duration { return 0 }

// StopBackOff is a fixed backoff policy that always returns backoff.Stop for
// NextBackOff(), meaning that the operation should never be retried.
type StopBackOff struct{}

// Reset ...
func (b *StopBackOff) Reset() {}

// NextBackOff ...
func (b *StopBackOff) NextBackOff() time.Duration { return Stop }

// ConstantBackOff is a backoff policy that always returns the same backoff delay.
// This is in contrast to an exponential backoff policy,
// which returns a delay that grows longer as you call NextBackOff() over and over again.
//
// Note: Implementation is not thread-safe
type ConstantBackOff struct {
	Interval time.Duration

	// After MaxElapsedTime the ConstantBackOff stops.
	// It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration
	startTime      time.Time
}

// Reset ...
func (b *ConstantBackOff) Reset() {
	b.startTime = time.Now()
}

// GetElapsedTime returns the elapsed time since an ExponentialBackOff instance
// is created and is reset when Reset() is called.
//
// The elapsed time is computed using time.Now().UnixNano().
func (b *ConstantBackOff) GetElapsedTime() time.Duration {
	return time.Since(b.startTime)
}

// NextBackOff ...
func (b *ConstantBackOff) NextBackOff() time.Duration {
	if b.MaxElapsedTime != 0 && b.GetElapsedTime() > b.MaxElapsedTime {
		return Stop
	}
	return b.Interval
}

// NewConstantBackOff ...
func NewConstantBackOff(d time.Duration) *ConstantBackOff {
	return &ConstantBackOff{
		Interval: d,
	}
}

// RetryEvery is an alias for NewConstantBackoff, with a nicer name for inline
// calls
func RetryEvery(d time.Duration) *ConstantBackOff {
	return NewConstantBackOff(d)
}

// For sets b.MaxElapsedTime to 'maxElapsed' and returns b
func (b *ConstantBackOff) For(maxElapsed time.Duration) *ConstantBackOff {
	b.MaxElapsedTime = maxElapsed
	return b
}

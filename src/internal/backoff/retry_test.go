package backoff_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	log "github.com/sirupsen/logrus"
)

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return nil
		}

		log.Println("error")
		return errors.New("error")
	}

	err := backoff.Retry(f, backoff.NewExponentialBackOff())
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryUntilCancel(t *testing.T) {
	var results []int
	ctx, cancel := context.WithCancel(context.Background())
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		results = append(results, len(results))
		if len(results) >= 3 {
			cancel()
		}
		return backoff.ErrContinue
	}, &backoff.ZeroBackOff{}, backoff.NotifyContinue(func(err error, d time.Duration) error {
		panic("this should not be called due to cancellation")
	}))
	require.Equal(t, []int{0, 1, 2}, results)
}

func TestRetryUntilCancelZeroBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	first := true
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		if first {
			first = false
		} else {
			t.Fatal("operation() should only run once")
		}
		return nil
	}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
		cancel() // This should prevent operation() from running
		return nil
	})
}

// if mustResetBackOff is called more than once without Reset() being called in
// between, then it returns Stop (i.e. it can back off exactly once). This is
// used by TestRetryUntilCancelResetsOnErrContinue
type mustResetBackOff int

func (m *mustResetBackOff) NextBackOff() time.Duration {
	if (*m) > 0 {
		return backoff.Stop
	}
	(*m)++
	return 0
}

// Reset to initial state.
func (m *mustResetBackOff) Reset() {
	(*m) = 0
}

// TestRetryUntilCancelResetsOnErrContinue tests that the backoff passed to
// RetryUntilCancel is reset if 'operation' returns ErrContinue
func TestRetryUntilCancelResetsOnErrContinue(t *testing.T) {
	var b mustResetBackOff = 0
	var results []int
	backoff.RetryUntilCancel(context.Background(), func() error { //nolint:errcheck
		results = append(results, len(results)+1)
		if len(results) < 3 {
			return backoff.ErrContinue
		}
		return nil
	}, &b, backoff.NotifyContinue(t.Name()))
	require.Equal(t, []int{1, 2, 3}, results)
}

func TestNotifyContinue(t *testing.T) {
	t.Run("NotifyContinueWithNil", func(t *testing.T) {
		var results []int
		backoff.RetryNotify(func() error { //nolint:errcheck
			results = append(results, len(results))
			if len(results) < 3 {
				return backoff.ErrContinue // notify fn below will be elided by NotifyContinue
			}
			return errors.New("done")
		}, &backoff.ZeroBackOff{}, backoff.NotifyContinue(nil))
		require.Equal(t, []int{0, 1, 2}, results)
	})

	t.Run("NotifyContinueWithNotify", func(t *testing.T) {
		var results []int
		backoff.RetryNotify(func() error { //nolint:errcheck
			results = append(results, len(results))
			if len(results) < 3 {
				return backoff.ErrContinue // notify fn below will be elided by NotifyContinue
			}
			return errors.New("done")
		}, &backoff.ZeroBackOff{}, backoff.NotifyContinue(backoff.Notify(func(err error, d time.Duration) error {
			results = append(results, -1)
			return errors.New("done")
		})))
		require.Equal(t, []int{0, 1, 2, -1}, results)
	})

	t.Run("NotifyContinueWithFunction", func(t *testing.T) {
		var results []int
		backoff.RetryNotify(func() error { //nolint:errcheck
			results = append(results, len(results))
			if len(results) < 3 {
				return backoff.ErrContinue // notify fn below will be elided by NotifyContinue
			}
			return errors.New("done")
		}, &backoff.ZeroBackOff{}, backoff.NotifyContinue(func(err error, d time.Duration) error {
			results = append(results, -1)
			return errors.New("done")
		}))
		require.Equal(t, []int{0, 1, 2, -1}, results)
	})

	t.Run("NotifyContinueWithString", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stdout)
		var results []int
		ctx, cancel := context.WithCancel(context.Background())
		backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
			results = append(results, len(results))
			if len(results) < 3 {
				return backoff.ErrContinue // causes NotifyContinue to re-run
			} else if len(results) < 4 {
				return errors.New("done") // causes NotifyContinue to log & re-run
			}
			cancel() // actually breaks out
			return nil
		}, &backoff.ZeroBackOff{}, backoff.NotifyContinue(t.Name()))
		require.Equal(t, []int{0, 1, 2, 3}, results)
		require.Equal(t, 1, strings.Count(buf.String(), t.Name())) // logged once
	})
}

func TestMustLoop(t *testing.T) {
	var results []int
	ctx, cancel := context.WithCancel(context.Background())
	backoff.RetryUntilCancel(ctx, backoff.MustLoop(func() error { //nolint:errcheck
		results = append(results, len(results)+1)
		switch len(results) {
		case 1:
			return backoff.ErrContinue
		case 2:
			break // return nil, but don't cancel context
		case 3:
			cancel() // actually break out
		}
		return nil
	}), &backoff.ZeroBackOff{}, backoff.NotifyContinue(t.Name()))
	require.Equal(t, []int{1, 2, 3}, results)
}

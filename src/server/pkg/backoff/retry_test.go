package backoff

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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

	err := Retry(f, NewExponentialBackOff())
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
	RetryUntilCancel(ctx, func() error {
		results = append(results, len(results))
		if len(results) >= 3 {
			cancel()
		}
		return Loop
	}, &ZeroBackOff{}, NotifyLoop(func(err error, d time.Duration) error {
		panic("this should not be called due to cancellation")
	}))
	require.Equal(t, []int{0, 1, 2}, results)
}

func TestRetryUntilCancelZeroBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	first := true
	RetryUntilCancel(ctx, func() error {
		if first {
			first = false
		} else {
			t.Fatal("operation() should only run once")
		}
		return nil
	}, &ZeroBackOff{}, func(err error, d time.Duration) error {
		cancel() // This should prevent operation() from running
		return nil
	})
}

func TestNotifyLoop(t *testing.T) {
	t.Run("NotifyLoopWithNotify", func(t *testing.T) {
		var results []int
		RetryNotify(func() error {
			results = append(results, len(results))
			if len(results) < 3 {
				return Loop // notify fn below will be elided by NotifyLoop
			}
			return errors.New("done")
		}, &ZeroBackOff{}, NotifyLoop(Notify(func(err error, d time.Duration) error {
			results = append(results, -1)
			return errors.New("done")
		})))
		require.Equal(t, []int{0, 1, 2, -1}, results)
	})

	t.Run("NotifyLoopWithFunction", func(t *testing.T) {
		var results []int
		RetryNotify(func() error {
			results = append(results, len(results))
			if len(results) < 3 {
				return Loop // notify fn below will be elided by NotifyLoop
			}
			return errors.New("done")
		}, &ZeroBackOff{}, NotifyLoop(func(err error, d time.Duration) error {
			results = append(results, -1)
			return errors.New("done")
		}))
		require.Equal(t, []int{0, 1, 2, -1}, results)
	})

	t.Run("NotifyLoopWithString", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stdout)
		var results []int
		ctx, cancel := context.WithCancel(context.Background())
		RetryUntilCancel(ctx, func() error {
			results = append(results, len(results))
			if len(results) < 3 {
				return Loop // causes NotifyLoop to re-run
			} else if len(results) < 4 {
				return errors.New("done") // causes NotifyLoop to log & re-run
			}
			cancel() // actually breaks out
			return nil
		}, &ZeroBackOff{}, NotifyLoop(t.Name()))
		require.Equal(t, []int{0, 1, 2, 3}, results)
		require.Equal(t, 1, strings.Count(buf.String(), t.Name())) // logged once
	})
}

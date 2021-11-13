package backoff_test

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func TestNextBackOffMillis(t *testing.T) {
	subtestNextBackOff(t, 0, new(backoff.ZeroBackOff))
	subtestNextBackOff(t, backoff.Stop, new(backoff.StopBackOff))
}

func subtestNextBackOff(t *testing.T, expectedValue time.Duration, backOffPolicy backoff.BackOff) {
	for i := 0; i < 10; i++ {
		next := backOffPolicy.NextBackOff()
		if next != expectedValue {
			t.Errorf("got: %d expected: %d", next, expectedValue)
		}
	}
}

func TestConstantBackOff(t *testing.T) {
	backoff := backoff.NewConstantBackOff(time.Second)
	if backoff.NextBackOff() != time.Second {
		t.Error("invalid interval")
	}
}

func abstime(t time.Duration) time.Duration {
	if t < 0 {
		return -t
	}
	return t
}

func TestConstantBackOffCompare(t *testing.T) {
	var callTimes [10]time.Time
	idx := 0
	start := time.Now()
	err := backoff.Retry(func() error {
		callTimes[idx] = time.Now()
		idx++
		return errors.Errorf("expected error")
	}, backoff.RetryEvery(time.Second).For(9*time.Second))
	if err.Error() != "expected error" {
		t.Fatalf("Retry loop didn't return internal error to caller")
	}

	epsilon := 500 * time.Millisecond
	if idx < 8 {
		t.Fatalf("expected 9 retries, but only saw %d", idx)
	}
	nextT := start
	for i := 0; i < idx; i++ {
		if abstime(callTimes[i].Sub(nextT)) > epsilon {
			t.Fatalf("expected retry %d to occur at %s but actually occurred at %s",
				i, nextT.String(), callTimes[i].String())
		}
		nextT = nextT.Add(1 * time.Second)
	}
}

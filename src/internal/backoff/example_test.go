package backoff_test

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
)

func ExampleRetry() {
	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		// Handle error.
		return
	}

	// Operation is successful.
}

func ExampleTicker() {
	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	ticker := backoff.NewTicker(backoff.NewExponentialBackOff())

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = operation(); err != nil {
			fmt.Println(err, "will retry...")
			continue
		}

		ticker.Stop()
		break
	}

	if err != nil {
		// Operation has failed.
		return
	}
}

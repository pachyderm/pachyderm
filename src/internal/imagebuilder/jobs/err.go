package jobs

import (
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Retryable is an error that indicates an operation is retryable.
type Retryable struct{}

func (*Retryable) Is(target error) bool {
	_, ok := target.(*Retryable)
	return ok
}
func (*Retryable) Error() string { return "retryable" }

func WrapRetryable(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(err, &Retryable{})
}

// CheckHTTPStatus checks that the response code in got equals want.  If not, an error is returned.
// The error is marked retryable if the status is >= 500.
func CheckHTTPStatus(got *http.Response, want int) error {
	if g := got.StatusCode; g != want {
		err := errors.Errorf("got status %v (%v), want status %v", g, got.Status, want)
		if g >= http.StatusInternalServerError {
			err = WrapRetryable(err)
		}
		return err
	}
	return nil
}

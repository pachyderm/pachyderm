package jobs

import "errors"

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

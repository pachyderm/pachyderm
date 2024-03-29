package cmdutil

import (
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// TimeFlag implements the pflag.Value interface, enabling RFC3339 (with
// optional nanoseconds) flag arguments.
type TimeFlag time.Time

// String implements the pflag.Value interface.  It returns the value of the
// timestamp in RFC3339 (with optional nanoseconds) format.
func (value *TimeFlag) String() string {
	return time.Time(*value).Format(time.RFC3339Nano)
}

// Set implements the pflag.Value interface.  It sets the value of the flag to
// its parsed argument, if it is acceptable RFC 3339 (with optional nanoseconds) format.
func (value *TimeFlag) Set(s string) error {
	ts, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return errors.Wrapf(err, "invalid RFC3339 date: %s", s)
	}
	*value = TimeFlag(ts)
	return nil
}

// Type implements the pflag.Value interface.  It simply returns a string
// describing the format.
func (value *TimeFlag) Type() string {
	return "RF3339 date (with optional nanoseconds)"
}

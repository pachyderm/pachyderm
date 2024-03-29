package cmdutil

import (
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type TimeFlag time.Time

func (value *TimeFlag) String() string {
	return time.Time(*value).Format(time.RFC3339Nano)
}

func (value *TimeFlag) Set(s string) error {
	ts, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return errors.Wrapf(err, "invalid RFC3339 date: %s", s)
	}
	*value = TimeFlag(ts)
	return nil
}

func (value *TimeFlag) Type() string {
	return "RF3339 date (with nanoseconds)"
}

func (value *TimeFlag) AsTime() time.Time {
	return time.Time(*value)
}

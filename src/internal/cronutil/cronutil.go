// Package cronutil provides an implementation of the cron.Schedule interface.
package cronutil

import (
	"strings"
	"time"

	"github.com/robfig/cron"
)

const (
	// epoch seconds for UTC 9999-01-01T00:00:00
	epoch_9999_01_01 = 253_370_764_800
)

// Year9999 is an empty struct used as a receiver for implementing the
// cron.Schedule interface.
type Year9999 struct{}

// Next implements the cron.Schedule interface.  It returns timepoint for year
// 9999.
func (Year9999) Next(_ time.Time) time.Time {
	return time.Unix(epoch_9999_01_01, 0)
}

// ParseCronExpression is a simple wrapper around cronParseStandard which
// supports the following cron syntax enhancements:
//
//   - @never: Returns year 9999 timepoint, a 4-digit year far enough in the
//     future to act as never.
func ParseCronExpression(cronExpr string) (cron.Schedule, error) {
	if strings.TrimSpace(cronExpr) == "@never" {
		return Year9999{}, nil
	}
	return cron.ParseStandard(cronExpr) //nolint:wrapcheck
}

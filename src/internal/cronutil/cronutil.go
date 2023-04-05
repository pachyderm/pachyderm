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

// empty struct for implementing cron.Schedule interface
type Year9999 struct{}

// implements Schedule.Next(time.Time) time.Time
// returns timepoint for year 9999
func (Year9999) Next(_ time.Time) time.Time {
	return time.Unix(epoch_9999_01_01, 0)
}

// simple wrapper around robfig/cron ParseStandard(standardSpec string) (Schedule, error)
// supports the following cron syntax enhancements
// @never: returns year 9999 timepoint, a 4-digit year far enough in the future to act as never
func ParseCronExpression(cronExpr string) (cron.Schedule, error) {
	if strings.TrimSpace(cronExpr) == "@never" {
		return Year9999{}, nil
	}
	return cron.ParseStandard(cronExpr) //nolint:wrapcheck
}

package pretty

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/fatih/color"
	"github.com/gogo/protobuf/types"
)

// UnescapeHTML returns s with < and > unescaped.
func UnescapeHTML(s string) string {
	s = strings.Replace(s, "\\u003c", "<", -1)
	s = strings.Replace(s, "\\u003e", ">", -1)
	return s
}

// Since pretty-prints the amount of time that has passed since timestamp as a
// human-readable string.
func Since(timestamp *types.Timestamp) string {
	t, _ := types.TimestampFromProto(timestamp)
	if t.Equal(time.Time{}) {
		return ""
	}
	return units.HumanDuration(time.Since(t))
}

// Ago pretty-prints the amount of time that has passed since timestamp as a
// human-readable string, and adds "ago" to the end.
func Ago(timestamp *types.Timestamp) string {
	since := Since(timestamp)
	if since == "" {
		return since
	}
	return fmt.Sprintf("%s ago", since)
}

// TimeDifference pretty-prints the duration of time between from
// and to as a human-reabable string.
func TimeDifference(from *types.Timestamp, to *types.Timestamp) string {
	tFrom, _ := types.TimestampFromProto(from)
	tTo, _ := types.TimestampFromProto(to)
	return units.HumanDuration(tTo.Sub(tFrom))
}

// Duration pretty prints a duration in a human readable way.
func Duration(d *types.Duration) string {
	duration, _ := types.DurationFromProto(d)
	return units.HumanDuration(duration)
}

// Size pretty-prints size amount of bytes as a human readable string.
func Size(size int64) string {
	return units.BytesSize(float64(size))
}

// ProgressBar pretty prints a progress bar with given width and green, yellow
// and red segments.  green, yellow and red need not add to width, they will be
// normalized. If red is nonzero there will always be at least one red segment,
// even if red is less than 1/width of the total bar.
func ProgressBar(width, green, yellow, red int) string {
	total := green + yellow + red
	var sb strings.Builder
	for i := 0; i < width; i++ {
		switch {
		case i == width-1 && red != 0:
			// if there is nonzero red then the final segment is always red,
			// this ensures that we don't present something as totally
			// successful when it wasn't
			sb.WriteString(color.RedString("▇"))
		case i*total < green*width:
			sb.WriteString(color.GreenString("▇"))
		case i*total < (green+yellow)*width:
			sb.WriteString(color.YellowString("▇"))
		case i*total < (green+yellow+red)*width:
			sb.WriteString(color.RedString("▇"))
		default:
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

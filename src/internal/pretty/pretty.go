// Package pretty provides pretty-printing.
package pretty

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/fatih/color"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UnescapeHTML returns s with < and > unescaped.
func UnescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "\\u003c", "<")
	s = strings.ReplaceAll(s, "\\u003e", ">")
	return s
}

// Since pretty-prints the amount of time that has passed since timestamp as a
// human-readable string.
func Since(timestamp *timestamppb.Timestamp) string {
	t := timestamp.AsTime()
	if t.Equal(time.Time{}) {
		return ""
	}
	return units.HumanDuration(time.Since(t))
}

// Ago pretty-prints the amount of time that has passed since timestamp as a
// human-readable string, and adds "ago" to the end.
func Ago(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		return "-"
	}
	since := Since(timestamp)
	if since == "" {
		return since
	}
	return fmt.Sprintf("%s ago", since)
}

// Timestamp pretty-prints a timestamp.
func Timestamp(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		return "-"
	}
	return timestamp.AsTime().Format(time.RFC3339)
}

// TimeDifference pretty-prints the duration of time between from
// and to as a human-reabable string.
func TimeDifference(from, to *timestamppb.Timestamp) string {
	return units.HumanDuration(to.AsTime().Sub(from.AsTime()))
}

// Duration pretty prints a duration in a human readable way.
func Duration(d *durationpb.Duration) string {
	return units.HumanDuration(d.AsDuration())
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

func Commafy(items interface{}) string {
	v := reflect.ValueOf(items)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		switch v.Len() {
		case 0:
			return ""
		case 1:
			return fmt.Sprintf("%v", v.Index(0).Interface())
		default:
			s := fmt.Sprintf("%v", v.Index(0).Interface())
			for i := 1; i < v.Len()-1; i++ {
				s = fmt.Sprintf("%s, %v", s, v.Index(i).Interface())
			}
			return fmt.Sprintf("%s and %v", s, v.Index(v.Len()-1))
		}
	default:
		return fmt.Sprintf("%v", items)
	}
}

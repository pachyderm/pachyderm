package pretty

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/go-units"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
)

// UnescapeHTML returns s with < and > unescaped.
func UnescapeHTML(s string) string {
	s = strings.Replace(s, "\\u003c", "<", -1)
	s = strings.Replace(s, "\\u003e", ">", -1)
	return s
}

// Ago pretty-prints the amount of time that has passed
// since timestamp as a human-readable string.
func Ago(timestamp *google_protobuf.Timestamp) string {
	return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(prototime.TimestampToTime(timestamp))))
}

// Duration pretty-prints the duration of time between from
// and to as a human-reabable string.
func Duration(from *google_protobuf.Timestamp, to *google_protobuf.Timestamp) string {
	return units.HumanDuration(prototime.TimestampToTime(to).Sub(prototime.TimestampToTime(from)))
}

// Size pretty-prints size amount of bytes as a human readable string.
func Size(size uint64) string {
	return units.BytesSize(float64(size))
}

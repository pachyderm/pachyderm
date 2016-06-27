package pretty

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/go-units"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
)

func UnescapeHTML(s string) string {
	s = strings.Replace(s, "\\u003c", "<", -1)
	s = strings.Replace(s, "\\u003e", ">", -1)
	return s
}

func Ago(timestamp *google_protobuf.Timestamp) string {
	return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(prototime.TimestampToTime(timestamp))))
}

func Duration(from *google_protobuf.Timestamp, to *google_protobuf.Timestamp) string {
	return units.HumanDuration(prototime.TimestampToTime(to).Sub(prototime.TimestampToTime(from)))
}

func Size(size uint64) string {
	return units.BytesSize(float64(size))
}

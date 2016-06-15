package pretty

import (
	"fmt"
	"time"

	"github.com/docker/go-units"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
)

func PrettyDuration(timestamp *google_protobuf.Timestamp) string {
	return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(prototime.TimestampToTime(timestamp))))
}

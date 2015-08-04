package store

import (
	"time"

	"github.com/dancannon/gorethink"
	"github.com/peter-edge/go-google-protobuf"
)

var (
	defaultTimer = &systemTimer{}

	tables = []string{
		"pipeline_runs",
		"pipeline_run_statuses",
		"pipeline_run_container_ids",
	}
)

func initTables(session *gorethink.Session, database string) error {
	for _, table := range tables {
		if err := gorethink.DB(database).TableCreate(table).Exec(session); err != nil {
			return err
		}
	}
	return nil
}

func timeToTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: t.UnixNano() / int64(time.Second),
		Nanos:   int32(t.UnixNano() % int64(time.Second)),
	}
}

type timer interface {
	Now() time.Time
}

type systemTimer struct{}

func (t *systemTimer) Now() time.Time {
	return time.Now().UTC()
}

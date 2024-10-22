// Package pretty implements pretty-printing for snapshot
package pretty

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"

	"github.com/pachyderm/pachyderm/v2/src/snapshot"
)

const SnapshotHeader = "ID\tCHUNKSET\tFILESET\tCREATED\t\n"

func PrintSnapshotInfo(w io.Writer, info *snapshot.SnapshotInfo) {
	fmt.Fprintf(w, "%d\t%d\t%v\t%v\n", info.GetId(), info.GetChunksetId(), info.GetSqlDumpFilesetId(), pretty.Ago(info.GetCreatedAt()))
}

// Package pretty implements pretty-printing for snapshot
package pretty

import (
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/snapshot"
)

const SnapshotHeader = "ID\tChunksetId\tCREATED\t\n"

func PrintSnapshotInfo(w io.Writer, info *snapshot.SnapshotInfo) {
	var line []string
	// ID
	line = append(line, fmt.Sprintf("%d", info.Id))
	// chunkset id
	line = append(line, fmt.Sprintf("%d", info.ChunksetId))
	// Created At
	line = append(line, pretty.Ago(info.CreatedAt))

	fmt.Fprintln(w, strings.Join(line, "\t"))
}

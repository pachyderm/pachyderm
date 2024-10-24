// Package pretty implements pretty-printing for snapshot
package pretty

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"github.com/pachyderm/pachyderm/v2/src/snapshot"
)

const SnapshotHeader = "ID\tCHUNKSET\tFILESET\tCREATED\t\n"

func PrintSnapshotInfo(w io.Writer, info *snapshot.SnapshotInfo) {
	fmt.Fprintf(w, "%d\t%d\t%v\t%v\n", info.GetId(), info.GetChunksetId(), info.GetSqlDumpFilesetId(), pretty.Ago(info.GetCreatedAt()))
}

func PrintDetailedSnapshotInfo(info *snapshot.SnapshotInfo) error {
	t, err := template.New("SnapshotInfo").Funcs(funcMap).Parse(
		`ID: {{.Id}}
Chunkset: {{.ChunksetId}}
Created: {{prettyAgo .Created}}{{if .SqlDumpFilesetId}}
SqlDumpFileset: {{.SqlDumpFilesetId}}{{end}}{{if .PachydermVersion}}
Version: {{.PachydermVersion}}{{end}}
`)
	if err != nil {
		return errors.Wrap(err, "parse template")
	}
	return errors.Wrap(t.Execute(os.Stdout, info), "execute template")
}

var funcMap = template.FuncMap{
	"prettyAgo":  pretty.Ago,
	"prettySize": pretty.Size,
	"prettyTime": pretty.Timestamp,
}

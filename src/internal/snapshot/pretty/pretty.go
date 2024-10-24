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

const SnapshotHeader = "ID\tCHUNKSET\tCREATED\t\n"

func PrintSnapshotInfo(w io.Writer, info *snapshot.SnapshotInfo) {
	fmt.Fprintf(w, "%v\t%v\t%v\n", info.GetId(), info.GetChunksetId(), pretty.Ago(info.GetCreatedAt()))
}

func PrintDetailedSnapshotInfo(resp *snapshot.InspectSnapshotResponse) error {
	t, err := template.New("SnapshotInfo").Funcs(funcMap).Parse(
		`ID: {{.Info.Id}}
Chunkset: {{.Info.ChunksetId}}
Created: {{prettyAgo .Info.CreatedAt}}{{if .Info.PachydermVersion}}
Version: {{.Info.PachydermVersion}}{{end}}{{if .Fileset}}
Fileset: {{.Fileset}}{{end}}
`)
	if err != nil {
		return errors.Wrap(err, "parse template")
	}
	return errors.Wrap(t.Execute(os.Stdout, resp), "execute template")
}

var funcMap = template.FuncMap{
	"prettyAgo":  pretty.Ago,
	"prettySize": pretty.Size,
	"prettyTime": pretty.Timestamp,
}

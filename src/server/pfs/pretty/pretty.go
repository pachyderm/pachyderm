package pretty

import (
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
)

func PrintRepoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tCREATED\tSIZE\t\n")
}

func PrintRepoInfo(w io.Writer, repoInfo *pfs.RepoInfo) {
	fmt.Fprintf(w, "%s\t", repoInfo.Repo.Name)
	fmt.Fprintf(
		w,
		"%s\t",
		pretty.PrettyDuration(repoInfo.Created),
	)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(repoInfo.SizeBytes)))
}

func PrintDetailedRepoInfo(repoInfo *pfs.RepoInfo) {
	template, err := template.New("RepoInfo").Funcs(funcMap).Parse(
		`Name: {{.Repo.Name}}
Created: {{prettyDuration .Created}}
Size: {{prettySize .SizeBytes}}{{if .Provenance}}
Provenance: {{range .Provenance}} {{.Name}} {{end}} {{end}}
`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = template.Execute(os.Stdout, repoInfo)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func PrintCommitInfoHeader(w io.Writer) {
	fmt.Fprint(w, "BRANCH\tID\tPARENT\tSTARTED\tFINISHED\tSIZE\t\n")
}

func PrintCommitInfo(w io.Writer, commitInfo *pfs.CommitInfo) {
	fmt.Fprintf(w, "%s\t", commitInfo.Branch)
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.ID)
	if commitInfo.ParentCommit != nil {
		fmt.Fprintf(w, "%s\t", commitInfo.ParentCommit.ID)
	} else {
		fmt.Fprint(w, "<none>\t")
	}
	fmt.Fprintf(
		w,
		"%s\t",
		pretty.PrettyDuration(commitInfo.Started),
	)
	finished := "\t"
	if commitInfo.Finished != nil {
		finished = fmt.Sprintf("%s\t", pretty.PrettyDuration(commitInfo.Finished))
	}
	fmt.Fprintf(w, finished)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(commitInfo.SizeBytes)))
}

func PrintFileInfoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tTYPE\tMODIFIED\tLAST_COMMIT_MODIFIED\tSIZE\t\n")
}

func PrintFileInfo(w io.Writer, fileInfo *pfs.FileInfo) {
	fmt.Fprintf(w, "%s\t", fileInfo.File.Path)
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		fmt.Fprint(w, "file\t")
	} else {
		fmt.Fprint(w, "dir\t")
	}
	fmt.Fprintf(
		w,
		"%s\t",
		pretty.PrettyDuration(fileInfo.Modified),
	)
	fmt.Fprint(w, "-\t")
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(fileInfo.SizeBytes)))
}

func PrintBlockInfoHeader(w io.Writer) {
	fmt.Fprintf(w, "HASH\tCREATED\tSIZE\t\n")
}

func PrintBlockInfo(w io.Writer, blockInfo *pfs.BlockInfo) {
	fmt.Fprintf(w, "%s\t", blockInfo.Block.Hash)
	fmt.Fprintf(
		w,
		"%s\t",
		pretty.PrettyDuration(blockInfo.Created),
	)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(blockInfo.SizeBytes)))
}

type uint64Slice []uint64

func (s uint64Slice) Len() int           { return len(s) }
func (s uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s uint64Slice) Less(i, j int) bool { return s[i] < s[j] }

var funcMap = template.FuncMap{
	"prettyDuration": pretty.PrettyDuration,
	"prettySize":     pretty.PrettySize,
}

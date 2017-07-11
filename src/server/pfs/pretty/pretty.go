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

// PrintRepoHeader prints a repo header.
func PrintRepoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tCREATED\tSIZE\t\n")
}

// PrintRepoInfo pretty-prints repo info.
func PrintRepoInfo(w io.Writer, repoInfo *pfs.RepoInfo) {
	fmt.Fprintf(w, "%s\t", repoInfo.Repo.Name)
	fmt.Fprintf(
		w,
		"%s\t",
		pretty.Ago(repoInfo.Created),
	)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(repoInfo.SizeBytes)))
}

// PrintDetailedRepoInfo pretty-prints detailed repo info.
func PrintDetailedRepoInfo(repoInfo *pfs.RepoInfo) error {
	template, err := template.New("RepoInfo").Funcs(funcMap).Parse(
		`Name: {{.Repo.Name}}{{if .Description}}
Description: {{.Description}}{{end}}
Created: {{prettyAgo .Created}}
Size: {{prettySize .SizeBytes}}{{if .Provenance}}
Provenance: {{range .Provenance}} {{.Name}} {{end}} {{end}}
`)
	if err != nil {
		return err
	}
	err = template.Execute(os.Stdout, repoInfo)
	if err != nil {
		return err
	}
	return nil
}

// PrintBranchHeader prints a branch header.
func PrintBranchHeader(w io.Writer) {
	fmt.Fprint(w, "BRANCH\tHEAD\t\n")
}

// PrintBranch pretty-prints a Branch.
func PrintBranch(w io.Writer, branch *pfs.BranchInfo) {
	fmt.Fprintf(w, "%s\t", branch.Name)
	fmt.Fprintf(w, "%s\t\n", branch.Head.ID)
}

// PrintCommitInfoHeader prints a commit info header.
func PrintCommitInfoHeader(w io.Writer) {
	fmt.Fprint(w, "REPO\tID\tPARENT\tSTARTED\tDURATION\tSIZE\t\n")
}

// PrintCommitInfo pretty-prints commit info.
func PrintCommitInfo(w io.Writer, commitInfo *pfs.CommitInfo) {
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.Repo.Name)
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.ID)
	if commitInfo.ParentCommit != nil {
		fmt.Fprintf(w, "%s\t", commitInfo.ParentCommit.ID)
	} else {
		fmt.Fprint(w, "<none>\t")
	}
	fmt.Fprintf(
		w,
		"%s\t",
		pretty.Ago(commitInfo.Started),
	)
	if commitInfo.Finished != nil {
		fmt.Fprintf(w, fmt.Sprintf("%s\t", pretty.TimeDifference(commitInfo.Started, commitInfo.Finished)))
		fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(commitInfo.SizeBytes)))
	} else {
		fmt.Fprintf(w, "-\t")
		// Open commits don't have meaningful size information
		fmt.Fprintf(w, "-\t\n")
	}
}

// PrintDetailedCommitInfo pretty-prints detailed commit info.
func PrintDetailedCommitInfo(commitInfo *pfs.CommitInfo) error {
	template, err := template.New("CommitInfo").Funcs(funcMap).Parse(
		`Commit: {{.Commit.Repo.Name}}/{{.Commit.ID}}{{if .ParentCommit}}
Parent: {{.ParentCommit.ID}} {{end}}
Started: {{prettyAgo .Started}}{{if .Finished}}
Finished: {{prettyAgo .Finished}} {{end}}
Size: {{prettySize .SizeBytes}}{{if .Provenance}}
Provenance: {{range .Provenance}} {{.Repo.Name}}/{{.ID}} {{end}} {{end}}
`)
	if err != nil {
		return err
	}
	err = template.Execute(os.Stdout, commitInfo)
	if err != nil {
		return err
	}
	return nil
}

// PrintFileInfoHeader prints a file info header.
func PrintFileInfoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tTYPE\tSIZE\t\n")
}

// PrintFileInfo pretty-prints file info.
// If recurse is false and directory size is 0, display "-" instead
// If fast is true and file size is 0, display "-" instead
func PrintFileInfo(w io.Writer, fileInfo *pfs.FileInfo) {
	fmt.Fprintf(w, "%s\t", fileInfo.File.Path)
	if fileInfo.FileType == pfs.FileType_FILE {
		fmt.Fprint(w, "file\t")
	} else {
		fmt.Fprint(w, "dir\t")
	}
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(fileInfo.SizeBytes)))
}

// PrintDetailedFileInfo pretty-prints detailed file info.
func PrintDetailedFileInfo(fileInfo *pfs.FileInfo) error {
	template, err := template.New("FileInfo").Funcs(funcMap).Parse(
		`Path: {{.File.Path}}
Type: {{fileType .FileType}}
Size: {{prettySize .SizeBytes}}
Children: {{range .Children}} {{.}} {{end}}
`)
	if err != nil {
		return err
	}
	if err := template.Execute(os.Stdout, fileInfo); err != nil {
		return err
	}
	return nil
}

type uint64Slice []uint64

func (s uint64Slice) Len() int           { return len(s) }
func (s uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s uint64Slice) Less(i, j int) bool { return s[i] < s[j] }

func fileType(fileType pfs.FileType) string {
	if fileType == pfs.FileType_FILE {
		return "file"
	}
	return "dir"
}

var funcMap = template.FuncMap{
	"prettyAgo":  pretty.Ago,
	"prettySize": pretty.Size,
	"fileType":   fileType,
}

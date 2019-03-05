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

const (
	// RepoHeader is the header for repos.
	RepoHeader = "NAME\tCREATED\tSIZE (MASTER)\t\n"
	// RepoAuthHeader is the header for repos with auth information attached.
	RepoAuthHeader = "NAME\tCREATED\tSIZE (MASTER)\tACCESS LEVEL\t\n"
	// CommitHeader is the header for commits.
	CommitHeader = "REPO\tCOMMIT\tPARENT\tSTARTED\tDURATION\tSIZE\t\n"
	// BranchHeader is the header for branches.
	BranchHeader = "BRANCH\tHEAD\t\n"
	// FileHeader is the header for files.
	FileHeader = "COMMIT\tNAME\tTYPE\tCOMMITTED\tSIZE\t\n"
)

// PrintRepoHeader prints a repo header.
func PrintRepoHeader(w io.Writer, printAuth bool) {
	if printAuth {
		fmt.Fprint(w, RepoAuthHeader)
		return
	}
	fmt.Fprint(w, RepoHeader)
}

// PrintRepoInfo pretty-prints repo info.
func PrintRepoInfo(w io.Writer, repoInfo *pfs.RepoInfo, fullTimestamp bool) {
	fmt.Fprintf(w, "%s\t", repoInfo.Repo.Name)
	if fullTimestamp {
		fmt.Fprintf(w, "%s\t", repoInfo.Created.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(repoInfo.Created))
	}
	fmt.Fprintf(w, "%s\t", units.BytesSize(float64(repoInfo.SizeBytes)))
	if repoInfo.AuthInfo != nil {
		fmt.Fprintf(w, "%s\t", repoInfo.AuthInfo.AccessLevel.String())
	}
	fmt.Fprintln(w)
}

// PrintableRepoInfo is a wrapper around RepoInfo containing any formatting options
// used within the template to conditionally print information.
type PrintableRepoInfo struct {
	*pfs.RepoInfo
	FullTimestamp bool
}

// PrintDetailedRepoInfo pretty-prints detailed repo info.
func PrintDetailedRepoInfo(repoInfo *PrintableRepoInfo) error {
	template, err := template.New("RepoInfo").Funcs(funcMap).Parse(
		`Name: {{.Repo.Name}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .FullTimestamp}}
Created: {{.Created}}{{else}}
Created: {{prettyAgo .Created}}{{end}}
Size of HEAD on master: {{prettySize .SizeBytes}}{{if .AuthInfo}}
Access level: {{ .AuthInfo.AccessLevel.String }}{{end}}
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
	fmt.Fprint(w, BranchHeader)
}

// PrintBranch pretty-prints a Branch.
func PrintBranch(w io.Writer, branchInfo *pfs.BranchInfo) {
	fmt.Fprintf(w, "%s\t", branchInfo.Branch.Name)
	if branchInfo.Head != nil {
		fmt.Fprintf(w, "%s\t\n", branchInfo.Head.ID)
	} else {
		fmt.Fprintf(w, "-\t\n")
	}
}

// PrintCommitInfoHeader prints a commit info header.
func PrintCommitInfoHeader(w io.Writer) {
	fmt.Fprint(w, CommitHeader)
}

// PrintCommitInfo pretty-prints commit info.
func PrintCommitInfo(w io.Writer, commitInfo *pfs.CommitInfo, fullTimestamp bool) {
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.Repo.Name)
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.ID)
	if commitInfo.ParentCommit != nil {
		fmt.Fprintf(w, "%s\t", commitInfo.ParentCommit.ID)
	} else {
		fmt.Fprint(w, "<none>\t")
	}
	if fullTimestamp {
		fmt.Fprintf(w, "%s\t", commitInfo.Started.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(commitInfo.Started))
	}
	if commitInfo.Finished != nil {
		fmt.Fprintf(w, fmt.Sprintf("%s\t", pretty.TimeDifference(commitInfo.Started, commitInfo.Finished)))
		fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(commitInfo.SizeBytes)))
	} else {
		fmt.Fprintf(w, "-\t")
		// Open commits don't have meaningful size information
		fmt.Fprintf(w, "-\t\n")
	}
}

// PrintableCommitInfo is a wrapper around CommitInfo containing any formatting options
// used within the template to conditionally print information.
type PrintableCommitInfo struct {
	*pfs.CommitInfo
	FullTimestamp bool
}

// PrintDetailedCommitInfo pretty-prints detailed commit info.
func PrintDetailedCommitInfo(commitInfo *PrintableCommitInfo) error {
	template, err := template.New("CommitInfo").Funcs(funcMap).Parse(
		`Commit: {{.Commit.Repo.Name}}/{{.Commit.ID}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .ParentCommit}}
Parent: {{.ParentCommit.ID}}{{end}}{{if .FullTimestamp}}
Started: {{.Started}}{{else}}
Started: {{prettyAgo .Started}}{{end}}{{if .Finished}}{{if .FullTimestamp}}
Finished: {{.Finished}}{{else}}
Finished: {{prettyAgo .Finished}}{{end}}{{end}}
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
	fmt.Fprint(w, FileHeader)
}

// PrintFileInfo pretty-prints file info.
// If recurse is false and directory size is 0, display "-" instead
// If fast is true and file size is 0, display "-" instead
func PrintFileInfo(w io.Writer, fileInfo *pfs.FileInfo, fullTimestamp bool) {
	fmt.Fprintf(w, "%s\t", fileInfo.File.Commit.ID)
	fmt.Fprintf(w, "%s\t", fileInfo.File.Path)
	if fileInfo.FileType == pfs.FileType_FILE {
		fmt.Fprint(w, "file\t")
	} else {
		fmt.Fprint(w, "dir\t")
	}
	if fullTimestamp {
		fmt.Fprintf(w, "%s\t", fileInfo.Committed.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(fileInfo.Committed))
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
	return template.Execute(os.Stdout, fileInfo)
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

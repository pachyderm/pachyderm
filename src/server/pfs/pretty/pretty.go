package pretty

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	units "github.com/docker/go-units"
	"github.com/fatih/color"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	// RepoHeader is the header for repos.
	RepoHeader = "NAME\tCREATED\tSIZE (MASTER)\tDESCRIPTION\t\n"
	// RepoAuthHeader is the header for repos with auth information attached.
	RepoAuthHeader = "NAME\tCREATED\tSIZE (MASTER)\tACCESS LEVEL\t\n"
	// CommitHeader is the header for commits.
	CommitHeader = "REPO\tBRANCH\tCOMMIT\tFINISHED\tSIZE\tPROGRESS\tDESCRIPTION\n"
	// BranchHeader is the header for branches.
	BranchHeader = "BRANCH\tHEAD\tTRIGGER\t\n"
	// FileHeader is the header for files.
	FileHeader = "NAME\tTYPE\tSIZE\t\n"
	// FileHeaderWithCommit is the header for files that includes a commit field.
	FileHeaderWithCommit = "COMMIT\tNAME\tTYPE\tCOMMITTED\tSIZE\t\n"
	// DiffFileHeader is the header for files produced by diff file.
	DiffFileHeader = "OP\t" + FileHeader
)

// PrintRepoInfo pretty-prints repo info.
func PrintRepoInfo(w io.Writer, repoInfo *pfs.RepoInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", CompactPrintRepo(repoInfo.Repo))
	if fullTimestamps {
		fmt.Fprintf(w, "%s\t", repoInfo.Created.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(repoInfo.Created))
	}
	fmt.Fprintf(w, "%s\t", units.BytesSize(float64(repoInfo.SizeBytes)))
	if repoInfo.AuthInfo != nil {
		fmt.Fprintf(w, "%s\t", repoInfo.AuthInfo.Roles)
	}
	fmt.Fprintf(w, "%s\t", repoInfo.Description)
	fmt.Fprintln(w)
}

// PrintableRepoInfo is a wrapper around RepoInfo containing any formatting options
// used within the template to conditionally print information.
type PrintableRepoInfo struct {
	*pfs.RepoInfo
	FullTimestamps bool
}

// NewPrintableRepoInfo constructs a PrintableRepoInfo from just a RepoInfo.
func NewPrintableRepoInfo(ri *pfs.RepoInfo) *PrintableRepoInfo {
	return &PrintableRepoInfo{
		RepoInfo: ri,
	}
}

// PrintDetailedRepoInfo pretty-prints detailed repo info.
func PrintDetailedRepoInfo(repoInfo *PrintableRepoInfo) error {
	template, err := template.New("RepoInfo").Funcs(funcMap).Parse(
		`Name: {{.Repo.Name}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .FullTimestamps}}
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

func printTrigger(trigger *pfs.Trigger) string {
	var conds []string
	if trigger.CronSpec != "" {
		conds = append(conds, fmt.Sprintf("Cron(%s)", trigger.CronSpec))
	}
	if trigger.Size_ != "" {
		conds = append(conds, fmt.Sprintf("Size(%s)", trigger.Size_))
	}
	if trigger.Commits != 0 {
		conds = append(conds, fmt.Sprintf("Commits(%d)", trigger.Commits))
	}
	cond := ""
	if trigger.All {
		cond = strings.Join(conds, " and ")
	} else {
		cond = strings.Join(conds, " or ")
	}
	return fmt.Sprintf("%s on %s", trigger.Branch, cond)
}

// PrintBranch pretty-prints a Branch.
func PrintBranch(w io.Writer, branchInfo *pfs.BranchInfo) {
	fmt.Fprintf(w, "%s\t", branchInfo.Branch.Name)
	if branchInfo.Head != nil {
		fmt.Fprintf(w, "%s\t", branchInfo.Head.ID)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	if branchInfo.Trigger != nil {
		fmt.Fprintf(w, "%s\t", printTrigger(branchInfo.Trigger))
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintln(w)
}

// PrintDetailedBranchInfo pretty-prints detailed branch info.
func PrintDetailedBranchInfo(branchInfo *pfs.BranchInfo) error {
	template, err := template.New("BranchInfo").Funcs(funcMap).Parse(
		`Name: {{.Branch.Repo.Name}}@{{.Branch.Name}}{{if .Head}}
Head Commit: {{ .Head.Branch.Repo.Name}}@{{.Head.ID}} {{end}}{{if .Provenance}}
Provenance: {{range .Provenance}} {{.Repo.Name}}@{{.Name}} {{end}} {{end}}{{if .Trigger}}
Trigger: {{printTrigger .Trigger}} {{end}}
`)
	if err != nil {
		return err
	}
	err = template.Execute(os.Stdout, branchInfo)
	if err != nil {
		return err
	}
	return nil
}

// PrintCommitInfo pretty-prints commit info.
func PrintCommitInfo(w io.Writer, commitInfo *pfs.CommitInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.Branch.Repo.Name)
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.Branch.Name)
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.ID)
	if commitInfo.Finished == nil {
		fmt.Fprintf(w, "-\t")
	} else {
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", commitInfo.Finished.String())
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(commitInfo.Finished))
		}
	}
	if commitInfo.Finished == nil {
		fmt.Fprintf(w, "-\t")
	} else {
		fmt.Fprintf(w, "%s\t", units.BytesSize(float64(commitInfo.SizeBytes)))
	}
	if commitInfo.SubvenantCommitsTotal == 0 {
		fmt.Fprintf(w, "-\t")
	} else {
		fmt.Fprintf(w, "%s\t", pretty.ProgressBar(
			8,
			int(commitInfo.SubvenantCommitsSuccess),
			int(commitInfo.SubvenantCommitsTotal-commitInfo.SubvenantCommitsSuccess-commitInfo.SubvenantCommitsFailure),
			int(commitInfo.SubvenantCommitsFailure)))
	}
	fmt.Fprintf(w, "%s\t", commitInfo.Description)
	fmt.Fprintln(w)
}

// PrintableCommitInfo is a wrapper around CommitInfo containing any formatting options
// used within the template to conditionally print information.
type PrintableCommitInfo struct {
	*pfs.CommitInfo
	FullTimestamps bool
}

// NewPrintableCommitInfo constructs a PrintableCommitInfo from just a CommitInfo.
func NewPrintableCommitInfo(ci *pfs.CommitInfo) *PrintableCommitInfo {
	return &PrintableCommitInfo{
		CommitInfo: ci,
	}
}

// PrintDetailedCommitInfo pretty-prints detailed commit info.
func PrintDetailedCommitInfo(w io.Writer, commitInfo *PrintableCommitInfo) error {
	template, err := template.New("CommitInfo").Funcs(funcMap).Parse(
		`Commit: {{.Commit.Branch.Repo.Name}}@{{.Commit.ID}}
Original Branch: {{.Commit.Branch.Name}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .ParentCommit}}
Parent: {{.ParentCommit.ID}}{{end}}{{if .FullTimestamps}}
Started: {{.Started}}{{else}}
Started: {{prettyAgo .Started}}{{end}}{{if .Finished}}{{if .FullTimestamps}}
Finished: {{.Finished}}{{else}}
Finished: {{prettyAgo .Finished}}{{end}}{{end}}
Size: {{prettySize .SizeBytes}}{{if .Provenance}}
Provenance: {{range .Provenance}} {{.Commit.Branch.Repo.Name}}@{{.Commit.ID}} ({{.Commit.Branch.Name}}) {{end}} {{end}}
`)
	if err != nil {
		return err
	}
	return template.Execute(w, commitInfo)
}

// PrintFileInfo pretty-prints file info.
// If recurse is false and directory size is 0, display "-" instead
// If fast is true and file size is 0, display "-" instead
func PrintFileInfo(w io.Writer, fileInfo *pfs.FileInfo, fullTimestamps, withCommit bool) {
	if withCommit {
		fmt.Fprintf(w, "%s\t", fileInfo.File.Commit.ID)
	}
	fmt.Fprintf(w, "%s\t", fileInfo.File.Path)
	if fileInfo.FileType == pfs.FileType_FILE {
		fmt.Fprint(w, "file\t")
	} else {
		fmt.Fprint(w, "dir\t")
	}
	if withCommit {
		if fileInfo.Committed == nil {
			fmt.Fprintf(w, "-\t")
		} else if fullTimestamps {
			fmt.Fprintf(w, "%s\t", fileInfo.Committed.String())
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(fileInfo.Committed))
		}
	}
	fmt.Fprintf(w, "%s\t", units.BytesSize(float64(fileInfo.SizeBytes)))
	fmt.Fprintln(w)
}

// PrintDiffFileInfo pretty-prints a file info from diff file.
func PrintDiffFileInfo(w io.Writer, added bool, fileInfo *pfs.FileInfo, fullTimestamps bool) {
	if added {
		fmt.Fprint(w, color.GreenString("+\t"))
	} else {
		fmt.Fprint(w, color.RedString("-\t"))
	}
	PrintFileInfo(w, fileInfo, fullTimestamps, false)
}

// PrintDetailedFileInfo pretty-prints detailed file info.
func PrintDetailedFileInfo(fileInfo *pfs.FileInfo) error {
	template, err := template.New("FileInfo").Funcs(funcMap).Parse(
		`Path: {{.File.Path}}
Type: {{fileType .FileType}}
Size: {{prettySize .SizeBytes}}
`)
	if err != nil {
		return err
	}
	return template.Execute(os.Stdout, fileInfo)
}

func fileType(fileType pfs.FileType) string {
	if fileType == pfs.FileType_FILE {
		return "file"
	}
	return "dir"
}

var funcMap = template.FuncMap{
	"prettyAgo":    pretty.Ago,
	"prettySize":   pretty.Size,
	"fileType":     fileType,
	"printTrigger": printTrigger,
}

func CompactPrintRepo(r *pfs.Repo) string {
	if r.Type == pfs.UserRepoType {
		return r.Name
	}
	return fmt.Sprintf("%s.%s", r.Name, r.Type)
}

// CompactPrintBranch renders 'b' as a compact string, e.g.
// "myrepo@master:/my/file"
func CompactPrintBranch(b *pfs.Branch) string {
	return fmt.Sprintf("%s@%s", CompactPrintRepo(b.Repo), b.Name)
}

// CompactPrintCommit renders 'c' as a compact string, e.g.
// "myrepo@123abc:/my/file"
func CompactPrintCommit(c *pfs.Commit) string {
	return fmt.Sprintf("%s@%s", CompactPrintRepo(c.Branch.Repo), c.ID)
}

// CompactPrintCommitSafe is similar to CompactPrintCommit but accepts 'nil'
// arguments
func CompactPrintCommitSafe(c *pfs.Commit) string {
	if c == nil {
		return "nil"
	}
	return CompactPrintCommit(c)
}

// CompactPrintFile renders 'f' as a compact string, e.g.
// "myrepo@master:/my/file"
func CompactPrintFile(f *pfs.File) string {
	return fmt.Sprintf("%s@%s:%s", CompactPrintRepo(f.Commit.Branch.Repo), f.Commit.ID, f.Path)
}

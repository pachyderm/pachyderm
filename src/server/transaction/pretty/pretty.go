package pretty

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
)

const (
	//TransactionHeader is the header for transactions.
	TransactionHeader = "TRANSACTION\tSTARTED\tOPS\t\n"
)

// PrintableTransactionInfo wraps a transaction.TransactionInfo with the
// information needed to format it when printing.
type PrintableTransactionInfo struct {
	*transaction.TransactionInfo
	FullTimestamps bool
}

// PrintTransactionInfo prints a short summary of a transaction to the provided
// device.
func PrintTransactionInfo(w io.Writer, info *transaction.TransactionInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", info.Transaction.ID)
	if fullTimestamps {
		fmt.Fprintf(w, "%s\t", info.Started.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(info.Started))
	}
	fmt.Fprintf(w, "%d\n", len(info.Requests))
}

// PrintDetailedTransactionInfo prints detailed information about a transaction
// to stdout.
func PrintDetailedTransactionInfo(info *PrintableTransactionInfo) error {
	template, err := template.New("TransactionInfo").Funcs(funcMap).Parse(
		`ID: {{.Transaction.ID}}{{if .FullTimestamps}}
Started: {{.Started}}{{else}}
Started: {{prettyAgo .Started}}{{end}}
Requests:
{{transactionRequests .Requests .Responses}}
`)
	if err != nil {
		return err
	}
	return template.Execute(os.Stdout, info)
}

func sprintCreateRepo(request *pfs.CreateRepoRequest) string {
	if request.Update {
		return fmt.Sprintf("update repo %s", request.Repo.Name)
	}
	return fmt.Sprintf("create repo %s", request.Repo.Name)
}

func sprintDeleteRepo(request *pfs.DeleteRepoRequest) string {
	force := ""
	if request.Force {
		force = " --force"
	}
	if request.All {
		return fmt.Sprintf("delete repo --all%s", request.Repo.Name, force)
	}
	return fmt.Sprintf("delete repo %s%s", request.Repo.Name, force)
}

func sprintStartCommit(request *pfs.StartCommitRequest, response *transaction.TransactionResponse) string {
	commit := "unknown"
	if response != nil {
		switch commitResponse := response.Response.(type) {
		case *transaction.TransactionResponse_Commit:
			commit = commitResponse.Commit.ID
		default:
			commit = "ERROR (unknown response type)"
		}
	}
	return fmt.Sprintf("start commit %s@%s (%s)", request.Parent.Repo.Name, request.Branch, commit)
}

func sprintFinishCommit(request *pfs.FinishCommitRequest) string {
	return fmt.Sprintf("finish commit %s@%s", request.Commit.Repo.Name, request.Commit.ID)
}

func sprintDeleteCommit(request *pfs.DeleteCommitRequest) string {
	return fmt.Sprintf("delete commit %s@%s", request.Commit.Repo.Name, request.Commit.ID)
}

func sprintCreateBranch(request *pfs.CreateBranchRequest) string {
	provenance := ""
	for _, p := range request.Provenance {
		provenance = fmt.Sprintf("%s -p %s@%s", provenance, p.Repo.Name, p.Name)
	}

	return fmt.Sprintf("create branch %s@%s", request.Branch.Repo.Name, request.Branch.Name)
}

func sprintDeleteBranch(request *pfs.DeleteBranchRequest) string {
	force := ""
	if request.Force {
		force = " --force"
	}
	return fmt.Sprintf("delete branch %s@%s%s", request.Branch.Repo.Name, request.Branch.Name, force)
}

func sprintCopyFile(request *pfs.CopyFileRequest) string {
	overwrite := ""
	if request.Overwrite {
		overwrite = " --overwrite"
	}
	return fmt.Sprintf(
		"copy file %s@%s:%s %s@%s:%s%s",
		request.Src.Commit.Repo.Name, request.Src.Commit.ID, request.Src.Path,
		request.Dst.Commit.Repo.Name, request.Src.Commit.ID, request.Src.Path,
		overwrite,
	)
}

func sprintDeleteFile(request *pfs.DeleteFileRequest) string {
	return fmt.Sprintf(
		"delete file %s@%s:%s",
		request.File.Commit.Repo.Name, request.File.Commit.ID, request.File.Path,
	)
}

func transactionRequests(
	requests []*transaction.TransactionRequest,
	responses []*transaction.TransactionResponse,
) string {
	if len(requests) == 0 {
		return "  -"
	}

	lines := []string{}
	for i, x := range requests {
		var line string
		switch request := x.Request.(type) {
		case *transaction.TransactionRequest_CreateRepo:
			line = sprintCreateRepo(request.CreateRepo)
		case *transaction.TransactionRequest_DeleteRepo:
			line = sprintDeleteRepo(request.DeleteRepo)
		case *transaction.TransactionRequest_StartCommit:
			if len(responses) > i {
				line = sprintStartCommit(request.StartCommit, responses[i])
			} else {
				line = sprintStartCommit(request.StartCommit, nil)
			}
		case *transaction.TransactionRequest_FinishCommit:
			line = sprintFinishCommit(request.FinishCommit)
		case *transaction.TransactionRequest_DeleteCommit:
			line = sprintDeleteCommit(request.DeleteCommit)
		case *transaction.TransactionRequest_CreateBranch:
			line = sprintCreateBranch(request.CreateBranch)
		case *transaction.TransactionRequest_DeleteBranch:
			line = sprintDeleteBranch(request.DeleteBranch)
		case *transaction.TransactionRequest_CopyFile:
			line = sprintCopyFile(request.CopyFile)
		case *transaction.TransactionRequest_DeleteFile:
			line = sprintDeleteFile(request.DeleteFile)
		default:
			line = "ERROR (unknown request type)"
		}
		lines = append(lines, fmt.Sprintf("  %s", line))
	}

	return strings.Join(lines, "\n")
}

var funcMap = template.FuncMap{
	"prettyAgo":           pretty.Ago,
	"prettySize":          pretty.Size,
	"transactionRequests": transactionRequests,
}

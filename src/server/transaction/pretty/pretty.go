package pretty

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfspretty "github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
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
	fmt.Fprintf(w, "%s\t", info.Transaction.Id)
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
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(template.Execute(os.Stdout, info))
}

func sprintCreateRepo(request *pfs.CreateRepoRequest) string {
	if request.Update {
		return fmt.Sprintf("update repo %s", request.Repo)
	}
	return fmt.Sprintf("create repo %s", request.Repo)
}

func sprintDeleteRepo(request *pfs.DeleteRepoRequest) string {
	force := ""
	if request.Force {
		force = " --force"
	}
	return fmt.Sprintf("delete repo %s %s", request.Repo, force)
}

func sprintStartCommit(request *pfs.StartCommitRequest, response *transaction.TransactionResponse) string {
	var commit string
	if response == nil || response.Commit == nil {
		commit = "ERROR (unknown response type)"
	} else {
		commit = response.Commit.Id
	}
	return fmt.Sprintf("start commit %s (%s)", request.Branch, commit)
}

func sprintFinishCommit(request *pfs.FinishCommitRequest) string {
	return fmt.Sprintf("finish commit %s", pfspretty.CompactPrintCommit(request.Commit))
}

func sprintSquashCommitSet(request *pfs.SquashCommitSetRequest) string {
	return fmt.Sprintf("squash commitset %s", request.CommitSet.Id)
}

func sprintCreateBranch(request *pfs.CreateBranchRequest) string {
	provenance := ""
	for _, p := range request.Provenance {
		provenance = fmt.Sprintf("%s -p %s", provenance, p)
	}

	return fmt.Sprintf("create branch %s", request.Branch)
}

func sprintDeleteBranch(request *pfs.DeleteBranchRequest) string {
	force := ""
	if request.Force {
		force = " --force"
	}
	return fmt.Sprintf("delete branch %s%s", request.Branch, force)
}

func sprintUpdateJobState(request *pps.UpdateJobStateRequest) string {
	state := func() string {
		switch request.State {
		case pps.JobState_JOB_STARTING:
			return "STARTING"
		case pps.JobState_JOB_RUNNING:
			return "RUNNING"
		case pps.JobState_JOB_FAILURE:
			return "FAILURE"
		case pps.JobState_JOB_SUCCESS:
			return "SUCCESS"
		case pps.JobState_JOB_KILLED:
			return "KILLED"
		default:
			return "<unknown state>"
		}
	}()
	return fmt.Sprintf(
		"update job %s -> %s (%s)",
		request.Job.Id, state, request.Reason,
	)
}

func sprintCreatePipeline(request *pps.CreatePipelineRequest) string {
	verb := "create"
	if request.Update {
		verb = "update"
	}
	return fmt.Sprintf("%s pipeline %s", verb, request.Pipeline)
}

func transactionRequests(
	requests []*transaction.TransactionRequest,
	responses []*transaction.TransactionResponse,
) string {
	if len(requests) == 0 {
		return "  -"
	}

	lines := []string{}
	for i, request := range requests {
		var line string
		if request.CreateRepo != nil {
			line = sprintCreateRepo(request.CreateRepo)
		} else if request.DeleteRepo != nil {
			line = sprintDeleteRepo(request.DeleteRepo)
		} else if request.StartCommit != nil {
			if len(responses) > i {
				line = sprintStartCommit(request.StartCommit, responses[i])
			} else {
				line = sprintStartCommit(request.StartCommit, nil)
			}
		} else if request.FinishCommit != nil {
			line = sprintFinishCommit(request.FinishCommit)
		} else if request.SquashCommitSet != nil {
			line = sprintSquashCommitSet(request.SquashCommitSet)
		} else if request.CreateBranch != nil {
			line = sprintCreateBranch(request.CreateBranch)
		} else if request.DeleteBranch != nil {
			line = sprintDeleteBranch(request.DeleteBranch)
		} else if request.UpdateJobState != nil {
			line = sprintUpdateJobState(request.UpdateJobState)
		} else if request.CreatePipeline != nil {
			line = sprintCreatePipeline(request.CreatePipeline)
		} else {
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

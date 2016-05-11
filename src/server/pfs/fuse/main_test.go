package fuse_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/server/pfs/fuse/spec"
	"os/exec"
	"strings"
)

var OpenCommitSyscallSpec *spec.Spec
var ClosedCommitSyscallSpec *spec.Spec
var CancelledCommitSyscallSpec *spec.Spec
var RootSyscallSpec *spec.Spec
var RepoSyscallSpec *spec.Spec

func TestMain(m *testing.M) {
	fmt.Println("========== Running Pre Fuse Test Hooks\n")

	OpenCommitSyscallSpec, _ = spec.New("Open Commit", "spec/syscalls.txt")
	ClosedCommitSyscallSpec, _ = spec.New("Closed Commit", "spec/syscalls.txt")
	CancelledCommitSyscallSpec, _ = spec.New("Cancelled Commit", "spec/syscalls.txt")
	RootSyscallSpec, _ = spec.New("Root Level Directory", "spec/syscalls.txt")
	RepoSyscallSpec, _ = spec.New("Repo Level directories", "spec/syscalls.txt")

	exitVal := m.Run()
	fmt.Printf("========== Running Post Fuse Test Hooks\n")

	// Now Generate the summary report

	allCommits := spec.NewCombinedSpec([]spec.Spec{*OpenCommitSyscallSpec, *ClosedCommitSyscallSpec})

	allReports := spec.NewCombinedSpec(
		[]spec.Spec{
			*RootSyscallSpec,
			*RepoSyscallSpec,
			*OpenCommitSyscallSpec,
			*ClosedCommitSyscallSpec,
			*CancelledCommitSyscallSpec,
		},
	)

	summary := spec.NewSummary()
	summary.SingleSpecs = []spec.Spec{
		*RootSyscallSpec,
		*RepoSyscallSpec,
		*OpenCommitSyscallSpec,
		*ClosedCommitSyscallSpec,
		*CancelledCommitSyscallSpec,
	}
	summary.CombinedSpecs = []spec.CombinedSpec{
		*allCommits,
		*allReports,
	}

	err := summary.GenerateReport("spec/reports")

	if err != nil {
		fmt.Printf("FAILURE!! Error generating summary: %v\n", err)
		os.Exit(1)
	}

	didSpecReportChange()
	// Todo - if the reports changed, fail CI, because it means this wasn't run
	// locally and couldn't have been run on linux and mac
	os.Exit(exitVal)
}

func didSpecReportChange() {
	// Checks to see if the generated report is different from what git expects.
	// If so ... exit 1 ... any changes to the spec need to be done manually.
	// This will prevent checks 'passing' on CI but the new spec not being committed

	c := exec.Command("git", "status", "-s")
	raw, err := c.Output()
	if err != nil {
		fmt.Printf("Couldn't query git status, so couldn't validate spec freshness")
		os.Exit(1)
	}

	if strings.Contains(string(raw), "reports/summary") {
		fmt.Printf("There are uncommitted changes!\nThe generated spec does not match the existing spec. Please run `make spec-generate` locally to vet and commit the changes as necessary\n")
		fmt.Printf("Uncommitted changes:\n%v\n", string(raw))
		os.Exit(1)
	}

}

package fuse_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/server/pfs/fuse/spec"
)

var OpenCommitSyscallSpec *spec.Spec
var ClosedCommitSyscallSpec *spec.Spec
var RootSyscallSpec *spec.Spec
var RepoSyscallSpec *spec.Spec

func TestMain(m *testing.M) {
	fmt.Println("========== Running Pre Fuse Test Hooks\n")

	OpenCommitSyscallSpec, _ = spec.New("Open Commit", "spec/syscalls.txt")
	ClosedCommitSyscallSpec, _ = spec.New("Closed Commit", "spec/syscalls.txt")
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
		},
	)

	summary := spec.NewSummary()
	summary.SingleSpecs = []spec.Spec{
		*RootSyscallSpec,
		*RepoSyscallSpec,
		*OpenCommitSyscallSpec,
		*ClosedCommitSyscallSpec,
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

	// Todo - if the reports changed, fail CI, because it means this wasn't run
	// locally and couldn't have been run on linux and mac
	os.Exit(exitVal)
}

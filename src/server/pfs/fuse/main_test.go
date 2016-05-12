package fuse_test

import (
	"flag"
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

var regeneratingSpec bool

func init() {
	flag.BoolVar(&regeneratingSpec, "spec.regenerate", false, "Set to true when regenerating the spec")
	flag.Parse()
}

func generateSummaryReport() {
	allCommits := spec.NewCombinedSpec(
		[]spec.Spec{
			*OpenCommitSyscallSpec,
			*ClosedCommitSyscallSpec,
			*CancelledCommitSyscallSpec,
		},
	)

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

}

func TestMain(m *testing.M) {
	// This runs before tests get called
	OpenCommitSyscallSpec, _ = spec.New("Open Commit", "spec/syscalls.txt")
	ClosedCommitSyscallSpec, _ = spec.New("Closed Commit", "spec/syscalls.txt")
	CancelledCommitSyscallSpec, _ = spec.New("Cancelled Commit", "spec/syscalls.txt")
	RootSyscallSpec, _ = spec.New("Root Level Directory", "spec/syscalls.txt")
	RepoSyscallSpec, _ = spec.New("Repo Level directories", "spec/syscalls.txt")

	// Call the tests
	exitVal := m.Run()

	// This runs after tests are called
	if exitVal > 0 {
		os.Exit(exitVal)
	}

	generateSummaryReport()

	if regeneratingSpec {
		fmt.Printf("Successfully regenerated spec reports.\n")
	} else {
		didSpecReportChange()
	}

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

	match := "reports/summary"

	if strings.Contains(string(raw), match) {
		fmt.Printf("There are uncommitted changes!\nThe generated spec does not match the existing spec. Please run `make spec-generate` locally to vet and commit the changes as necessary\n")
		fmt.Printf("Specs that don't match committed versions:\n")

		lines := strings.Split(string(raw), "\n")
		for _, line := range lines {
			if strings.Contains(line, match) {
				fmt.Println(line)
			}
		}

		os.Exit(1)
	}

}

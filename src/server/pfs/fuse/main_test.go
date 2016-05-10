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
	fmt.Println("This gets run BEFORE any tests get run!")

	OpenCommitSyscallSpec, _ = spec.New("Syscalls During an Open Commit", "spec/syscalls.txt")
	ClosedCommitSyscallSpec, _ = spec.New("Syscalls During a Closed Commit", "spec/syscalls.txt")
	RootSyscallSpec, _ = spec.New("Syscalls on root level directory", "spec/syscalls.txt")
	RepoSyscallSpec, _ = spec.New("Syscalls on repo level directories", "spec/syscalls.txt")

	exitVal := m.Run()

	fmt.Println("This gets run AFTER any tests get run!")
	err := OpenCommitSyscallSpec.GenerateReport("spec/reports/syscall-open-commits.html")
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err.Error())
	}
	err = ClosedCommitSyscallSpec.GenerateReport("spec/reports/syscall-closed-commits.html")
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err.Error())
	}
	err = RootSyscallSpec.GenerateReport("spec/reports/syscall-root.html")
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err.Error())
	}
	err = RepoSyscallSpec.GenerateReport("spec/reports/syscall-repo.html")
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err.Error())
	}

	spec.CombinedReport([]spec.Spec{*OpenCommitSyscallSpec, *ClosedCommitSyscallSpec}, "spec/reports/syscall-commits.html")

	// Todo - if the reports changed, fail CI, because it means this wasn't run
	// locally and couldn't have been run on linux and mac
	os.Exit(exitVal)
}

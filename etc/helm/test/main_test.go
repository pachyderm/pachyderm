package helmtest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func TestMain(m *testing.M) {
	// This puts the bazel helm into $PATH, because terratest is hard-coded to look there.
	helmRel, err := runfiles.Rlocation("_main/tools/helm/_helm")
	if err != nil {
		fmt.Fprintf(os.Stderr, "find helm runfiles: %v", err)
		os.Exit(1)
	}
	helm, err := filepath.Abs(helmRel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "find absolute path of helm: %v", err)
		os.Exit(1)
	}
	tmp, err := os.MkdirTemp("", "helm")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create /tmp/helm: %v", err)
		os.Exit(1)
	}
	// We don't bother cleaning up this directory because Bazel does it.
	if err := os.Symlink(helm, filepath.Join(tmp, "helm")); err != nil {
		fmt.Fprintf(os.Stderr, "symlink helm into /tmp/helm: %v", err)
		os.Exit(1)
	}
	path := os.Getenv("PATH")
	os.Setenv("PATH", tmp+":"+path)
	os.Exit(m.Run())
}

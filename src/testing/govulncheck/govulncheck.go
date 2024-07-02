package govulncheck

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// SetupEnv sets up the environment for govulncheck (which shells out to Go), and returns the
// working directory for the _main module.
func SetupEnv() string {
	// Find the go Bazel uses and put it in $PATH.
	if goBin, err := runfiles.Rlocation("rules_go~~go_sdk~go_sdk/bin/go"); err != nil {
		fmt.Fprintf(os.Stderr, "can't find go binary; using installed go instead: %v\n", err)
	} else {
		path := os.Getenv("PATH")
		path = filepath.Dir(goBin) + ":" + path
		os.Setenv("PATH", path)
	}

	// Use a common cache between tests and the //tools/govulncheck command.  We avoid
	// $XDG_CONFIG_CACHE because that's the cache for the "host" go and govulncheck.  This might
	// be a different version of Go than what the user has installed.
	os.Setenv("GOPATH", "/tmp/pach-govulncheck/gopath")
	os.Setenv("GOCACHE", "/tmp/pach-govulncheck/gocache")

	// Find the source code.
	mod, err := runfiles.Rlocation("_main/MODULE.bazel")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't find MODULE.bazel: %v\n", err)
		mod = "."
	}
	stat, err := os.Lstat(mod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stat MODULE.bazel: %v\n", err)
	}
	if stat.Mode()&fs.ModeSymlink == 0 {
		return filepath.Dir(mod)
	}
	real, err := os.Readlink(mod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read MODULE.bazel runfile symlink: %v\n", err)
		return filepath.Dir(mod)
	}
	return filepath.Dir(real)
}

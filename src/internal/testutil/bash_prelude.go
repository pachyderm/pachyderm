package testutil

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func writeBashPrelude(w io.Writer) error {
	formats := []string{bashPrelude, "\n", matchTest, "\n"}
	formatArgs := [][]any{{}, {}, {}, {}}

	if matchBin, ok := bazel.FindBinary("//src/testing/match", "match"); ok {
		// Running with Bazel; stick built-for-this-invocation match binary in $PATH in the
		// generated prelude.  It is unlikely that Bazel will be misconfigured; the match
		// binary is a data dependency of this file, so if this code is running, match is in
		// the runfiles directory.
		formats = []string{bashPrelude, "\n", matchPath, "\n", matchTest, "\n"}
		formatArgs = [][]any{{}, {}, {filepath.Dir(matchBin)}, {}, {}, {}}
	}
	for i, f := range formats {
		args := formatArgs[i]
		if _, err := fmt.Fprintf(w, f, args...); err != nil {
			return errors.Wrapf(err, "format prelude (fmt.Printf(%v, %v)", f, args)
		}
	}
	return nil
}

var (
	bashPrelude = `set -e -o pipefail
# Try to ignore pipefail errors (encountered when writing to a closed pipe).
# Processes like 'yes' are essentially guaranteed to hit this, and because of
# -e -o pipefail they will crash the whole script. We need these options,
# though, for 'match' to work, so for now we work around pipefail errors on a
# cmd-by-cmd basis. See "The Infamous SIGPIPE Signal"
# http://www.tldp.org/LDP/lpg/node20.html
pipeerr=141 # typical error code returned by unix utils when SIGPIPE is raised
function yes {
	/usr/bin/yes || test "$?" -eq "${pipeerr}"
}
export -f yes # use in subshells too`

	matchPath = `PATH=$PATH:%s`

	matchTest = `which match >/dev/null || {
	echo "You must have 'match' installed to run these tests. Please run:" >&2
	echo "  go install ./src/testing/match" >&2
	exit 1
}`
)

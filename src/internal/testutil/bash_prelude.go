package testutil

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func writeBashPrelude(w io.Writer) error {
	var path []string

	// prelude
	if _, err := w.Write([]byte(bashPrelude + "\n")); err != nil {
		return errors.Wrap(err, "write bash prelude")
	}

	// runfiles
	if match, ok := bazel.FindBinary("//src/testing/match", "match"); ok {
		// Running with Bazel; stick built-for-this-invocation match binary in $PATH in the
		// generated prelude.  It is unlikely that Bazel will be misconfigured; the match
		// binary is a data dependency of this file, so if this code is running, match is in
		// the runfiles directory.
		path = append(path, filepath.Dir(match))
	}
	if pachctl, ok := bazel.FindBinary("//src/server/cmd/pachctl", "pachctl"); ok {
		path = append(path, filepath.Dir(pachctl))
	}
	if len(path) > 0 {
		if _, err := fmt.Fprintf(w, matchPathFormat+"\n", strings.Join(path, ":")); err != nil {
			return errors.Wrap(err, "write $PATH override")
		}
	}

	// match tester
	if _, err := w.Write([]byte(matchTest + "\n")); err != nil {
		return errors.Wrap(err, "write `match` tester")
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
export -f yes # use in subshells too
`

	matchPathFormat = `PATH=%s:$PATH`

	matchTest = `which match >/dev/null || {
	echo "You must have 'match' installed to run these tests. Please run:" >&2
	echo "  go install ./src/testing/match" >&2
	exit 1
}
`
)

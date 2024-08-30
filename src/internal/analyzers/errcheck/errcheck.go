package errcheck

import "github.com/kisielk/errcheck/errcheck"

func init() {
	// The default excluded symbols are things like fmt.Print, (*bytes.Buffer).Write, etc.
	errcheck.DefaultExcludedSymbols = append(errcheck.DefaultExcludedSymbols,
		"(*database/sql.Tx).Rollback",
		"(*github.com/spf13/cobra.Command).MarkFlagCustom",
		"(*github.com/spf13/cobra.Command).Usage",
		"(*github.com/pachyderm/pachyderm/v2/src/internal/client.APIClient).Close",
		"(*github.com/pachyderm/pachyderm/v2/src/client.APIClient).Close",

		// Avoid flagging os.Stdout/os.Stderr writes; other files should be checked, but the
		// codebase expects this to not be a lint error right now.
		"(*os.File).Write",
		"fmt.Fprintf",
		"fmt.Fprintln",
		"fmt.Fprint",
		"(*github.com/juju/ansiterm.TabWriter).Flush",

		// Avoid flagging (net/http).Response.Body.Close,
		// opentracing.GlobalTracer().Close(), etc.; these should be checked, of course.
		"(io.ReadCloser).Close",
		"(io.Closer).Close",
		"(*database/sql.Rows).Close",

		// Syscalls that rarely fail (but should be fixed).
		"os.Setenv",
		"os.Unsetenv",
	)
}

var Analyzer = errcheck.Analyzer

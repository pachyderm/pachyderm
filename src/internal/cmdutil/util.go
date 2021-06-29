package cmdutil

import (
	"fmt"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/spf13/pflag"
)

// Encoder creates an encoder that writes data structures to the given writer in
// the serialization format 'format'.  If the 'format' passed is unrecognized
// (currently, 'format' must be 'json' or 'yaml') then pachctl exits
// immediately.
func Encoder(format string, w io.Writer) serde.Encoder {
	format = strings.ToLower(format)
	if format == "" {
		format = "json"
	}
	e, err := serde.GetEncoder(format, w,
		serde.WithIndent(2),
		serde.WithOrigName(true),
	)
	if err != nil {
		// TODO: would be nice to not have this exit as it means we can't exercise
		// this branch in tests, but not having this function return an error saves
		// a lot of boilerplate.
		ErrorAndExit("%s", err.Error())
	}
	return e
}

func OutputFlags(raw *bool, output *string) *pflag.FlagSet {
	outputFlags := pflag.NewFlagSet("", pflag.ExitOnError)
	outputFlags.BoolVar(raw, "raw", false, "Disable pretty printing; serialize data structures to an encoding such as json or yaml")
	// --output is empty by default, so that we can print an error if a user
	// explicitly sets --output without --raw, but the effective default is set in
	// encode(), which assumes "json" if 'format' is empty.
	// Note: because of how spf13/flags works, no other StringVarP that sets
	// 'output' can have a default value either
	outputFlags.StringVarP(output, "output", "o", "", "Output format when --raw is set: \"json\" or \"yaml\" (default \"json\")")
	return outputFlags
}

func TimestampFlags(fullTimestamps *bool) *pflag.FlagSet {
	timestampFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	timestampFlags.BoolVar(fullTimestamps, "full-timestamps", false, "Return absolute timestamps (as opposed to the default, relative timestamps).")
	return timestampFlags
}

func PagerFlags(noPager *bool) *pflag.FlagSet {
	pagerFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	pagerFlags.BoolVar(noPager, "no-pager", false, "Don't pipe output into a pager (i.e. less).")
	return pagerFlags
}

func readLine(r io.Reader) (string, error) {
	// Read one byte at a time to avoid over-reading
	result := ""
	buf := make([]byte, 1)
	for string(buf) != "\n" {
		_, err := r.Read(buf)
		if err != nil {
			return "", err
		}
		result += string(buf)
	}
	return result, nil
}

// InteractiveConfirm will ask the user to confirm an action on the command-line with a y/n response.
func InteractiveConfirm(env Env) (bool, error) {
	fmt.Fprintf(env.Err(), "Are you sure you want to do this? (y/N): ")
	s, err := readLine(env.In())
	if err != nil {
		return false, err
	}
	if s[0] == 'y' || s[0] == 'Y' {
		return true, nil
	}
	return false, nil
}

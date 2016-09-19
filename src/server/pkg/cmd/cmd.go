package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/spf13/cobra"
)

// RunFixedArgs wraps a function in a function
// that checks its exact argument count.
func RunFixedArgs(numArgs int, run func([]string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) != numArgs {
			fmt.Printf("expected %d arguments, got %d\n\n", numArgs, len(args))
			cmd.Usage()
		} else {
			if err := run(args); err != nil {
				ErrorAndExit("%v", err)
			}
		}
	}
}

// RunBoundedArgs wraps a function in a function
// that checks its argument count is within a range.
func RunBoundedArgs(min int, max int, run func([]string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) < min || len(args) > max {
			fmt.Printf("expected %d to %d arguments, got %d\n\n", min, max, len(args))
			cmd.Usage()
		} else {
			if err := run(args); err != nil {
				ErrorAndExit("%v", err)
			}
		}
	}
}

// ErrorAndExit errors with the given format and args, and then exits.
func ErrorAndExit(format string, args ...interface{}) {
	if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
		fmt.Fprintf(os.Stderr, "%s\n", errString)
	}
	os.Exit(1)
}

// ParseCommits takes a slice of arguments of the form "repo/commit-id" or
// "repo" (in which case we consider the commit ID to be empty), and returns
// a list of Commits
func ParseCommits(args []string) ([]*pfs.Commit, error) {
	var commits []*pfs.Commit
	for _, arg := range args {
		parts := strings.SplitN(arg, "/", 2)
		commit := &pfs.Commit{
			Repo: &pfs.Repo{
				Name: parts[0],
			},
		}
		if len(parts) == 2 {
			commit.ID = parts[1]
		} else {
			commit.ID = ""
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

// RepeatedStringArg is an alias for []string
type RepeatedStringArg []string

func (r *RepeatedStringArg) String() string {
	result := "["
	for i, s := range *r {
		if i != 0 {
			result += ", "
		}
		result += s
	}
	return result + "]"
}

// Set adds a string to r
func (r *RepeatedStringArg) Set(s string) error {
	*r = append(*r, s)
	return nil
}

// Type returns the string representation of the type of r
func (r *RepeatedStringArg) Type() string {
	return "[]string"
}

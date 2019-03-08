package cmdutil

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

// Run makes a new cobra run function that wraps the given function.
func Run(run func(args []string) error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		if err := run(args); err != nil {
			ErrorAndExit(err.Error())
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
// a list of *pfs.Commits
func ParseCommits(args []string) ([]*pfs.Commit, error) {
	var commits []*pfs.Commit
	for _, arg := range args {
		parts := strings.SplitN(arg, "/", 2)
		hasRepo := len(parts) > 0 && parts[0] != ""
		hasCommit := len(parts) == 2 && parts[1] != ""
		if hasCommit && !hasRepo {
			return nil, fmt.Errorf("invalid commit id \"%s\": repo cannot be empty", arg)
		}
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

// ParseBranches takes a slice of arguments of the form "repo/branch-name" or
// "repo" (in which case we consider the branch name to be empty), and returns
// a list of *pfs.Branches
func ParseBranches(args []string) ([]*pfs.Branch, error) {
	commits, err := ParseCommits(args)
	if err != nil {
		return nil, err
	}
	var result []*pfs.Branch
	for _, commit := range commits {
		result = append(result, &pfs.Branch{Repo: commit.Repo, Name: commit.ID})
	}
	return result, nil
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

func MakeBatchCommand(
    positionalCount int,
    cmd *cobra.Command,
    provision func(*cobra.Command),
    processFlags func (),
    run func ([][]string) error,
) {
    provision(cmd)
    cmd.Run = func(cmd *cobra.Command, _ []string) {
        // Remove non parameter args from the original args so we can reparse
        // them iteratively with cobra.
        names := []string{cmd.Name()}
        names = append(names, cmd.Aliases...)
        var startArg int
        for i, x := range os.Args {
            for _, y := range names {
                if x == y {
                    startArg = i + 1
                    break
                }
            }
        }
        args := os.Args[startArg:]

        // Partition by sets of positional args, flags apply to the previous positional args
        sets := [][]string{{}}
        index := 0
        count := 0

        for _, x := range args {
            if strings.HasPrefix(x, "-") {
                sets[index] = append(sets[index], x)
                // TODO: this only works because all flags require a value, otherwise we'll need to check individual flags
                if !strings.Contains(x, "=") {
                    count -= 1
                }
            } else if count < positionalCount {
                count += 1
                sets[index] = append(sets[index], x)
            } else {
                sets = append(sets, []string{x})
                index += 1
                count = 0
            }
        }

        parsedSets := [][]string{}

        // Create an inner command for parsing individual commands
        innerCommand := &cobra.Command{}
        provision(innerCommand)

        // Copy over inherited flags so we don't choke on them
        ancestor := cmd.Parent()
        for ancestor != nil {
            innerCommand.Flags().AddFlagSet(ancestor.Flags())
            ancestor = ancestor.Parent()
        }

        // Set a run function that appends to our set of parsed args
        innerCommand.Run = func (runCmd *cobra.Command, runArgs []string) {
            parsedSets = append(parsedSets, runArgs)
        }

        for _, argSet := range sets {
            innerCommand.SetArgs(argSet)
            innerCommand.Execute()
            processFlags()
        }

        fmt.Printf("flags: %s\n", innerCommand.LocalFlags())

        run(parsedSets)
    }
}

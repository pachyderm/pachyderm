package cmdutil

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// This layer of indirection is so we can capture calls to ErrorAndExit in tests
var internalErrorAndExit = func(message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, "%s\n", message)
	}
	os.Exit(1)
}

// ErrorAndExit errors with the given format and args, and then exits.
func ErrorAndExit(format string, args ...interface{}) {
	internalErrorAndExit(strings.TrimSpace(fmt.Sprintf(format, args...)))
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

// BatchArgs is a struct used to pass arguments for one of many calls to a
// batch command registered using 'RunBatchCommand'.
type BatchArgs struct {
	Positionals []string
	Flags map[string][]string
}

// GetStringFlag is a helper function for accessing BatchArgs.Flags.  In order
// to avoid type problems, BatchArgs assumes all flags may be repeated strings.
// This helper function should be used when a flag should only have one value,
// it will return the last-specified value for that flag, or the zero-value if
// the flag is not specified.
func (args *BatchArgs) GetStringFlag(name string) string {
	length := len(args.Flags[name])
	if length > 0 {
		return args.Flags[name][length - 1]
	}
	return ""
}

// partitionBatchArgs splits up command-line arguments into groups of equal size
// determined by positionalCount.  Flags are grouped with their preceeding
// positional arguments.  This function mimics the parsing behavior of cobra,
// so it may be a bit sensitive to cobra's behavior changing.
func partitionBatchArgs(args []string, positionalCount int, cmd *cobra.Command) ([][]string, error) {
	sets := [][]string{{}}
	index := 0
	count := 0
	value := false

	for _, x := range args {
		if value {
			sets[index] = append(sets[index], x)
			value = false
		} else if strings.HasPrefix(x, "--") {
			sets[index] = append(sets[index], x)
			split := strings.SplitN(x[2:], "=", 2)
			flag := cmd.Flags().Lookup(split[0])

			if flag == nil {
				return nil, fmt.Errorf("Error: unknown flag: %s", x)
			} else if flag.Name == "help" {
				return nil, errors.New("")
			} else if len(split) == 1 && flag.NoOptDefVal == "" {
				// The following arg is the value for this flag, make sure to include it
				value = true
			}
		} else if strings.HasPrefix(x, "-") && len(x) > 1 {
			sets[index] = append(sets[index], x)

			// Iterate through shorthand flags until we find one that requires a value
			for i, shorthand := range x[1:] {
				flag := cmd.Flags().ShorthandLookup(string(shorthand))
				if flag == nil {
					return nil, fmt.Errorf("Error: unknown shorthand flag: '%c' in %s", shorthand, x)
				} else if flag.Name == "help" {
					return nil, errors.New("")
				} else if flag.NoOptDefVal == "" {
					if len(x) == i + 2 {
						// The following arg is the value for this flag, make sure to include it
						value = true
					}
					break
				}
			}
		} else if count < positionalCount {
			sets[index] = append(sets[index], x)
			count++
		} else {
			sets = append(sets, []string{x})
			index++
			count = 1
		}
	}

	if count != positionalCount {
		lastSetString := strings.Join(sets[len(sets) - 1], " ")
		return nil, fmt.Errorf("each request must have %d arguments, but found %d in:\n  %s", positionalCount, count, lastSetString)
	}

	return sets, nil
}

// newBatchParserCommand creates a dummy command parser that inherits flags from
// the 'originalCommand', but is not attached to the main command tree.  This
// gives us a bit more freedom to iteratively parse a batch of commands without
// jumping through as many hoops.
func newBatchParserCommand(originalCommand *cobra.Command) *cobra.Command {
	// Make a list of commands in the hierarchy from this command up to the root
	var hierarchy []*cobra.Command
	for cmd := originalCommand; cmd != nil; cmd = cmd.Parent() {
		hierarchy = append(hierarchy, cmd)
	}

	// Apparently cobra only runs one each of PersistentPreRun and
	// PersistentPostRun, so we don't have to wrap them.  Just pull out the
	// first one that cobra would run and copy it over.
	var preRun func(*cobra.Command, []string)
	var preRunE func(*cobra.Command, []string) error
	for _, cmd := range hierarchy {
		if cmd.PersistentPreRunE != nil {
			preRunE = cmd.PersistentPreRunE
			break
		} else if cmd.PersistentPreRun != nil {
			preRun = cmd.PersistentPreRun
			break
		}
	}

	var postRun func(*cobra.Command, []string)
	var postRunE func(*cobra.Command, []string) error
	for _, cmd := range hierarchy {
		if cmd.PersistentPostRunE != nil {
			postRunE = cmd.PersistentPostRunE
			break
		} else if cmd.PersistentPostRun != nil {
			postRun = cmd.PersistentPostRun
			break
		}
	}

	// Create an inner command for parsing individual commands
	result := &cobra.Command{
		PersistentPreRunE: preRunE,
		PersistentPreRun: preRun,
		PersistentPostRunE: postRunE,
		PersistentPostRun: postRun,
	}

	// Copy over flags from the hierarchy
	result.Flags().AddFlagSet(originalCommand.LocalFlags())
	for _, cmd := range hierarchy {
		result.Flags().AddFlagSet(cmd.PersistentFlags())
	}

	return result
}

// RunBatchCommand generates a Run function for a batch CLI command.  This works
// the same as a normal command except that it parses any number of sets of
// arguments from the command-line args, each set corresponding to a request.
// These parsed sets will be passed to the specified 'run' callback, which
// can then be validated and handled.  The number of positional arguments to
// each request must be fixed (as set by 'positionalCount'), and any flags
// passed to the command will bind to their preceeding positional arguments.
func RunBatchCommand(
	positionalCount int,
	run func([]BatchArgs) error,
) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, originalArgs []string) {
		// As long as there's any test coverage, hopefully this will save some time
		if !cmd.DisableFlagParsing {
			ErrorAndExit("ERROR: batch commands must disable flag parsing")
		}

		// Partition the args based on the number of positional arguments
		args, err := partitionBatchArgs(originalArgs, positionalCount, cmd)
		if err != nil {
			ErrorAndExit("%v\n%s", err, cmd.UsageString())
		}

		parsedArgs := []BatchArgs{}
		innerCommand := newBatchParserCommand(cmd)

		for _, argSet := range args {
			positionals := []string{}
			innerCommand.Run = func (runCmd *cobra.Command, runArgs []string) {
				positionals = append(positionals, runArgs...)
			}

			// This will collect the positionals - we discard parsed flags
			innerCommand.SetArgs(argSet)
			innerCommand.Execute()

			// We then reprocess flag parsing using a lower-level interface
			flags := map[string][]string{}
			innerCommand.Flags().ParseAll(
				argSet,
				func (flag *pflag.Flag, value string) error {
					flags[flag.Name] = append(flags[flag.Name], value)
					return nil
				},
			)

			parsedArgs = append(parsedArgs, BatchArgs{positionals, flags})
		}

		if err = run(parsedArgs); err != nil {
			ErrorAndExit("%v", err)
		}
	}
}

// SetDocsUsage sets the usage string for a docs-style command.  Docs commands
// have no functionality except to output some docs and related commands, and
// should not specify a 'Run' attribute.
func SetDocsUsage(command *cobra.Command, subcommands []*cobra.Command) {
	command.SetUsageTemplate(`Usage:
  pachctl [command]{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`)

	command.SetHelpTemplate(`{{or .Long .Short}}
{{.UsageString}}`)

	// This song-and-dance is so that we can render the related commands without
	// actually having them usable as subcommands of the docs command.
	// That is, we don't want `pachctl job list-job` to work, it should just
	// be `pachctl list-job`.  Therefore, we lazily add/remove the subcommands
	// only when we try to render usage for the docs command.
	originalUsage := command.UsageFunc()
	command.SetUsageFunc(func (c *cobra.Command) error {
		newUsage := command.UsageFunc()
		command.SetUsageFunc(originalUsage)
		defer command.SetUsageFunc(newUsage)

		command.AddCommand(subcommands...)
		defer command.RemoveCommand(subcommands...)

		command.Usage()
		return nil
	})
}

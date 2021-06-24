package cmdutil

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
)

// PrintErrorStacks should be set to true if you want to print out a stack for
// errors that are returned by the run commands.
var PrintErrorStacks bool

type clientOnce struct {
	once sync.Once
	eg   errgroup.Group
	c    *client.APIClient
}

func (c *clientOnce) Get(cb func() (*client.APIClient, error)) *client.APIClient {
	c.once.Do(func() {
		c.eg.Go(func() error {
			var err error
			c.c, err = cb()
			return err
		})
	})
	if err := c.eg.Wait(); err != nil {
		ErrorAndExit("%v", errors.Wrap(err, "failed to connect"))
	}
	return c.c
}

func (c *clientOnce) Close() error {
	c.eg.Wait()
	if c.c != nil {
		return c.c.Close()
	}
	return nil
}

// Env provides a mockable interface for testing pachctl commands
type Env interface {
	// Client returns the client for talking to pachd
	Client(string, ...client.Option) *client.APIClient
	// EnterpriseClient returns the client for talking to the enterprise server
	EnterpriseClient(string, ...client.Option) *client.APIClient

	// In returns the input reader (defaults to os.Stdin)
	In() io.Reader
	// Out returns the output writer (defaults to os.Stdout)
	Out() io.Writer
	// Err returns the error writer (defaults to os.Stderr)
	Err() io.Writer

	// Close is called automatically at the end of the Run function and closes any
	// open connections.
	Close() error
}

type realEnv struct {
	cmd              *cobra.Command
	pachClient       clientOnce
	enterpriseClient clientOnce
}

func (env *realEnv) Client(name string, options ...client.Option) *client.APIClient {
	return env.pachClient.Get(func() (*client.APIClient, error) {
		if inWorkerStr, ok := os.LookupEnv("PACH_IN_WORKER"); ok {
			inWorker, err := strconv.ParseBool(inWorkerStr)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't parse PACH_IN_WORKER")
			}
			if inWorker {
				return client.NewInWorker(options...)
			}
		}
		return client.NewOnUserMachine(name, options...)
	})
}

func (env *realEnv) EnterpriseClient(name string, options ...client.Option) *client.APIClient {
	return env.enterpriseClient.Get(func() (*client.APIClient, error) {
		c, err := client.NewEnterpriseClientOnUserMachine(name, client.WithDialTimeout(30*time.Second))
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(env.Out(), "Using enterprise context: %v\n", c.ClientContextName())
		return c, nil
	})
}

func (env *realEnv) In() io.Reader {
	return env.cmd.InOrStdin()
}

func (env *realEnv) Out() io.Writer {
	return env.cmd.OutOrStdout()
}

func (env *realEnv) Err() io.Writer {
	return env.cmd.ErrOrStderr()
}

func (env *realEnv) Close() (retErr error) {
	setError := func(err error) {
		if retErr == nil {
			retErr = err
		}
	}
	setError(env.pachClient.Close())
	setError(env.enterpriseClient.Close())
	return retErr
}

func envFromCommand(cmd *cobra.Command) Env {
	if env := cmd.Context().Value("env"); env != nil {
		return env.(Env)
	}
	return &realEnv{cmd: cmd}
}

func withEnv(cmd *cobra.Command, cb func(Env) error) (retErr error) {
	env := envFromCommand(cmd)
	defer func() {
		if err := env.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(env)
}

// RunFixedArgs wraps a function in a function
// that checks its exact argument count.
func RunFixedArgs(numArgs int, run func([]string, Env) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return withEnv(cmd, func(env Env) error {
			if len(args) != numArgs {
				fmt.Fprintf(env.Err(), "expected %d arguments, got %d\n\n", numArgs, len(args))
				cmd.Usage()
				return nil
			}
			return run(args, env)
		})
	}
}

// RunBoundedArgs wraps a function in a function
// that checks its argument count is within a range.
func RunBoundedArgs(min int, max int, run func([]string, Env) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return withEnv(cmd, func(env Env) error {
			if len(args) < min || len(args) > max {
				fmt.Fprintf(env.Err(), "expected %d to %d arguments, got %d\n\n", min, max, len(args))
				cmd.Usage()
				return nil
			}
			return run(args, env)
		})
	}
}

// RunMinimumArgs wraps a function in a function
// that checks its argument count is above a minimum amount
func RunMinimumArgs(min int, run func([]string, Env) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return withEnv(cmd, func(env Env) error {
			if len(args) < min {
				fmt.Fprintf(env.Err(), "expected at least %d arguments, got %d\n\n", min, len(args))
				cmd.Usage()
				return nil
			}
			return run(args, env)
		})
	}
}

// Run makes a new cobra run function that wraps the given function.
func Run(run func([]string, Env) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return withEnv(cmd, func(env Env) error {
			return run(args, env)
		})
	}
}

// ErrorAndExit errors with the given format and args, and then exits.
func ErrorAndExit(format string, args ...interface{}) {
	if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
		fmt.Fprintf(os.Stderr, "%s\n", errString)
	}
	if len(args) > 0 && PrintErrorStacks {
		if err, ok := args[0].(error); ok {
			errors.ForEachStackFrame(err, func(frame errors.Frame) {
				fmt.Fprintf(os.Stderr, "%+v\n", frame)
			})
		}
	}
	os.Exit(1)
}

func isValidBranch(name string) bool {
	err := ancestry.ValidateName(name)
	return err == nil
}

func ParseRepo(name string) *pfs.Repo {
	var repo pfs.Repo
	if strings.Contains(name, ".") {
		repoParts := strings.SplitN(name, ".", 2)
		repo.Name = repoParts[0]
		repo.Type = repoParts[1]
	} else {
		repo.Name = name
		repo.Type = pfs.UserRepoType
	}
	return &repo
}

// Parses the following formats, any unspecified fields will be left as empty
// strings in the pfs.File structure.  The second return value is the number of fields parsed -
// (1: repo only, 2: repo and branch-or-commit, 3: repo, branch, and file).
//   repo
//   repo@branch
//   repo@branch:path
//   repo@branch=commit
//   repo@branch=commit:path
//   repo@commit
//   repo@commit:path
func parseFile(arg string) (*pfs.File, int, error) {
	var repo, branch, commit, path string
	parts := strings.SplitN(arg, "@", 2)
	if parts[0] == "" {
		return nil, 0, errors.Errorf("invalid format \"%s\": repo cannot be empty", arg)
	}
	numFields := 1
	repo = parts[0]

	if len(parts) == 2 {
		numFields = 2
		parts = strings.SplitN(parts[1], ":", 2)
		if len(parts) == 2 {
			numFields = 3
			path = parts[1]
		}

		parts = strings.SplitN(parts[0], "=", 2)
		if len(parts) == 1 {
			if uuid.IsUUIDWithoutDashes(parts[0]) || !isValidBranch(parts[0]) {
				commit = parts[0]
			} else {
				branch = parts[0]
			}
		} else if len(parts) == 2 {
			branch = parts[0]
			commit = parts[1]
		}
	}
	return ParseRepo(repo).NewCommit(branch, commit).NewFile(path), numFields, nil
}

// ParseCommit takes an argument of the form "repo[@branch-or-commit]" and
// returns the corresponding *pfs.Commit.
func ParseCommit(arg string) (*pfs.Commit, error) {
	file, numFields, err := parseFile(arg)
	if err != nil {
		return nil, err
	}
	if numFields > 2 {
		return nil, errors.Errorf("invalid format \"%s\": cannot specify a file path")
	}
	return file.Commit, nil
}

// ParseCommits converts all arguments to *pfs.Commit structs using the
// semantics of ParseCommit
func ParseCommits(args []string) ([]*pfs.Commit, error) {
	var results []*pfs.Commit
	for _, arg := range args {
		commit, err := ParseCommit(arg)
		if err != nil {
			return nil, err
		}
		results = append(results, commit)
	}
	return results, nil
}

// ParseBranch takes an argument of the form "repo[@branch]" and returns the
// corresponding *pfs.Branch. This uses ParseCommit under the hood because a
// branch name is semantically interchangeable with a commit-id on the
// command-line.
func ParseBranch(arg string) (*pfs.Branch, error) {
	commit, err := ParseCommit(arg)
	if err != nil {
		return nil, err
	}
	return commit.Branch, nil
}

// ParseJob takes an argument of the form "pipeline@job-id" and returns
// the corresponding *pps.Job.
func ParseJob(arg string) (*pps.Job, error) {
	parts := strings.SplitN(arg, "@", 2)
	if parts[0] == "" {
		return nil, errors.Errorf("invalid format \"%s\": pipeline must be specified", arg)
	}
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid format \"%s\": expected pipeline@job-id", arg)
	}
	return client.NewJob(parts[0], parts[1]), nil
}

// ParseBranches converts all arguments to *pfs.Commit structs using the
// semantics of ParseBranch.
func ParseBranches(args []string) ([]*pfs.Branch, error) {
	var results []*pfs.Branch
	for _, arg := range args {
		branch, err := ParseBranch(arg)
		if err != nil {
			return nil, err
		}
		results = append(results, branch)
	}
	return results, nil
}

// ParseFile takes an argument of the form "repo[@branch-or-commit[:path]]", and
// returns the corresponding *pfs.File.
func ParseFile(arg string) (*pfs.File, error) {
	file, _, err := parseFile(arg)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// ParsePartialFile returns the same thing as ParseFile, unless ParseFile would
// error on this input, in which case it returns as much as it was able to
// parse.
func ParsePartialFile(arg string) *pfs.File {
	// ParseFile already returns as much as it can parse since all fields are
	// optional - the only thing we have to do is discard any errors.
	file, err := ParseFile(arg)
	if err == nil {
		return file
	}
	return client.NewFile(arg, "", "", "")
}

// ParseFiles converts all arguments to *pfs.Commit structs using the
// semantics of ParseFile
func ParseFiles(args []string) ([]*pfs.File, error) {
	var results []*pfs.File
	for _, arg := range args {
		commit, err := ParseFile(arg)
		if err != nil {
			return nil, err
		}
		results = append(results, commit)
	}
	return results, nil
}

// ParseHistory parses a --history flag argument. Permissable values are "all"
// "none", and integers greater than or equal to -1 (as strings).
func ParseHistory(history string) (int64, error) {
	if history == "all" {
		return -1, nil
	}
	if history == "none" {
		return 0, nil
	}
	result, err := strconv.Atoi(history)
	return int64(result), err
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

// CreateAlias generates a nested command tree for the invocation specified,
// which should be space-delimited as on the command-line.  The 'Use' field of
// 'cmd' should specify '{{alias}}' instead of the command name as that will be
// filled in based on each invocation.  Similarly, for the 'Example' field,
// '{{alias}}' will be replaced with the full command path.  These commands can
// later be merged into the final Command tree using 'MergeCommands' below.
func CreateAlias(cmd *cobra.Command, invocation string) *cobra.Command {

	// Create logical commands for each substring in each invocation
	var root, prev *cobra.Command
	args := strings.Split(invocation, " ")

	for i, arg := range args {
		cur := &cobra.Command{}

		// The leaf command node should include the usage from the given cmd,
		// while logical nodes just need one piece of the invocation.
		if i == len(args)-1 {
			*cur = *cmd
			if cmd.Use == "" {
				cur.Use = arg
			} else {
				cur.Use = strings.Replace(cmd.Use, "{{alias}}", arg, -1)
			}
			cur.Example = strings.Replace(cmd.Example, "{{alias}}", fmt.Sprintf("%s %s", os.Args[0], invocation), -1)
		} else {
			cur.Use = arg
		}

		if root == nil {
			root = cur
		} else if prev != nil {
			prev.AddCommand(cur)
		}
		prev = cur
	}

	return root
}

// MergeCommands merges several command aliases (generated by 'CreateAlias'
// above) into a single coherent cobra command tree (with root command 'root').
// Because 'CreateAlias' generates empty commands to preserve the right
// hierarchy, we go through a little extra effort to allow intermediate 'docs'
// commands to be preserved in the final command structure.
func MergeCommands(root *cobra.Command, children []*cobra.Command) {
	// Implement our own 'find' function because Command.Find is not reliable?
	findCommand := func(parent *cobra.Command, name string) *cobra.Command {
		for _, cmd := range parent.Commands() {
			if cmd.Name() == name {
				return cmd
			}
		}
		return nil
	}

	// Sort children by max nesting depth - this will put 'docs' commands first,
	// so they are not overwritten by logical commands added by aliases.
	var depth func(*cobra.Command) int
	depth = func(cmd *cobra.Command) int {
		maxDepth := 0
		for _, subcmd := range cmd.Commands() {
			subcmdDepth := depth(subcmd)
			if subcmdDepth > maxDepth {
				maxDepth = subcmdDepth
			}
		}
		return maxDepth + 1
	}

	sort.Slice(children, func(i, j int) bool {
		return depth(children[i]) < depth(children[j])
	})

	// Move each child command over to the main command tree recursively
	for _, cmd := range children {
		parent := findCommand(root, cmd.Name())
		if parent == nil {
			root.AddCommand(cmd)
		} else {
			MergeCommands(parent, cmd.Commands())
		}
	}
}

// CreateDocsAlias sets the usage string for a docs-style command.  Docs
// commands have no functionality except to output some docs and related
// commands, and should not specify a 'Run' attribute.
func CreateDocsAlias(command *cobra.Command, invocation string, pattern string) *cobra.Command {
	// This should create a linked-list-shaped tree, follow it to the one leaf
	root := CreateAlias(command, invocation)
	command = root
	for len(command.Commands()) != 0 {
		command = command.Commands()[0]
	}

	// Normally cobra will not render usage if the command is not runnable or has
	// no subcommands, specify our own help template to override that.
	command.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimRightSpace}}

{{end}}{{.UsageString}}`)

	originalUsageFunc := command.UsageFunc()
	command.SetUsageFunc(func(cmd *cobra.Command) error {
		// commands inherit their parents' usage function, so we'll want to pass-
		// through anything but usage for this specific command
		if cmd.CommandPath() != command.CommandPath() {
			return originalUsageFunc(cmd)
		}
		rootCmd := cmd.Root()

		// Walk the command tree, finding commands with the documented word
		var associated []*cobra.Command
		var walk func(*cobra.Command)
		walk = func(cursor *cobra.Command) {
			isMatch, _ := regexp.MatchString(pattern, cursor.CommandPath())
			if isMatch && cursor.CommandPath() != cmd.CommandPath() && cursor.Runnable() {
				associated = append(associated, cursor)
			}
			for _, subcmd := range cursor.Commands() {
				walk(subcmd)
			}
		}
		walk(rootCmd)

		var maxCommandPath int
		for _, x := range associated {
			commandPathLen := len(x.CommandPath())
			if commandPathLen > maxCommandPath {
				maxCommandPath = commandPathLen
			}
		}

		templateFuncs := template.FuncMap{
			"pad": func(s string) string {
				format := fmt.Sprintf("%%-%ds", maxCommandPath+1)
				return fmt.Sprintf(format, s)
			},
			"associated": func() []*cobra.Command {
				return associated
			},
		}

		text := `Associated Commands:{{range associated}}{{if .IsAvailableCommand}}
  {{pad .CommandPath}} {{.Short}}{{end}}{{end}}
`

		t := template.New("top")
		t.Funcs(templateFuncs)
		template.Must(t.Parse(text))
		return t.Execute(cmd.OutOrStderr(), cmd)
	})
	return root
}

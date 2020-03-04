package cmd

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfspretty "github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/tabwriter"
	ppspretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"

	"github.com/spf13/cobra"
)

func applyCommandCompat1_8(rootCmd *cobra.Command) {
	var commands []*cobra.Command

	// Command.Find matches args as well as command names, so implement our own
	findCommand := func(fullName string) *cobra.Command {
		cursor := rootCmd
		for _, name := range strings.SplitN(fullName, " ", -1) {
			var next *cobra.Command
			for _, cmd := range cursor.Commands() {
				if cmd.Name() == name {
					next = cmd
				}
			}
			cursor = next
			if cursor == nil {
				panic(fmt.Sprintf("Could not find '%s' command to apply v1.8 compatibility\n", fullName))
			}
		}
		return cursor
	}

	// These commands are backwards compatible aside from the reorganization, just
	// add new aliases.
	simpleCompat := map[string]string{
		"create repo":              "create-repo",
		"update repo":              "update-repo",
		"inspect repo":             "inspect-repo",
		"list repo":                "list-repo",
		"delete repo":              "delete-repo",
		"list branch":              "list-branch",
		"get object":               "get-object",
		"get tag":                  "get-tag",
		"inspect job":              "inspect-job",
		"delete job":               "delete-job",
		"stop job":                 "stop-job",
		"restart datum":            "restart-datum",
		"list datum":               "list-datum",
		"inspect datum":            "inspect-datum",
		"logs":                     "get-logs",
		"create pipeline":          "create-pipeline",
		"update pipeline":          "update-pipeline",
		"inspect pipeline":         "inspect-pipeline",
		"extract pipeline":         "extract-pipeline",
		"edit pipeline":            "edit-pipeline",
		"list pipeline":            "list-pipeline",
		"delete pipeline":          "delete-pipeline",
		"start pipeline":           "start-pipeline",
		"stop pipeline":            "stop-pipeline",
		"inspect cluster":          "inspect-cluster",
		"debug dump":               "debug-dump",
		"debug profile":            "debug-profile",
		"debug binary":             "debug-binary",
		"debug pprof":              "debug-pprof",
		"delete all":               "delete-all",
		"deploy storage amazon":    "deploy storage aws",
		"deploy storage microsoft": "deploy storage azure",
	}

	for newName, oldName := range simpleCompat {
		compatCmd := &cobra.Command{}
		*compatCmd = *findCommand(newName)

		useSplit := strings.SplitN(compatCmd.Use, " ", 2)
		if len(useSplit) == 2 {
			compatCmd.Use = fmt.Sprintf("{{alias}} %s", useSplit[1])
		} else {
			compatCmd.Use = "{{alias}}"
		}

		commands = append(commands, cmdutil.CreateAlias(compatCmd, oldName))
	}

	// Helper types for organizing more complicated command compatibility
	type RunFunc func(*cobra.Command, []string)
	type CompatChanges struct {
		Use     string
		Example string
		Run     func(*cobra.Command, RunFunc) RunFunc
	}

	// These helper functions will transform positional command-line args and
	// pass-through to the new implementations of the command so we can maintain
	// a single code path.
	transformRepoBranch := func(cmd *cobra.Command, newRun RunFunc) func([]string) error {
		return func(args []string) error {
			var newArgs []string
			var repo string

			for i, arg := range args {
				switch i % 2 {
				case 0:
					repo = arg
				case 1:
					newArgs = append(newArgs, fmt.Sprintf("%s@%s", repo, arg))
					repo = ""
				}
			}

			if repo != "" {
				newArgs = append(newArgs, repo)
			}

			newRun(cmd, newArgs)
			return nil
		}
	}

	transformRepoBranchFile := func(cmd *cobra.Command, newRun RunFunc) func([]string) error {
		return func(args []string) error {
			var newArgs []string
			var repoBranch string

			for i, arg := range args {
				switch i % 3 {
				case 0:
					repoBranch = arg
				case 1:
					repoBranch = fmt.Sprintf("%s@%s", repoBranch, arg)
				case 2:
					newArgs = append(newArgs, fmt.Sprintf("%s:%s", repoBranch, arg))
					repoBranch = ""
				}
			}

			if repoBranch != "" {
				newArgs = append(newArgs, repoBranch)
			}

			newRun(cmd, newArgs)
			return nil
		}
	}

	// A few places pre-v1.9 accepted arguments of the form repo/branch -
	// post-v1.9, we now expect repo@branch
	transformRepoSlashBranch := func(cmd *cobra.Command, newRun RunFunc) func([]string) error {
		return func(args []string) error {
			var newArgs []string

			for _, arg := range args {
				newArgs = append(newArgs, strings.Replace(arg, "/", "@", 1))
			}

			newRun(cmd, newArgs)
			return nil
		}
	}

	// These commands require transforming their old parameter format to the new
	// format, as well as maintaining their old Use strings.
	complexCompat := map[string]CompatChanges{
		"start commit": {
			Use: "start-commit <repo> [<branch-or-commit>]",
			Example: `
# Start a new commit in repo "test" that's not on any branch
$ pachctl start-commit test

# Start a commit in repo "test" on branch "master"
$ pachctl start-commit test master

# Start a commit with "master" as the parent in repo "test", on a new branch "patch"; essentially a fork.
$ pachctl start-commit test patch -p master

# Start a commit with XXX as the parent in repo "test", not on any branch
$ pachctl start-commit test -p XXX
			`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunBoundedArgs(1, 2, transformRepoBranch(cmd, newRun))
			},
		},

		"finish commit": {
			Use: "finish-commit <repo> <branch-or-commit>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(2, transformRepoBranch(cmd, newRun))
			},
		},

		"subscribe commit": {
			Use: "subscribe-commit <repo> <branch>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(2, transformRepoBranch(cmd, newRun))
			},
		},

		"list commit": {
			Use: "list-commit <repo> [<branch>]",
			Example: `
# return commits in repo "foo"
$ {{alias}} foo

# return commits in repo "foo" on branch "master"
$ {{alias}} foo master

# return the last 20 commits in repo "foo" on branch "master"
$ {{alias}} foo master -n 20

# return commits in repo "foo" since commit XXX
$ {{alias}} foo master --from XXX`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunBoundedArgs(1, 2, transformRepoBranch(cmd, newRun))
			},
		},

		"delete commit": {
			Use: "delete-commit <repo> <commit>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(2, transformRepoBranch(cmd, newRun))
			},
		},

		"delete branch": {
			Use: "delete-branch <repo> <branch>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(2, transformRepoBranch(cmd, newRun))
			},
		},

		"flush job": {
			Use: "flush-job <repo>/<commit> ...",
			Example: `
# return jobs caused by foo/XXX and bar/YYY
$ pachctl flush-job foo/XXX bar/YYY

# return jobs caused by foo/XXX leading to pipelines bar and baz
$ pachctl flush-job foo/XXX -p bar -p baz`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.Run(transformRepoSlashBranch(cmd, newRun))
			},
		},

		"flush commit": {
			Use: "flush-commit <repo>/<commit> ...",
			Example: `
# return commits caused by foo/XXX and bar/YYY
$ pachctl flush-commit foo/XXX bar/YYY

# return commits caused by foo/XXX leading to repos bar and baz
$ pachctl flush-commit foo/XXX -r bar -r baz`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.Run(transformRepoSlashBranch(cmd, newRun))
			},
		},

		"put file": {
			Use: "put-file repo-name branch [path/to/file/in/pfs]",
			Example: `
# Put data from stdin as repo/branch/path:
$ echo "data" | pachctl put-file repo branch path

# Put data from stdin as repo/branch/path and start / finish a new commit on the branch.
$ echo "data" | pachctl put-file -c repo branch path

# Put a file from the local filesystem as repo/branch/path:
$ pachctl put-file repo branch path -f file

# Put a file from the local filesystem as repo/branch/file:
$ pachctl put-file repo branch -f file

# Put the contents of a directory as repo/branch/path/dir/file:
$ pachctl put-file -r repo branch path -f dir

# Put the contents of a directory as repo/branch/dir/file:
$ pachctl put-file -r repo branch -f dir

# Put the contents of a directory as repo/branch/file, i.e. put files at the top level:
$ pachctl put-file -r repo branch / -f dir

# Put the data from a URL as repo/branch/path:
$ pachctl put-file repo branch path -f http://host/path

# Put the data from a URL as repo/branch/path:
$ pachctl put-file repo branch -f http://host/path

# Put the data from an S3 bucket as repo/branch/s3_object:
$ pachctl put-file repo branch -r -f s3://my_bucket

# Put several files or URLs that are listed in file.
# Files and URLs should be newline delimited.
$ pachctl put-file repo branch -i file

# Put several files or URLs that are listed at URL.
# NOTE this URL can reference local files, so it could cause you to put sensitive
# files into your Pachyderm cluster.
$ pachctl put-file repo branch -i http://host/path`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunBoundedArgs(2, 3, transformRepoBranchFile(cmd, newRun))
			},
		},

		"get file": {
			Use: "get-file <repo> <commit> <path/to/file>",
			Example: `
# get file "XXX" on branch "master" in repo "foo"
$ pachctl get-file foo master XXX

# get file "XXX" in the parent of the current head of branch "master"
# in repo "foo"
$ pachctl get-file foo master^ XXX

# get file "XXX" in the grandparent of the current head of branch "master"
# in repo "foo"
$ pachctl get-file foo master^2 XXX`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(3, transformRepoBranchFile(cmd, newRun))
			},
		},

		"inspect file": {
			Use: "inspect-file <repo> <commit> <path/to/file>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(3, transformRepoBranchFile(cmd, newRun))
			},
		},

		"list file": {
			Use: "list-file repo-name commit-id path/to/dir",
			Example: `
# list top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master

# list files under directory "dir" on branch "master" in repo "foo"
$ pachctl list-file foo master dir

# list top-level files in the parent commit of the current head of "master"
# in repo "foo"
$ pachctl list-file foo master^

# list top-level files in the grandparent of the current head of "master"
# in repo "foo"
$ pachctl list-file foo master^2

# list the last n versions of top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master --history n

# list all versions of top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master --history -1`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunBoundedArgs(2, 3, transformRepoBranchFile(cmd, newRun))
			},
		},

		"glob file": {
			Use: "glob-file <repo> <commit> <pattern>",
			Example: `
# Return files in repo "foo" on branch "master" that start
# with the character "A".  Note how the double quotation marks around "A*" are
# necessary because otherwise your shell might interpret the "*".
$ pachctl glob-file foo master "A*"

# Return files in repo "foo" on branch "master" under directory "data".
$ pachctl glob-file foo master "data/*"`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(3, transformRepoBranchFile(cmd, newRun))
			},
		},

		"delete file": {
			Use: "delete-file <repo> <commit> <path/to/file>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(3, transformRepoBranchFile(cmd, newRun))
			},
		},

		"copy file": {
			Use: "copy-file <src-repo> <src-commit> <src-path> <dst-repo> <dst-commit> <dst-path>",
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunFixedArgs(6, transformRepoBranchFile(cmd, newRun))
			},
		},

		"diff file": {
			Use: "diff-file <new-repo> <new-commit> <new-path> [<old-repo> <old-commit> <old-path>]",
			Example: `
# Return the diff between foo master path and its parent.
$ pachctl diff-file foo master path

# Return the diff between foo master path1 and bar master path2.
$ pachctl diff-file foo master path1 bar master path2`,
			Run: func(cmd *cobra.Command, newRun RunFunc) RunFunc {
				return cmdutil.RunBoundedArgs(3, 6, transformRepoBranchFile(cmd, newRun))
			},
		},
	}

	for newName, changes := range complexCompat {
		newCmd := findCommand(newName)
		oldCmd := &cobra.Command{}
		*oldCmd = *newCmd
		oldCmd.Use = changes.Use
		oldCmd.Example = changes.Example
		oldCmd.Run = changes.Run(newCmd, newCmd.Run)
		commands = append(commands, oldCmd)
	}

	// ParseCommits takes a slice of arguments of the form "repo/commit-id" or
	// "repo" (in which case we consider the commit ID to be empty), and returns
	// a list of *pfs.Commits
	oldParseCommits := func(args []string) ([]*pfs.Commit, error) {
		var commits []*pfs.Commit
		for _, arg := range args {
			parts := strings.SplitN(arg, "/", 2)
			hasRepo := len(parts) > 0 && parts[0] != ""
			hasCommit := len(parts) == 2 && parts[1] != ""
			if hasCommit && !hasRepo {
				return nil, errors.Errorf("invalid commit id \"%s\": repo cannot be empty", arg)
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
	oldParseBranches := func(args []string) ([]*pfs.Branch, error) {
		commits, err := oldParseCommits(args)
		if err != nil {
			return nil, err
		}
		var result []*pfs.Branch
		for _, commit := range commits {
			result = append(result, &pfs.Branch{Repo: commit.Repo, Name: commit.ID})
		}
		return result, nil
	}

	// create-branch has affected flags, so just duplicate the command entirely
	var branchProvenance cmdutil.RepeatedStringArg
	var head string
	newCreateBranch := findCommand("create branch")
	oldCreateBranch := &cobra.Command{
		Use:   "create-branch <repo> <branch>",
		Short: newCreateBranch.Short,
		Long:  newCreateBranch.Long,
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			provenance, err := oldParseBranches(branchProvenance)
			if err != nil {
				return err
			}
			return client.CreateBranch(args[0], args[1], head, provenance)
		}),
	}
	oldCreateBranch.Flags().VarP(&branchProvenance, "provenance", "p", "The provenance for the branch.")
	oldCreateBranch.Flags().StringVarP(&head, "head", "", "", "The head of the newly created branch.")
	commands = append(commands, oldCreateBranch)

	// list-job has affected flags, so just duplicate the command entirely
	var raw bool
	var fullTimestamps bool
	var pipelineName string
	var outputCommitStr string
	var inputCommitStrs []string
	newListJob := findCommand("list job")
	oldListJob := &cobra.Command{
		Use:   "list-job",
		Short: newListJob.Short,
		Long:  newListJob.Long,
		Example: `
# return all jobs
$ pachctl list-job

# return all jobs in pipeline foo
$ pachctl list-job -p foo

# return all jobs whose input commits include foo/XXX and bar/YYY
$ pachctl list-job foo/XXX bar/YYY

# return all jobs in pipeline foo and whose input commits include bar/YYY
$ pachctl list-job -p foo bar/YYY`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()

			commits, err := oldParseCommits(inputCommitStrs)
			if err != nil {
				return err
			}

			var outputCommit *pfs.Commit
			if outputCommitStr != "" {
				outputCommits, err := oldParseCommits([]string{outputCommitStr})
				if err != nil {
					return err
				}
				if len(outputCommits) == 1 {
					outputCommit = outputCommits[0]
				}
			}

			if raw {
				return client.ListJobF(pipelineName, commits, outputCommit, -1, true,
					func(ji *pps.JobInfo) error {
						marshaller := &jsonpb.Marshaler{Indent: "  "}
						if err := marshaller.Marshal(os.Stdout, ji); err != nil {
							return err
						}
						return nil
					})
			}
			writer := tabwriter.NewWriter(os.Stdout, ppspretty.JobHeader)
			if err := client.ListJobF(pipelineName, commits, outputCommit, -1, false,
				func(ji *pps.JobInfo) error {
					ppspretty.PrintJobInfo(writer, ji, fullTimestamps)
					return nil
				}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	oldListJob.Flags().StringVarP(&pipelineName, "pipeline", "p", "", "Limit to jobs made by pipeline.")
	oldListJob.Flags().StringVarP(&outputCommitStr, "output", "o", "", "List jobs with a specific output commit.")
	oldListJob.Flags().StringSliceVarP(&inputCommitStrs, "input", "i", []string{}, "List jobs with a specific set of input commits.")
	oldListJob.Flags().BoolVar(&fullTimestamps, "full-timestamps", false, "Return absolute timestamps (as opposed to the default, relative timestamps).")
	oldListJob.Flags().BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")
	commands = append(commands, oldListJob)

	oldPrintDetailedCommitInfo := func(commitInfo *pfspretty.PrintableCommitInfo) error {
		var funcMap = template.FuncMap{
			"prettyAgo":  pretty.Ago,
			"prettySize": pretty.Size,
		}

		template, err := template.New("CommitInfo").Funcs(funcMap).Parse(
			`Commit: {{.Commit.Repo.Name}}/{{.Commit.ID}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .ParentCommit}}
Parent: {{.ParentCommit.ID}}{{end}}{{if .FullTimestamps}}
Started: {{.Started}}{{else}}
Started: {{prettyAgo .Started}}{{end}}{{if .Finished}}{{if .FullTimestamps}}
Finished: {{.Finished}}{{else}}
Finished: {{prettyAgo .Finished}}{{end}}{{end}}
Size: {{prettySize .SizeBytes}}{{if .Provenance}}
Provenance: {{range .Provenance}} {{.Repo.Name}}/{{.ID}}{{end}}{{end}}
`)
		if err != nil {
			return err
		}
		err = template.Execute(os.Stdout, commitInfo)
		if err != nil {
			return err
		}
		return nil
	}

	// inspect-commit had its output changed to use 'repo@commit' instead of
	// 'repo/commit' - reimplement it here to use the old pretty formatter.
	newInspectCommit := findCommand("inspect commit")
	oldInspectCommit := &cobra.Command{
		Use:   "inspect-commit <repo> <commit>",
		Short: newInspectCommit.Short,
		Long:  newInspectCommit.Long,
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer client.Close()
			commitInfo, err := client.InspectCommit(args[0], args[1])
			if err != nil {
				return err
			}
			if commitInfo == nil {
				return errors.Errorf("commit %s not found", args[1])
			}
			if raw {
				marshaller := &jsonpb.Marshaler{Indent: "  "}
				return marshaller.Marshal(os.Stdout, commitInfo)
			}
			ci := &pfspretty.PrintableCommitInfo{
				CommitInfo:     commitInfo,
				FullTimestamps: fullTimestamps,
			}
			return oldPrintDetailedCommitInfo(ci)
		}),
	}
	oldInspectCommit.Flags().BoolVar(&fullTimestamps, "full-timestamps", false, "Return absolute timestamps (as opposed to the default, relative timestamps).")
	oldInspectCommit.Flags().BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")
	commands = append(commands, oldInspectCommit)

	// Apply the 'Hidden' attribute to all these commands so they don't pollute help
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Run != nil {
			cmd.Hidden = true

			// Dear god
			oldPreRun := cmd.PreRun
			oldPreRunE := cmd.PreRunE
			cmd.PreRun = nil
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				oldPath := cmd.CommandPath()

				// Special case for a couple commands that don't match the pattern
				var newPath string
				if oldPath == fmt.Sprintf("%s deploy storage aws", os.Args[0]) {
					newPath = fmt.Sprintf("%s deploy storage amazon", os.Args[0])
				} else if oldPath == fmt.Sprintf("%s deploy storage azure", os.Args[0]) {
					newPath = fmt.Sprintf("%s deploy storage microsoft", os.Args[0])
				} else if oldPath == fmt.Sprintf("%s get-logs", os.Args[0]) {
					newPath = fmt.Sprintf("%s logs", os.Args[0])
				} else {
					newPath = strings.Replace(oldPath, "-", " ", -1)
				}

				fmt.Fprintf(os.Stderr, "WARNING: '%s' is deprecated and will be removed in a future release, use '%s' instead.\n", oldPath, newPath)
				if oldPreRunE != nil {
					return oldPreRunE(cmd, args)
				}
				if oldPreRun != nil {
					oldPreRun(cmd, args)
				}
				return nil
			}
		}

		for _, subcmd := range cmd.Commands() {
			walk(subcmd)
		}
	}

	for _, cmd := range commands {
		walk(cmd)
	}

	cmdutil.MergeCommands(rootCmd, commands)
}

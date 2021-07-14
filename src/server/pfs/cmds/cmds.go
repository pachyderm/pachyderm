package cmds

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pager"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/shell"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	txncmds "github.com/pachyderm/pachyderm/v2/src/server/transaction/cmds"
)

const (
	// DefaultParallelism is the default parallelism used by 'get file' and 'put file'.
	DefaultParallelism = 10
)

// Cmds returns a slice containing pfs commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	var fullTimestamps bool
	timestampFlags := cmdutil.TimestampFlags(&fullTimestamps)

	var noPager bool
	pagerFlags := cmdutil.PagerFlags(&noPager)

	repoDocs := &cobra.Command{
		Short: "Docs for repos.",
		Long: `Repos, short for repository, are the top level data objects in Pachyderm.

Repos contain version-controlled directories and files. Files can be of any size
or type (e.g. csv, binary, images, etc).`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(repoDocs, "repo", " repo$"))

	var description string
	createRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			err := txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				_, err := c.PfsAPIClient.CreateRepo(
					c.Ctx(),
					&pfs.CreateRepoRequest{
						Repo:        client.NewRepo(args[0]),
						Description: description,
					},
				)
				return err
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createRepo.Flags().StringVarP(&description, "description", "d", "", "A description of the repo.")
	commands = append(commands, cmdutil.CreateAlias(createRepo, "create repo"))

	updateRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Update a repo.",
		Long:  "Update a repo.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			err := txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				_, err := c.PfsAPIClient.CreateRepo(
					c.Ctx(),
					&pfs.CreateRepoRequest{
						Repo:        cmdutil.ParseRepo(args[0]),
						Description: description,
						Update:      true,
					},
				)
				return err
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateRepo.Flags().StringVarP(&description, "description", "d", "", "A description of the repo.")
	shell.RegisterCompletionFunc(updateRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAlias(updateRepo, "update repo"))

	inspectRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.Client("user")
			repoInfo, err := c.PfsAPIClient.InspectRepo(c.Ctx(), &pfs.InspectRepoRequest{Repo: cmdutil.ParseRepo(args[0])})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if repoInfo == nil {
				return errors.Errorf("repo %s not found", args[0])
			}
			if raw {
				return cmdutil.Encoder(output, env.Stdout()).EncodeProto(repoInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			ri := &pretty.PrintableRepoInfo{
				RepoInfo:       repoInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedRepoInfo(env.Stdout(), ri)
		}),
	}
	inspectRepo.Flags().AddFlagSet(outputFlags)
	inspectRepo.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(inspectRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectRepo, "inspect repo"))

	var all bool
	var repoType string
	listRepo := &cobra.Command{
		Short: "Return a list of repos.",
		Long:  "Return a list of repos. By default, only show user repos",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			if all && repoType != "" {
				return errors.Errorf("cannot set a repo type with --all")
			}

			if repoType == "" && !all {
				repoType = pfs.UserRepoType // default to user
			}
			repoInfos, err := env.Client("user").ListRepoByType(repoType)
			if err != nil {
				return err
			}
			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				for _, repoInfo := range repoInfos {
					if err := encoder.EncodeProto(repoInfo); err != nil {
						return err
					}
				}
				return nil
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			header := pretty.RepoHeader
			if (len(repoInfos) > 0) && (repoInfos[0].AuthInfo != nil) {
				header = pretty.RepoAuthHeader
			}
			writer := tabwriter.NewWriter(env.Stdout(), header)
			for _, repoInfo := range repoInfos {
				pretty.PrintRepoInfo(writer, repoInfo, fullTimestamps)
			}
			return writer.Flush()
		}),
	}
	listRepo.Flags().AddFlagSet(outputFlags)
	listRepo.Flags().AddFlagSet(timestampFlags)
	listRepo.Flags().BoolVar(&all, "all", false, "include system repos of all types")
	listRepo.Flags().StringVar(&repoType, "type", "", "only include repos of the given type")
	commands = append(commands, cmdutil.CreateAlias(listRepo, "list repo"))

	var force bool
	deleteRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		RunE: cmdutil.RunBoundedArgs(0, 1, func(args []string, env cmdutil.Env) error {
			request := &pfs.DeleteRepoRequest{
				Force: force,
			}
			if len(args) > 0 {
				if all {
					return errors.Errorf("cannot use the --all flag with an argument")
				}
				request.Repo = cmdutil.ParseRepo(args[0])
			} else if !all {
				return errors.Errorf("either a repo name or the --all flag needs to be provided")
			}

			err := txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				var err error
				if all {
					_, err = c.PfsAPIClient.DeleteAll(c.Ctx(), &types.Empty{})
				} else {
					_, err = c.PfsAPIClient.DeleteRepo(c.Ctx(), request)
				}
				return err
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	deleteRepo.Flags().BoolVarP(&force, "force", "f", false, "remove the repo regardless of errors; use with care")
	deleteRepo.Flags().BoolVar(&all, "all", false, "remove all repos")
	shell.RegisterCompletionFunc(deleteRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAlias(deleteRepo, "delete repo"))

	commitDocs := &cobra.Command{
		Short: "Docs for commits.",
		Long: `Commits are atomic transactions on the content of a repo.

Creating a commit is a multistep process:
- start a new commit with 'start commit'
- write files to the commit via 'put file'
- finish the new commit with 'finish commit'

Commits that have been started but not finished are NOT durable storage.
Commits become reliable (and immutable) when they are finished.

Commits can be created with another commit as a parent.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(commitDocs, "commit", " commit$"))

	var parent string
	startCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Start a new commit.",
		Long:  "Start a new commit with parent-commit as the parent, or start a commit on the given branch; if the branch does not exist, it will be created.",
		Example: `# Start a new commit in repo "test" that's not on any branch
$ {{alias}} test

# Start a commit in repo "test" on branch "master"
$ {{alias}} test@master

# Start a commit with "master" as the parent in repo "test", on a new branch "patch"; essentially a fork.
$ {{alias}} test@patch -p master

# Start a commit with XXX as the parent in repo "test", not on any branch
$ {{alias}} test -p XXX`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}

			var parentCommit *pfs.Commit
			if parent != "" {
				// We don't know if the parent is a commit ID, branch, or ancestry, so
				// construct a string to parse.
				parentCommit, err = cmdutil.ParseCommit(fmt.Sprintf("%s@%s", branch.Repo, parent))
				if err != nil {
					return err
				}
			}

			var commit *pfs.Commit
			err = txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				var err error
				commit, err = c.PfsAPIClient.StartCommit(
					c.Ctx(),
					&pfs.StartCommitRequest{
						Branch:      branch,
						Parent:      parentCommit,
						Description: description,
					},
				)
				return err
			})
			if err == nil {
				fmt.Fprintln(env.Stdout(), commit.ID)
			}
			return grpcutil.ScrubGRPC(err)
		}),
	}
	startCommit.Flags().StringVarP(&parent, "parent", "p", "", "The parent of the new commit, unneeded if branch is specified and you want to use the previous head of the branch as the parent.")
	startCommit.MarkFlagCustom("parent", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	startCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents")
	startCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")
	shell.RegisterCompletionFunc(startCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(startCommit, "start commit"))

	finishCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			commit, err := cmdutil.ParseCommit(args[0])
			if err != nil {
				return err
			}

			err = txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				_, err = c.PfsAPIClient.FinishCommit(
					c.Ctx(),
					&pfs.FinishCommitRequest{
						Commit:      commit,
						Description: description,
						Force:       force,
					},
				)
				return err
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	finishCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents (overwrites any existing commit description)")
	finishCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")
	finishCommit.Flags().BoolVarP(&force, "force", "f", false, "finish the commit even if it has provenance, which could break jobs; prefer 'stop job'")
	shell.RegisterCompletionFunc(finishCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(finishCommit, "finish commit"))

	inspectCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			commit, err := cmdutil.ParseCommit(args[0])
			if err != nil {
				return err
			}

			c := env.Client("user")
			commitInfo, err := c.PfsAPIClient.InspectCommit(
				c.Ctx(),
				&pfs.InspectCommitRequest{
					Commit: commit,
					Wait:   pfs.CommitState_STARTED,
				})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if commitInfo == nil {
				return errors.Errorf("commit %s not found", commit.ID)
			}
			if raw {
				return cmdutil.Encoder(output, env.Stdout()).EncodeProto(commitInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			ci := &pretty.PrintableCommitInfo{
				CommitInfo:     commitInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedCommitInfo(env.Stdout(), ci)
		}),
	}
	inspectCommit.Flags().AddFlagSet(outputFlags)
	inspectCommit.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(inspectCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectCommit, "inspect commit"))

	var from string
	var number int64
	var originStr string
	listCommit := &cobra.Command{
		Use:   "{{alias}} <repo>[@<branch>]",
		Short: "Return all commits on a repo.",
		Long:  "Return all commits on a repo.",
		Example: `
# return commits in repo "foo"
$ {{alias}} foo

# return commits in repo "foo" on branch "master"
$ {{alias}} foo@master

# return the last 20 commits in repo "foo" on branch "master"
$ {{alias}} foo@master -n 20

# return commits in repo "foo" since commit XXX
$ {{alias}} foo@master --from XXX`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) (retErr error) {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}

			var fromCommit *pfs.Commit
			if from != "" {
				fromCommit = branch.Repo.NewCommit("", from)
			}

			var toCommit *pfs.Commit
			if branch.Name != "" {
				toCommit = branch.NewCommit("")
			}

			if all && originStr != "" {
				return errors.New("cannot specify both --all and --origin")
			}

			origin, err := parseOriginKind(originStr)
			if err != nil {
				return err
			}

			c := env.Client("user")
			listClient, err := c.PfsAPIClient.ListCommit(c.Ctx(), &pfs.ListCommitRequest{
				Repo:       branch.Repo,
				From:       fromCommit,
				To:         toCommit,
				Number:     number,
				All:        all,
				OriginKind: origin,
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				return clientsdk.ForEachCommit(listClient, func(ci *pfs.CommitInfo) error {
					return encoder.EncodeProto(ci)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			writer := tabwriter.NewWriter(env.Stdout(), pretty.CommitHeader)
			defer func() {
				if err := writer.Flush(); retErr == nil {
					retErr = err
				}
			}()
			return clientsdk.ForEachCommit(listClient, func(ci *pfs.CommitInfo) error {
				pretty.PrintCommitInfo(writer, ci, fullTimestamps)
				return nil
			})
		}),
	}
	listCommit.Flags().StringVarP(&from, "from", "f", "", "list all commits since this commit")
	listCommit.Flags().Int64VarP(&number, "number", "n", 0, "list only this many commits; if set to zero, list all commits")
	listCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	listCommit.Flags().BoolVar(&all, "all", false, "return all types of commits, including aliases")
	listCommit.Flags().StringVar(&originStr, "origin", "", "only return commits of a specific type")
	listCommit.Flags().AddFlagSet(outputFlags)
	listCommit.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(listCommit, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAlias(listCommit, "list commit"))

	var branches cmdutil.RepeatedStringArg
	waitCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Wait for the specified commit to finish and return it.",
		Long:  "Wait for the specified commit to finish and return it.",
		Example: `
# wait for the commit foo@XXX to finish and return it
$ {{alias}} foo@XXX -b bar@baz`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			commit, err := cmdutil.ParseCommit(args[0])
			if err != nil {
				return err
			}

			commitInfo, err := env.Client("user").WaitCommit(commit.Branch.Repo.Name, commit.Branch.Name, commit.ID)
			if err != nil {
				return err
			}

			if raw {
				return cmdutil.Encoder(output, env.Stdout()).EncodeProto(commitInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			ci := &pretty.PrintableCommitInfo{
				CommitInfo:     commitInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedCommitInfo(env.Stdout(), ci)
		}),
	}
	waitCommit.Flags().AddFlagSet(outputFlags)
	waitCommit.Flags().AddFlagSet(timestampFlags)
	commands = append(commands, cmdutil.CreateAlias(waitCommit, "wait commit"))

	var newCommits bool
	subscribeCommit := &cobra.Command{
		Use:   "{{alias}} <repo>[@<branch>]",
		Short: "Print commits as they are created (finished).",
		Long:  "Print commits as they are created in the specified repo and branch.  By default, all existing commits on the specified branch are returned first.  A commit is only considered 'created' when it's been finished.",
		Example: `
# subscribe to commits in repo "test" on branch "master"
$ {{alias}} test@master

# subscribe to commits in repo "test" on branch "master", but only since commit XXX.
$ {{alias}} test@master --from XXX

# subscribe to commits in repo "test" on branch "master", but only for new commits created from now on.
$ {{alias}} test@master --new`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) (retErr error) {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}

			var fromCommit *pfs.Commit
			if newCommits && from != "" {
				return errors.Errorf("--new and --from cannot be used together")
			} else if newCommits || from != "" {
				fromCommit = branch.NewCommit(from)
			}

			if all && originStr != "" {
				return errors.New("cannot specify both --all and --origin")
			}

			origin, err := parseOriginKind(originStr)
			if err != nil {
				return err
			}

			c := env.Client("user")
			subscribeClient, err := c.PfsAPIClient.SubscribeCommit(c.Ctx(), &pfs.SubscribeCommitRequest{
				Repo:       branch.Repo,
				Branch:     branch.Name,
				From:       fromCommit,
				State:      pfs.CommitState_STARTED,
				All:        all,
				OriginKind: origin,
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				return clientsdk.ForEachSubscribeCommit(subscribeClient, func(ci *pfs.CommitInfo) error {
					return encoder.EncodeProto(ci)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			writer := tabwriter.NewWriter(env.Stdout(), pretty.CommitHeader)
			defer func() {
				if err := writer.Flush(); retErr == nil {
					retErr = err
				}
			}()

			return clientsdk.ForEachSubscribeCommit(subscribeClient, func(ci *pfs.CommitInfo) error {
				pretty.PrintCommitInfo(writer, ci, fullTimestamps)
				return nil
			})
		}),
	}
	subscribeCommit.Flags().StringVar(&from, "from", "", "subscribe to all commits since this commit")
	subscribeCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	subscribeCommit.Flags().BoolVar(&newCommits, "new", false, "subscribe to only new commits created from now on")
	subscribeCommit.Flags().BoolVar(&all, "all", false, "return all types of commits, including aliases")
	subscribeCommit.Flags().StringVar(&originStr, "origin", "", "only return commits of a specific type")
	subscribeCommit.Flags().AddFlagSet(outputFlags)
	subscribeCommit.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(subscribeCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(subscribeCommit, "subscribe commit"))

	writeCommitTable := func(env cmdutil.Env, commitInfos []*pfs.CommitInfo) error {
		if raw {
			encoder := cmdutil.Encoder(output, env.Stdout())
			for _, commitInfo := range commitInfos {
				if err := encoder.EncodeProto(commitInfo); err != nil {
					return err
				}
			}
			return nil
		} else if output != "" {
			return errors.New("cannot set --output (-o) without --raw")
		}

		return pager.Page(noPager, env, func(w io.Writer) error {
			writer := tabwriter.NewWriter(w, pretty.CommitHeader)
			for _, commitInfo := range commitInfos {
				pretty.PrintCommitInfo(writer, commitInfo, fullTimestamps)
			}
			return writer.Flush()
		})
	}

	waitCommitSet := &cobra.Command{
		Use:   "{{alias}} <commitset-id>",
		Short: "Wait for commits in a commitset to finish and return them.",
		Long:  "Wait for commits in a commitset to finish and return them.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			toBranches := []*pfs.Branch{}
			for _, arg := range branches {
				branch, err := cmdutil.ParseBranch(arg)
				if err != nil {
					return err
				}
				toBranches = append(toBranches, branch)
			}

			var commitInfos []*pfs.CommitInfo
			if len(toBranches) != 0 {
				for _, branch := range toBranches {
					ci, err := env.Client("user").WaitCommit(branch.Repo.Name, branch.Name, args[0])
					if err != nil {
						return errors.Wrap(err, "error from InspectCommit")
					}
					commitInfos = append(commitInfos, ci)
				}
			} else {
				var err error
				commitInfos, err = env.Client("user").WaitCommitSetAll(args[0])
				if err != nil {
					return errors.Wrap(err, "error from InspectCommitSet")
				}
			}
			return writeCommitTable(env, commitInfos)
		}),
	}
	waitCommitSet.Flags().VarP(&branches, "branch", "b", "Wait only for commits in the specified set of branches")
	waitCommitSet.MarkFlagCustom("branch", "__pachctl_get_branch")
	waitCommitSet.Flags().AddFlagSet(outputFlags)
	waitCommitSet.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(waitCommitSet, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(waitCommitSet, "wait commitset"))

	inspectCommitSet := &cobra.Command{
		Use:   "{{alias}} <job-id>",
		Short: "Return info about all the commits in a commitset.",
		Long:  "Return info about all the commits in a commitset.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			commitInfos, err := env.Client("user").InspectCommitSet(args[0])
			if err != nil {
				return errors.Wrap(err, "error from InspectCommitSet")
			}
			return writeCommitTable(env, commitInfos)
		}),
	}
	inspectCommitSet.Flags().AddFlagSet(outputFlags)
	inspectCommitSet.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(inspectCommitSet, shell.JobCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectCommitSet, "inspect commitset"))

	listCommitSet := &cobra.Command{
		Short: "Return info about commitsets.",
		Long:  "Return info about commitsets.",
		Example: `
# Return all commitsets
$ {{alias}}`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.Client("user")
			listCommitSetClient, err := c.PfsAPIClient.ListCommitSet(c.Ctx(), &pfs.ListCommitSetRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				return clientsdk.ForEachCommitSet(listCommitSetClient, func(commitSetInfo *pfs.CommitSetInfo) error {
					return encoder.EncodeProto(commitSetInfo)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			return pager.Page(noPager, env, func(w io.Writer) (retErr error) {
				writer := tabwriter.NewWriter(w, pretty.CommitSetHeader)
				defer func() {
					if err := writer.Flush(); retErr == nil {
						retErr = err
					}
				}()

				return clientsdk.ForEachCommitSet(listCommitSetClient, func(commitSetInfo *pfs.CommitSetInfo) error {
					pretty.PrintCommitSetInfo(writer, commitSetInfo, fullTimestamps)
					return nil
				})
			})
		}),
	}
	listCommitSet.Flags().AddFlagSet(outputFlags)
	listCommitSet.Flags().AddFlagSet(timestampFlags)
	listCommitSet.Flags().AddFlagSet(pagerFlags)
	commands = append(commands, cmdutil.CreateAlias(listCommitSet, "list commitset"))

	squashCommitSet := &cobra.Command{
		Use:   "{{alias}} <commitset>",
		Short: "Squash the commits of a commitset.",
		Long:  "Squash the commits of a commitset.  The data in the commits will remain in their child commits unless there are no children.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			return txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				return c.SquashCommitSet(args[0])
			})
		}),
	}
	shell.RegisterCompletionFunc(squashCommitSet, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(squashCommitSet, "squash commitset"))

	branchDocs := &cobra.Command{
		Short: "Docs for branches.",
		Long: `A branch in Pachyderm is an alias for a Commit ID.

The branch reference will "float" to always refer to the latest commit on the
branch, known as the HEAD commit. Not all commits must be on a branch and
multiple branches can refer to the same commit.

Any pachctl command that can take a Commit ID, can take a branch name instead.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(branchDocs, "branch", " branch$"))

	var branchProvenance cmdutil.RepeatedStringArg
	var head string
	trigger := &pfs.Trigger{}
	createBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Create a new branch, or update an existing branch, on a repo.",
		Long:  "Create a new branch, or update an existing branch, on a repo, starting a commit on the branch will also create it, so there's often no need to call this.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}
			provenance, err := cmdutil.ParseBranches(branchProvenance)
			if err != nil {
				return err
			}
			if len(provenance) != 0 && trigger.Branch != "" {
				return errors.Errorf("cannot use provenance and triggers on the same branch")
			}
			if (trigger.CronSpec != "" || trigger.Size_ != "" || trigger.Commits != 0) && trigger.Branch == "" {
				return errors.Errorf("trigger condition specified without a branch to trigger on, specify a branch with --trigger")
			}
			if proto.Equal(trigger, &pfs.Trigger{}) {
				trigger = nil
			}
			var headCommit *pfs.Commit
			if head != "" {
				if strings.Contains(head, "@") {
					headCommit, err = cmdutil.ParseCommit(head)
					if err != nil {
						return err
					}
				} else {
					// treat head as the commitID or branch name
					headCommit = branch.Repo.NewCommit("", head)
				}
			}

			return txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				_, err := c.PfsAPIClient.CreateBranch(
					c.Ctx(),
					&pfs.CreateBranchRequest{
						Head:       headCommit,
						Branch:     branch,
						Provenance: provenance,
						Trigger:    trigger,
					})
				return grpcutil.ScrubGRPC(err)
			})
		}),
	}
	createBranch.Flags().VarP(&branchProvenance, "provenance", "p", "The provenance for the branch. format: <repo>@<branch-or-commit>")
	createBranch.MarkFlagCustom("provenance", "__pachctl_get_repo_commit")
	createBranch.Flags().StringVarP(&head, "head", "", "", "The head of the newly created branch. Either pass the commit with format: <branch-or-commit>, or fully-qualified as <repo>@<branch>=<id>")
	createBranch.MarkFlagCustom("head", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	createBranch.Flags().StringVarP(&trigger.Branch, "trigger", "t", "", "The branch to trigger this branch on.")
	createBranch.Flags().StringVar(&trigger.CronSpec, "trigger-cron", "", "The cron spec to use in triggering.")
	createBranch.Flags().StringVar(&trigger.Size_, "trigger-size", "", "The data size to use in triggering.")
	createBranch.Flags().Int64Var(&trigger.Commits, "trigger-commits", 0, "The number of commits to use in triggering.")
	createBranch.Flags().BoolVar(&trigger.All, "trigger-all", false, "Only trigger when all conditions are met, rather than when any are met.")
	commands = append(commands, cmdutil.CreateAlias(createBranch, "create branch"))

	inspectBranch := &cobra.Command{
		Use:   "{{alias}}  <repo>@<branch>",
		Short: "Return info about a branch.",
		Long:  "Return info about a branch.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}

			c := env.Client("user")
			branchInfo, err := c.PfsAPIClient.InspectBranch(c.Ctx(), &pfs.InspectBranchRequest{Branch: branch})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if branchInfo == nil {
				return errors.Errorf("branch %s not found", args[0])
			}
			if raw {
				return cmdutil.Encoder(output, env.Stdout()).EncodeProto(branchInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			return pretty.PrintDetailedBranchInfo(env.Stdout(), branchInfo)
		}),
	}
	inspectBranch.Flags().AddFlagSet(outputFlags)
	inspectBranch.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(inspectBranch, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectBranch, "inspect branch"))

	listBranch := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return all branches on a repo.",
		Long:  "Return all branches on a repo.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) (retErr error) {
			c := env.Client("user")
			branchClient, err := c.PfsAPIClient.ListBranch(c.Ctx(), &pfs.ListBranchRequest{Repo: cmdutil.ParseRepo(args[0])})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				return clientsdk.ForEachBranchInfo(branchClient, func(branch *pfs.BranchInfo) error {
					return encoder.EncodeProto(branch)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			writer := tabwriter.NewWriter(env.Stdout(), pretty.BranchHeader)
			defer func() {
				if err := writer.Flush(); retErr == nil {
					retErr = err
				}
			}()

			return clientsdk.ForEachBranchInfo(branchClient, func(branch *pfs.BranchInfo) error {
				pretty.PrintBranch(writer, branch)
				return nil
			})
		}),
	}
	listBranch.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(listBranch, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAlias(listBranch, "list branch"))

	deleteBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Delete a branch",
		Long:  "Delete a branch, while leaving the commits intact",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}

			return txncmds.WithActiveTransaction(env, func(c *client.APIClient) error {
				_, err := c.PfsAPIClient.DeleteBranch(c.Ctx(), &pfs.DeleteBranchRequest{Branch: branch, Force: force})
				return grpcutil.ScrubGRPC(err)
			})
		}),
	}
	deleteBranch.Flags().BoolVarP(&force, "force", "f", false, "remove the branch regardless of errors; use with care")
	shell.RegisterCompletionFunc(deleteBranch, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAlias(deleteBranch, "delete branch"))

	fileDocs := &cobra.Command{
		Short: "Docs for files.",
		Long: `Files are the lowest level data objects in Pachyderm.

Files can be of any type (e.g. csv, binary, images, etc) or size and can be
written to started (but not finished) commits with 'put file'. Files can be read
from commits with 'get file'.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(fileDocs, "file", " file$"))

	var filePaths []string
	var inputFile string
	var recursive bool
	var parallelism int
	var appendFile bool
	var compress bool
	var enableProgress bool
	var fullPath bool
	putFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>[:<path/to/file>]",
		Short: "Put a file into the filesystem.",
		Long:  "Put a file into the filesystem.  This command supports a number of ways to insert data into PFS.",
		Example: `
# Put data from stdin as repo/branch/path:
$ echo "data" | {{alias}} repo@branch:/path

# Put data from stdin as repo/branch/path and start / finish a new commit on the branch.
$ echo "data" | {{alias}} -c repo@branch:/path

# Put a file from the local filesystem as repo/branch/path:
$ {{alias}} repo@branch:/path -f file

# Put a file from the local filesystem as repo/branch/file:
$ {{alias}} repo@branch -f file

# Put the contents of a directory as repo/branch/path/dir/file:
$ {{alias}} -r repo@branch:/path -f dir

# Put the contents of a directory as repo/branch/dir/file:
$ {{alias}} -r repo@branch -f dir

# Put the contents of a directory as repo/branch/file, i.e. put files at the top level:
$ {{alias}} -r repo@branch:/ -f dir

# Put the data from a URL as repo/branch/path:
$ {{alias}} repo@branch:/path -f http://host/path

# Put the data from a URL as repo/branch/path:
$ {{alias}} repo@branch -f http://host/path

# Put the data from an S3 bucket as repo/branch/s3_object:
$ {{alias}} repo@branch -r -f s3://my_bucket

# Put several files or URLs that are listed in file.
# Files and URLs should be newline delimited.
$ {{alias}} repo@branch -i file

# Put several files or URLs that are listed at URL.
# NOTE this URL can reference local files, so it could cause you to put sensitive
# files into your Pachyderm cluster.
$ {{alias}} repo@branch -i http://host/path`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) (retErr error) {
			if !enableProgress {
				progress.Disable()
			}
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			opts := []client.Option{client.WithMaxConcurrentStreams(parallelism)}
			if compress {
				opts = append(opts, client.WithGZIPCompression())
			}
			defer progress.Wait()

			// TODO: Rethink put file parallelism for 2.0.
			// Doing parallel uploads at the file level for small files will be bad, but we still want a clear way to parallelize large file uploads.
			//limiter := limit.New(int(parallelism))
			var sources []string
			if inputFile != "" {
				// User has provided a file listing sources, one per line. Read sources
				var r io.Reader
				if inputFile == "-" {
					r = env.Stdin()
				} else if url, err := url.Parse(inputFile); err == nil && url.Scheme != "" {
					resp, err := http.Get(url.String())
					if err != nil {
						return err
					}
					defer func() {
						if err := resp.Body.Close(); err != nil && retErr == nil {
							retErr = err
						}
					}()
					r = resp.Body
				} else {
					inputFile, err := os.Open(inputFile)
					if err != nil {
						return err
					}
					defer func() {
						if err := inputFile.Close(); err != nil && retErr == nil {
							retErr = err
						}
					}()
					r = inputFile
				}
				// scan line by line
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					if filePath := scanner.Text(); filePath != "" {
						sources = append(sources, filePath)
					}
				}
			} else {
				// User has provided a single source
				sources = filePaths
			}

			return env.Client("user", opts...).WithModifyFileClient(file.Commit, func(mf client.ModifyFile) error {
				for _, source := range sources {
					source := source
					if file.Path == "" {
						// The user has not specified a path so we use source as path.
						if source == "-" {
							return errors.Errorf("must specify filename when reading data from stdin")
						}
						target := source
						if !fullPath {
							target = filepath.Base(source)
						}
						if err := putFileHelper(mf, joinPaths("", target), source, recursive, appendFile); err != nil {
							return err
						}
					} else if len(sources) == 1 {
						// We have a single source and the user has specified a path,
						// we use the path and ignore source (in terms of naming the file).
						if err := putFileHelper(mf, file.Path, source, recursive, appendFile); err != nil {
							return err
						}
					} else {
						// We have multiple sources and the user has specified a path,
						// we use that path as a prefix for the filepaths.
						target := source
						if !fullPath {
							target = filepath.Base(source)
						}
						if err := putFileHelper(mf, joinPaths(file.Path, target), source, recursive, appendFile); err != nil {
							return err
						}
					}
				}
				return nil
			})
		}),
	}
	putFile.Flags().StringSliceVarP(&filePaths, "file", "f", []string{"-"}, "The file to be put, it can be a local file or a URL.")
	putFile.Flags().StringVarP(&inputFile, "input-file", "i", "", "Read filepaths or URLs from a file.  If - is used, paths are read from the standard input.")
	putFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively put the files in a directory.")
	putFile.Flags().BoolVarP(&compress, "compress", "", false, "Compress data during upload. This parameter might help you upload your uncompressed data, such as CSV files, to Pachyderm faster. Use 'compress' with caution, because if your data is already compressed, this parameter might slow down the upload speed instead of increasing.")
	putFile.Flags().IntVarP(&parallelism, "parallelism", "p", DefaultParallelism, "The maximum number of files that can be uploaded in parallel.")
	putFile.Flags().BoolVarP(&appendFile, "append", "a", false, "Append to the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.")
	putFile.Flags().BoolVar(&enableProgress, "progress", isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()), "Print progress bars.")
	putFile.Flags().BoolVar(&fullPath, "full-path", false, "If true, use the entire path provided to -f as the target filename in PFS. By default only the base of the path is used.")
	shell.RegisterCompletionFunc(putFile,
		func(flag, text string, maxCompletions int64) ([]prompt.Suggest, shell.CacheFunc) {
			if flag == "-f" || flag == "--file" || flag == "-i" || flag == "input-file" {
				cs, cf := shell.FilesystemCompletion(flag, text, maxCompletions)
				return cs, shell.AndCacheFunc(cf, shell.SameFlag(flag))
			} else if flag == "" || flag == "-c" || flag == "--commit" || flag == "-o" || flag == "--append" {
				cs, cf := shell.FileCompletion(flag, text, maxCompletions)
				return cs, shell.AndCacheFunc(cf, shell.SameFlag(flag))
			}
			return nil, shell.SameFlag(flag)
		})
	commands = append(commands, cmdutil.CreateAlias(putFile, "put file"))

	copyFile := &cobra.Command{
		Use:   "{{alias}} <src-repo>@<src-branch-or-commit>:<src-path> <dst-repo>@<dst-branch-or-commit>:<dst-path>",
		Short: "Copy files between pfs paths.",
		Long:  "Copy files between pfs paths.",
		RunE: cmdutil.RunFixedArgs(2, func(args []string, env cmdutil.Env) error {
			srcFile, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			destFile, err := cmdutil.ParseFile(args[1])
			if err != nil {
				return err
			}

			var opts []client.CopyFileOption
			if appendFile {
				opts = append(opts, client.WithAppendCopyFile())
			}
			return env.Client("user", client.WithMaxConcurrentStreams(parallelism)).CopyFile(
				destFile.Commit, destFile.Path,
				srcFile.Commit, srcFile.Path,
				opts...,
			)
		}),
	}
	copyFile.Flags().BoolVarP(&appendFile, "append", "a", false, "Append to the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.")
	shell.RegisterCompletionFunc(copyFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(copyFile, "copy file"))

	var outputPath string
	getFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return the contents of a file.",
		Long:  "Return the contents of a file.",
		Example: `
# get file "XXX" on branch "master" in repo "foo"
$ {{alias}} foo@master:XXX

# get file "XXX" in the parent of the current head of branch "master"
# in repo "foo"
$ {{alias}} foo@master^:XXX

# get file "XXX" in the grandparent of the current head of branch "master"
# in repo "foo"
$ {{alias}} foo@master^2:XXX

# get file "test[].txt" on branch "master" in repo "foo"
# the path is interpreted as a glob pattern: quote and protect regex characters
$ {{alias}} 'foo@master:/test\[\].txt'`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			if !enableProgress {
				progress.Disable()
			}
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			defer progress.Wait()
			var w io.Writer
			// If an output path is given, print the output to stdout
			if outputPath == "" {
				w = env.Stdout()
			} else {
				if url, err := url.Parse(outputPath); err == nil && url.Scheme != "" {
					return env.Client("user").GetFileURL(file.Commit, file.Path, url.String())
				}
				fi, err := env.Client("user").InspectFile(file.Commit, file.Path)
				if err != nil {
					return err
				}
				f, err := progress.Create(outputPath, int64(fi.SizeBytes))
				if err != nil {
					return err
				}
				defer f.Close()
				w = f
			}
			return env.Client("user").GetFile(file.Commit, file.Path, w)
		}),
	}
	getFile.Flags().StringVarP(&outputPath, "output", "o", "", "The path where data will be downloaded.")
	getFile.Flags().BoolVar(&enableProgress, "progress", isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()), "{true|false} Whether or not to print the progress bars.")
	shell.RegisterCompletionFunc(getFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(getFile, "get file"))

	inspectFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			fileInfo, err := env.Client("user").InspectFile(file.Commit, file.Path)
			if err != nil {
				return err
			}
			if fileInfo == nil {
				return errors.Errorf("file %s not found", file.Path)
			}
			if raw {
				return cmdutil.Encoder(output, env.Stdout()).EncodeProto(fileInfo)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			return pretty.PrintDetailedFileInfo(env.Stdout(), fileInfo)
		}),
	}
	inspectFile.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(inspectFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(inspectFile, "inspect file"))

	var history string
	listFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>[:<path/in/pfs>]",
		Short: "Return the files in a directory.",
		Long:  "Return the files in a directory.",
		Example: `
# list top-level files on branch "master" in repo "foo"
$ {{alias}} foo@master

# list files under directory "dir" on branch "master" in repo "foo"
$ {{alias}} foo@master:dir

# list top-level files in the parent commit of the current head of "master"
# in repo "foo"
$ {{alias}} foo@master^

# list top-level files in the grandparent of the current head of "master"
# in repo "foo"
$ {{alias}} foo@master^2

# list the last n versions of top-level files on branch "master" in repo "foo"
$ {{alias}} foo@master --history n

# list all versions of top-level files on branch "master" in repo "foo"
$ {{alias}} foo@master --history all

# list file under directory "dir[1]" on branch "master" in repo "foo"
# the path is interpreted as a glob pattern: quote and protect regex characters
$ {{alias}} 'foo@master:dir\[1\]'`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) (retErr error) {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}

			history, err := cmdutil.ParseHistory(history)
			if err != nil {
				return errors.Wrapf(err, "error parsing history flag")
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				return env.Client("user").ListFile(file.Commit, file.Path, func(fi *pfs.FileInfo) error {
					return encoder.EncodeProto(fi)
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			header := pretty.FileHeader
			if history != 0 {
				header = pretty.FileHeaderWithCommit
			}
			writer := tabwriter.NewWriter(env.Stdout(), header)
			defer func() {
				if err := writer.Flush(); retErr == nil {
					retErr = err
				}
			}()

			return env.Client("user").ListFile(file.Commit, file.Path, func(fi *pfs.FileInfo) error {
				pretty.PrintFileInfo(writer, fi, fullTimestamps, history != 0)
				return nil
			})
		}),
	}
	listFile.Flags().AddFlagSet(outputFlags)
	listFile.Flags().AddFlagSet(timestampFlags)
	listFile.Flags().StringVar(&history, "history", "none", "Return revision history for files.")
	shell.RegisterCompletionFunc(listFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(listFile, "list file"))

	globFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<pattern>",
		Short: "Return files that match a glob pattern in a commit.",
		Long:  "Return files that match a glob pattern in a commit (that is, match a glob pattern in a repo at the state represented by a commit). Glob patterns are documented [here](https://golang.org/pkg/path/filepath/#Match).",
		Example: `
# Return files in repo "foo" on branch "master" that start
# with the character "A".  Note how the double quotation marks around the
# parameter are necessary because otherwise your shell might interpret the "*".
$ {{alias}} "foo@master:A*"

# Return files in repo "foo" on branch "master" under directory "data".
$ {{alias}} "foo@master:data/*"`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}

			fileInfos, err := env.Client("user").GlobFileAll(file.Commit, file.Path)
			if err != nil {
				return err
			}

			if raw {
				encoder := cmdutil.Encoder(output, env.Stdout())
				for _, fileInfo := range fileInfos {
					if err := encoder.EncodeProto(fileInfo); err != nil {
						return err
					}
				}
				return nil
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			writer := tabwriter.NewWriter(env.Stdout(), pretty.FileHeader)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo, fullTimestamps, false)
			}
			return writer.Flush()
		}),
	}
	globFile.Flags().AddFlagSet(outputFlags)
	globFile.Flags().AddFlagSet(timestampFlags)
	shell.RegisterCompletionFunc(globFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(globFile, "glob file"))

	var shallow bool
	var nameOnly bool
	var diffCmdArg string
	diffFile := &cobra.Command{
		Use:   "{{alias}} <new-repo>@<new-branch-or-commit>:<new-path> [<old-repo>@<old-branch-or-commit>:<old-path>]",
		Short: "Return a diff of two file trees in input repo. Diff of file trees in output repo coming soon.",
		Long:  "Return a diff of two file trees in input repo. Diff of file trees in output repo coming soon.",
		Example: `
# Return the diff of the file "path" of the input repo "foo" between the head of the
# "master" branch and its parent.
$ {{alias}} foo@master:path

# Return the diff between the master branches of input repos foo and bar at paths
# path1 and path2, respectively.
$ {{alias}} foo@master:path1 bar@master:path2`,
		RunE: cmdutil.RunBoundedArgs(1, 2, func(args []string, env cmdutil.Env) error {
			newFile, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			oldFile := &pfs.File{}
			if len(args) == 2 {
				oldFile, err = cmdutil.ParseFile(args[1])
				if err != nil {
					return err
				}
			}

			return pager.Page(noPager, env, func(w io.Writer) (retErr error) {
				var writer *tabwriter.Writer
				if nameOnly {
					writer = tabwriter.NewWriter(w, pretty.DiffFileHeader)
					defer func() {
						if err := writer.Flush(); retErr == nil {
							retErr = err
						}
					}()
				}

				newFiles, oldFiles, err := env.Client("user").DiffFileAll(
					newFile.Commit, newFile.Path,
					oldFile.Commit, oldFile.Path,
					shallow,
				)
				if err != nil {
					return err
				}
				diffCmd := diffCommand(diffCmdArg)
				return forEachDiffFile(newFiles, oldFiles, func(nFI, oFI *pfs.FileInfo) error {
					if nameOnly {
						if nFI != nil {
							pretty.PrintDiffFileInfo(writer, true, nFI, fullTimestamps)
						}
						if oFI != nil {
							pretty.PrintDiffFileInfo(writer, false, oFI, fullTimestamps)
						}
						return nil
					}
					nPath, oPath := "/dev/null", "/dev/null"
					if nFI != nil {
						nPath, err = dlFile(env.Client("user"), nFI.File)
						if err != nil {
							return err
						}
						defer func() {
							if err := os.RemoveAll(nPath); retErr == nil {
								retErr = err
							}
						}()
					}
					if oFI != nil {
						oPath, err = dlFile(env.Client("user"), oFI.File)
						defer func() {
							if err := os.RemoveAll(oPath); retErr == nil {
								retErr = err
							}
						}()
					}
					cmd := exec.Command(diffCmd[0], append(diffCmd[1:], oPath, nPath)...)
					cmd.Stdout = w
					cmd.Stderr = env.Stderr()
					// Diff returns exit code 1 when it finds differences
					// between the files, so we catch it.
					if err := cmd.Run(); err != nil && cmd.ProcessState.ExitCode() != 1 {
						return err
					}
					return nil
				})
			})
		}),
	}
	diffFile.Flags().BoolVarP(&shallow, "shallow", "s", false, "Don't descend into sub directories.")
	diffFile.Flags().BoolVar(&nameOnly, "name-only", false, "Show only the names of changed files.")
	diffFile.Flags().StringVar(&diffCmdArg, "diff-command", "", "Use a program other than git to diff files.")
	diffFile.Flags().AddFlagSet(timestampFlags)
	diffFile.Flags().AddFlagSet(pagerFlags)
	shell.RegisterCompletionFunc(diffFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(diffFile, "diff file"))

	deleteFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}

			var opts []client.DeleteFileOption
			if recursive {
				opts = append(opts, client.WithRecursiveDeleteFile())
			}
			return env.Client("user").DeleteFile(file.Commit, file.Path, opts...)
		}),
	}
	deleteFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively delete the files in a directory.")
	shell.RegisterCompletionFunc(deleteFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAlias(deleteFile, "delete file"))

	objectDocs := &cobra.Command{
		Short: "Docs for objects.",
		Long: `Objects are content-addressed blobs of data that are directly stored in the backend object store.

Objects are a low-level resource and should not be accessed directly by most users.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(objectDocs, "object", " object$"))

	var fix bool
	fsck := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Run a file system consistency check on pfs.",
		Long:  "Run a file system consistency check on the pachyderm file system, ensuring the correct provenance relationships are satisfied.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			errors := false
			if err := env.Client("user").Fsck(fix, func(resp *pfs.FsckResponse) error {
				if resp.Error != "" {
					errors = true
					fmt.Fprintf(env.Stdout(), "Error: %s\n", resp.Error)
				} else {
					fmt.Fprintf(env.Stdout(), "Fix applied: %v", resp.Fix)
				}
				return nil
			}); err != nil {
				return err
			}
			if !errors {
				fmt.Fprintln(env.Stdout(), "No errors found.")
			}
			return nil
		}),
	}
	fsck.Flags().BoolVarP(&fix, "fix", "f", false, "Attempt to fix as many issues as possible.")
	commands = append(commands, cmdutil.CreateAlias(fsck, "fsck"))

	var seed int64
	runLoadTest := &cobra.Command{
		Use:     "{{alias}} <spec>",
		Short:   "Run a PFS load test.",
		Long:    "Run a PFS load test.",
		Example: pfsload.LoadSpecification,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			spec, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			resp, err := env.Client("user").RunPFSLoadTest(spec, seed)
			if err != nil {
				return err
			}
			return cmdutil.Encoder(output, env.Stdout()).EncodeProto(resp)
		}),
	}
	runLoadTest.Flags().Int64VarP(&seed, "seed", "s", 0, "The seed to use for generating the load.")
	commands = append(commands, cmdutil.CreateAlias(runLoadTest, "run pfs-load-test"))

	// Add the mount commands (which aren't available on Windows, so they're in
	// their own file)
	commands = append(commands, mountCmds()...)

	return commands
}

func putFileHelper(mf client.ModifyFile, path, source string, recursive, appendFile bool) (retErr error) {
	// Resolve the path and convert to unix path in case we're on windows.
	path = filepath.ToSlash(filepath.Clean(path))
	var opts []client.PutFileOption
	if appendFile {
		opts = append(opts, client.WithAppendPutFile())
	}
	// try parsing the filename as a url, if it is one do a PutFileURL
	if url, err := url.Parse(source); err == nil && url.Scheme != "" {
		return mf.PutFileURL(path, url.String(), recursive, opts...)
	}
	if source == "-" {
		if recursive {
			return errors.New("cannot set -r and read from stdin (must also set -f or -i)")
		}
		// TODO: this escapes the environment rather than using env.Stdin(), so the behavior can't be tested
		stdin := progress.Stdin()
		defer stdin.Finish()
		return mf.PutFile(path, stdin, opts...)
	}
	// Resolve the source and convert to unix path in case we're on windows.
	source = filepath.ToSlash(filepath.Clean(source))
	if recursive {
		return filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			// file doesn't exist
			if info == nil {
				return errors.Errorf("%s doesn't exist", filePath)
			}
			if info.IsDir() {
				return nil
			}
			childDest := filepath.Join(path, strings.TrimPrefix(filePath, source))
			// don't do a second recursive 'put file', just put the one file at
			// filePath into childDest, and then this walk loop will go on to the
			// next one
			return putFileHelper(mf, childDest, filePath, false, appendFile)
		})
	}
	f, err := progress.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return mf.PutFile(path, f, opts...)
}

func joinPaths(prefix, filePath string) string {
	if url, err := url.Parse(filePath); err == nil && url.Scheme != "" {
		if url.Scheme == "pfs" {
			// pfs paths are of the form pfs://host/repo/branch/path we don't
			// want to prefix every file with host/repo so we remove those
			splitPath := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
			if len(splitPath) < 3 {
				return prefix
			}
			return filepath.Join(append([]string{prefix}, splitPath[2:]...)...)
		}
		return filepath.Join(prefix, strings.TrimPrefix(url.Path, "/"))
	}
	return filepath.Join(prefix, filePath)
}

func dlFile(pachClient *client.APIClient, f *pfs.File) (_ string, retErr error) {
	tempDir := filepath.Join(os.TempDir(), filepath.Dir(f.Path))
	if err := os.MkdirAll(tempDir, 0777); err != nil {
		return "", err
	}
	file, err := ioutil.TempFile(tempDir, filepath.Base(f.Path+"_"))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err := pachClient.GetFile(f.Commit, f.Path, file); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func diffCommand(cmdArg string) []string {
	if cmdArg != "" {
		return strings.Fields(cmdArg)
	}
	_, err := exec.LookPath("git")
	if err == nil {
		return []string{"git", "-c", "color.ui=always", "--no-pager", "diff", "--no-index"}
	}
	return []string{"diff"}
}

func forEachDiffFile(newFiles, oldFiles []*pfs.FileInfo, f func(newFile, oldFile *pfs.FileInfo) error) error {
	nI, oI := 0, 0
	for {
		if nI == len(newFiles) && oI == len(oldFiles) {
			return nil
		}
		var oFI *pfs.FileInfo
		var nFI *pfs.FileInfo
		switch {
		case oI == len(oldFiles) || (nI < len(newFiles) && newFiles[nI].File.Path < oldFiles[oI].File.Path):
			nFI = newFiles[nI]
			nI++
		case nI == len(newFiles) || (oI < len(oldFiles) && oldFiles[oI].File.Path < newFiles[nI].File.Path):
			oFI = oldFiles[oI]
			oI++
		case newFiles[nI].File.Path == oldFiles[oI].File.Path:
			nFI = newFiles[nI]
			nI++
			oFI = oldFiles[oI]
			oI++
		}
		if err := f(nFI, oFI); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
}

func parseOriginKind(input string) (pfs.OriginKind, error) {
	if input == "" {
		return pfs.OriginKind_ORIGIN_KIND_UNKNOWN, nil
	}

	result := pfs.OriginKind(pfs.OriginKind_value[strings.ToUpper(input)])
	if result == pfs.OriginKind_ORIGIN_KIND_UNKNOWN {
		names := []string{}
		for name, value := range pfs.OriginKind_value {
			if pfs.OriginKind(value) != pfs.OriginKind_ORIGIN_KIND_UNKNOWN {
				names = append(names, name)
			}
		}
		return pfs.OriginKind_ORIGIN_KIND_UNKNOWN, errors.Errorf("unknown commit origin type '%s', must be one of: %s", input, strings.Join(names, ", "))
	}

	return result, nil
}

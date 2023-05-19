package cmds

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pager"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/shell"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	txncmds "github.com/pachyderm/pachyderm/v2/src/server/transaction/cmds"
)

const (
	// DefaultParallelism is the default parallelism used by 'get file' and 'put file'.
	DefaultParallelism = 10

	// Plural variables are used below for user convenience.
	branches = "branches"
	commits  = "commits"
	files    = "files"
	repos    = "repos"
	projects = "projects"
)

// Cmds returns a slice containing pfs commands.
func Cmds(mainCtx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	var fullTimestamps bool
	timestampFlags := cmdutil.TimestampFlags(&fullTimestamps)

	var timeout time.Duration
	var limit uint32
	var jsonOutput bool

	var noPager bool
	pagerFlags := cmdutil.PagerFlags(&noPager)

	repoDocs := &cobra.Command{
		Short: "Docs for repos.",
		Long: `Repos, short for repository, are the top level data objects in Pachyderm.

Repos contain version-controlled directories and files. Files can be of any size
or type (e.g. csv, binary, images, etc).`,
	}
	commands = append(commands, cmdutil.CreateDocsAliases(repoDocs, "repo", " repo$", repos))

	var description string
	var allProjects bool
	project := pachCtx.Project
	createRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err = c.PfsAPIClient.CreateRepo(
					c.Ctx(),
					&pfs.CreateRepoRequest{
						Repo:        client.NewRepo(project, args[0]),
						Description: description,
					},
				)
				return errors.EnsureStack(err)
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createRepo.Flags().StringVarP(&description, "description", "d", "", "A description of the repo.")
	createRepo.Flags().StringVar(&project, "project", project, "The project to create the repo in.")
	commands = append(commands, cmdutil.CreateAliases(createRepo, "create repo", repos))

	updateRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Update a repo.",
		Long:  "Update a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err = c.PfsAPIClient.CreateRepo(
					c.Ctx(),
					&pfs.CreateRepoRequest{
						Repo:        cmdutil.ParseRepo(project, args[0]),
						Description: description,
						Update:      true,
					},
				)
				return errors.EnsureStack(err)
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateRepo.Flags().StringVarP(&description, "description", "d", "", "A description of the repo.")
	updateRepo.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(updateRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(updateRepo, "update repo", repos))

	inspectRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			repoInfo, err := c.PfsAPIClient.InspectRepo(c.Ctx(), &pfs.InspectRepoRequest{Repo: cmdutil.ParseRepo(project, args[0])})
			if err != nil {
				return errors.EnsureStack(err)
			}
			if repoInfo == nil {
				return errors.Errorf("repo %s not found", args[0])
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(repoInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			ri := &pretty.PrintableRepoInfo{
				RepoInfo:       repoInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedRepoInfo(ri)
		}),
	}
	inspectRepo.Flags().AddFlagSet(outputFlags)
	inspectRepo.Flags().AddFlagSet(timestampFlags)
	inspectRepo.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(inspectRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectRepo, "inspect repo", repos))

	var all bool
	var repoType string
	listRepo := &cobra.Command{
		Short: "Return a list of repos.",
		Long:  "Return a list of repos. By default, hide system repos like pipeline metadata",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			// Call ListRepo RPC with sane defaults
			if all {
				repoType = ""
			}
			projectsFilter := []*pfs.Project{{Name: project}}
			if allProjects {
				projectsFilter = nil
			}
			repoInfos, err := c.ListProjectRepo(&pfs.ListRepoRequest{Type: repoType, Projects: projectsFilter})
			if err != nil {
				return errors.Wrap(err, "cannot list repos")
			}

			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				for _, repoInfo := range repoInfos {
					if err := encoder.EncodeProto(repoInfo); err != nil {
						return errors.EnsureStack(err)
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
			writer := tabwriter.NewWriter(os.Stdout, header)
			for _, repoInfo := range repoInfos {
				pretty.PrintRepoInfo(writer, repoInfo, fullTimestamps)
			}
			return writer.Flush()
		}),
	}
	listRepo.Flags().AddFlagSet(outputFlags)
	listRepo.Flags().AddFlagSet(timestampFlags)
	listRepo.Flags().BoolVar(&all, "all", false, "include system repos of all types")
	listRepo.Flags().StringVar(&repoType, "type", pfs.UserRepoType, "only include repos of the given type")
	listRepo.Flags().StringVar(&project, "project", project, "project in which repo is located")
	listRepo.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "show repos from all projects")
	commands = append(commands, cmdutil.CreateAliases(listRepo, "list repo", repos))

	var force bool
	deleteRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			request := &pfs.DeleteRepoRequest{
				Force: force,
			}

			if len(args) > 0 {
				if all {
					return errors.Errorf("cannot use the --all flag with an argument")
				}
				if allProjects {
					return errors.Errorf("cannot use the --all-projects flag with an argument")
				}
				request.Repo = cmdutil.ParseRepo(project, args[0])
			} else if !all {
				return errors.Errorf("either a repo name or the --all flag needs to be provided")
			}
			if all {
				var req = new(pfs.DeleteReposRequest)
				if allProjects {
					req.All = true
				} else {
					req.Projects = []*pfs.Project{
						{
							Name: project,
						},
					}
				}
				_, err := c.PfsAPIClient.DeleteRepos(c.Ctx(), req)
				return errors.EnsureStack(err)
			}

			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err := c.PfsAPIClient.DeleteRepo(c.Ctx(), request)
				return errors.EnsureStack(err)
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	deleteRepo.Flags().BoolVarP(&force, "force", "f", false, "remove the repo regardless of errors; use with care")
	deleteRepo.Flags().BoolVar(&all, "all", false, "remove all repos")
	deleteRepo.Flags().StringVar(&project, "project", project, "project in which repo is located")
	deleteRepo.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "delete repos from all projects; only valid with --all")
	shell.RegisterCompletionFunc(deleteRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteRepo, "delete repo", repos))

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
	commands = append(commands, cmdutil.CreateDocsAliases(commitDocs, "commit", " commit$", commits))

	var parent string
	startCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Start a new commit.",
		Long:  "Start a new commit with parent-commit as the parent on the given branch; if the branch does not exist, it will be created.",
		Example: `
# Start a commit in repo "test" on branch "master"
$ {{alias}} test@master

# Start a commit with "master" as the parent in repo "test", on a new branch "patch"; essentially a fork.
$ {{alias}} test@patch -p master

# Start a commit with XXX as the parent in repo "test" on the branch "fork"
$ {{alias}} test@fork -p XXX`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(project, args[0])
			if err != nil {
				return err
			}
			c, err := newClient(mainCtx, pachctlCfg)
			if err != nil {
				return err
			}
			defer c.Close()

			var parentCommit *pfs.Commit
			if parent != "" {
				// We don't know if the parent is a commit ID, branch, or ancestry, so
				// construct a string to parse.
				parentCommit, err = cmdutil.ParseCommit(project, fmt.Sprintf("%s@%s", branch.Repo, parent))
				if err != nil {
					return err
				}
			}

			var commit *pfs.Commit
			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				var err error
				commit, err = c.PfsAPIClient.StartCommit(
					c.Ctx(),
					&pfs.StartCommitRequest{
						Branch:      branch,
						Parent:      parentCommit,
						Description: description,
					},
				)
				return errors.EnsureStack(err)
			})
			if err == nil {
				fmt.Println(commit.ID)
			}
			return grpcutil.ScrubGRPC(err)
		}),
	}
	startCommit.Flags().StringVarP(&parent, "parent", "p", "", "The parent of the new commit, unneeded if branch is specified and you want to use the previous head of the branch as the parent.")
	startCommit.MarkFlagCustom("parent", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	startCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents")
	startCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")
	startCommit.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(startCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(startCommit, "start commit", commits))

	finishCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			commit, err := cmdutil.ParseCommit(project, args[0])
			if err != nil {
				return err
			}
			c, err := newClient(mainCtx, pachctlCfg)
			if err != nil {
				return err
			}
			defer c.Close()

			err = txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err = c.PfsAPIClient.FinishCommit(
					c.Ctx(),
					&pfs.FinishCommitRequest{
						Commit:      commit,
						Description: description,
						Force:       force,
					},
				)
				return errors.EnsureStack(err)
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	finishCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents (overwrites any existing commit description)")
	finishCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")
	finishCommit.Flags().BoolVarP(&force, "force", "f", false, "finish the commit even if it has provenance, which could break jobs; prefer 'stop job'")
	finishCommit.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(finishCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(finishCommit, "finish commit", commits))

	inspectCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			commit, err := cmdutil.ParseCommit(project, args[0])
			if err != nil && uuid.IsUUIDWithoutDashes(args[0]) {
				return errors.New(`Use "list commit <id>" to see commits with a given ID across different repos`)
			} else if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

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
				return errors.Errorf("commit %s not found", commit)
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(commitInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			ci := &pretty.PrintableCommitInfo{
				CommitInfo:     commitInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedCommitInfo(os.Stdout, ci)
		}),
	}
	inspectCommit.Flags().AddFlagSet(outputFlags)
	inspectCommit.Flags().AddFlagSet(timestampFlags)
	inspectCommit.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(inspectCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectCommit, "inspect commit", commits))

	var from string
	var number int64
	var originStr string
	var expand bool
	listCommit := &cobra.Command{
		Use:   "{{alias}} [<commit-id>|<repo>[@<branch-or-commit>]]",
		Short: "Return a list of commits.",
		Long:  "Return a list of commits, either across the entire pachyderm cluster or restricted to a single repo.",
		Example: `
# return all commits
$ {{alias}}

# return commits in repo "foo"
$ {{alias}} foo

# return all sub-commits in a commit
$ {{alias}} <commit-id>

# return commits in repo "foo" on branch "master"
$ {{alias}} foo@master

# return the last 20 commits in repo "foo" on branch "master"
$ {{alias}} foo@master -n 20

# return commits in repo "foo" on branch "master" since commit XXX
$ {{alias}} foo@master --from XXX`,
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			if !raw && output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			} else if all && originStr != "" {
				return errors.New("cannot specify both --all and --origin")
			}

			if len(args) == 0 {
				// Outputting all commitsets
				if originStr != "" {
					return errors.Errorf("cannot specify --origin when listing all commits")
				} else if from != "" {
					return errors.Errorf("cannot specify --from when listing all commits")
				}

				listCommitSetClient, err := c.PfsAPIClient.ListCommitSet(c.Ctx(), &pfs.ListCommitSetRequest{Project: &pfs.Project{Name: project}})
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}

				count := 0
				if !expand {
					if raw {
						e := cmdutil.Encoder(output, os.Stdout)
						return grpcutil.ForEach[*pfs.CommitSetInfo](listCommitSetClient, func(commitSetInfo *pfs.CommitSetInfo) error {
							if err := e.EncodeProto(commitSetInfo); err != nil {
								return errors.EnsureStack(err)
							}
							count++
							if number != 0 && count >= int(number) {
								return errutil.ErrBreak
							}
							return nil
						})
					}

					return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
						writer := tabwriter.NewWriter(w, pretty.CommitSetHeader)
						if err := grpcutil.ForEach[*pfs.CommitSetInfo](listCommitSetClient, func(commitSetInfo *pfs.CommitSetInfo) error {
							pretty.PrintCommitSetInfo(writer, commitSetInfo, fullTimestamps)
							count++
							if number != 0 && count >= int(number) {
								return errutil.ErrBreak
							}
							return nil
						}); err != nil {
							return err
						}
						return writer.Flush()
					})
				} else {
					if raw {
						e := cmdutil.Encoder(output, os.Stdout)
						return grpcutil.ForEach[*pfs.CommitSetInfo](listCommitSetClient, func(commitSetInfo *pfs.CommitSetInfo) error {
							for _, commitInfo := range commitSetInfo.Commits {
								if err := e.EncodeProto(commitInfo); err != nil {
									return errors.EnsureStack(err)
								}
								count++
								if number != 0 && count >= int(number) {
									return errutil.ErrBreak
								}
							}
							return nil
						})
					}

					return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
						writer := tabwriter.NewWriter(w, pretty.CommitHeader)
						if err := grpcutil.ForEach[*pfs.CommitSetInfo](listCommitSetClient, func(commitSetInfo *pfs.CommitSetInfo) error {
							for _, commitInfo := range commitSetInfo.Commits {
								pretty.PrintCommitInfo(writer, commitInfo, fullTimestamps)
								count++
								if number != 0 && count >= int(number) {
									return errutil.ErrBreak
								}
							}
							return nil
						}); err != nil {
							return err
						}
						return writer.Flush()
					})
				}
			} else if len(args) == 1 && uuid.IsUUIDWithoutDashes(args[0]) {
				// Outputting commits from one commitset
				if from != "" {
					return errors.Errorf("cannot specify --from when listing subcommits")
				} else if all {
					return errors.Errorf("cannot specify --all when listing subcommits")
				} else if originStr != "" {
					return errors.Errorf("cannot specify --origin when listing subcommits")
				}

				commitInfos, err := c.InspectCommitSet(args[0])
				if err != nil {
					return errors.Wrap(err, "error from InspectCommitSet")
				}

				if number != 0 && len(commitInfos) > int(number) {
					commitInfos = commitInfos[:number]
				}

				if raw {
					encoder := cmdutil.Encoder(output, os.Stdout)
					for _, commitInfo := range commitInfos {
						if err := encoder.EncodeProto(commitInfo); err != nil {
							return errors.EnsureStack(err)
						}
					}
					return nil
				}

				return pager.Page(noPager, os.Stdout, func(w io.Writer) error {
					writer := tabwriter.NewWriter(w, pretty.CommitHeader)
					for _, commitInfo := range commitInfos {
						pretty.PrintCommitInfo(writer, commitInfo, fullTimestamps)
					}
					return writer.Flush()
				})
			} else {
				// Outputting filtered commits
				toCommit, err := cmdutil.ParseCommit(project, args[0])
				if err != nil {
					return err
				}

				repo := toCommit.Branch.Repo

				var fromCommit *pfs.Commit
				if from != "" {
					fromCommit = repo.NewCommit("", from)
				}

				if toCommit.ID == "" && toCommit.Branch.Name == "" {
					// just a repo
					toCommit = nil
				}

				origin, err := parseOriginKind(originStr)
				if err != nil {
					return err
				}

				listClient, err := c.PfsAPIClient.ListCommit(c.Ctx(), &pfs.ListCommitRequest{
					Repo:       repo,
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
					encoder := cmdutil.Encoder(output, os.Stdout)
					return grpcutil.ForEach[*pfs.CommitInfo](listClient, func(ci *pfs.CommitInfo) error {
						return errors.EnsureStack(encoder.EncodeProto(ci))
					})
				}
				writer := tabwriter.NewWriter(os.Stdout, pretty.CommitHeader)
				if err := grpcutil.ForEach[*pfs.CommitInfo](listClient, func(ci *pfs.CommitInfo) error {
					pretty.PrintCommitInfo(writer, ci, fullTimestamps)
					return nil
				}); err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				return writer.Flush()
			}
		}),
	}
	listCommit.Flags().StringVarP(&from, "from", "f", "", "list all commits since this commit")
	listCommit.Flags().Int64VarP(&number, "number", "n", 0, "list only this many commits; if set to zero, list all commits")
	listCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	listCommit.Flags().BoolVar(&all, "all", false, "return all types of commits, including aliases")
	listCommit.Flags().BoolVarP(&expand, "expand", "x", false, "show one line for each sub-commmit and include more columns")
	listCommit.Flags().StringVar(&originStr, "origin", "", "only return commits of a specific type")
	listCommit.Flags().AddFlagSet(outputFlags)
	listCommit.Flags().AddFlagSet(timestampFlags)
	listCommit.Flags().StringVar(&project, "project", project, "Project in which commit is located.")
	shell.RegisterCompletionFunc(listCommit, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(listCommit, "list commit", commits))

	waitCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Wait for the specified commit to finish and return it.",
		Long:  "Wait for the specified commit to finish and return it.",
		Example: `
# wait for the commit foo@XXX to finish and return it
$ {{alias}} foo@XXX -b bar@baz`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			commit, err := cmdutil.ParseCommit(project, args[0])
			if err != nil {
				return err
			}

			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			commitInfo, err := c.WaitCommit(commit.Branch.Repo.Project.GetName(), commit.Branch.Repo.Name, commit.Branch.Name, commit.ID)
			if err != nil {
				return err
			}

			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(commitInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			ci := &pretty.PrintableCommitInfo{
				CommitInfo:     commitInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedCommitInfo(os.Stdout, ci)
		}),
	}
	waitCommit.Flags().AddFlagSet(outputFlags)
	waitCommit.Flags().AddFlagSet(timestampFlags)
	waitCommit.Flags().StringVar(&project, "project", project, "Project containing commit.")
	commands = append(commands, cmdutil.CreateAliases(waitCommit, "wait commit", commits))

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
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			branch, err := cmdutil.ParseBranch(project, args[0])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

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
				encoder := cmdutil.Encoder(output, os.Stdout)
				return grpcutil.ForEach[*pfs.CommitInfo](subscribeClient, func(ci *pfs.CommitInfo) error {
					return errors.EnsureStack(encoder.EncodeProto(ci))
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			w := tabwriter.NewWriter(os.Stdout, pretty.CommitHeader)
			defer func() {
				if err := w.Flush(); retErr == nil {
					retErr = err
				}
			}()
			if err := grpcutil.ForEach[*pfs.CommitInfo](subscribeClient, func(ci *pfs.CommitInfo) error {
				pretty.PrintCommitInfo(w, ci, fullTimestamps)
				return nil
			}); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return errors.EnsureStack(err)
		}),
	}
	subscribeCommit.Flags().StringVar(&from, "from", "", "subscribe to all commits since this commit")
	subscribeCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	subscribeCommit.Flags().BoolVar(&newCommits, "new", false, "subscribe to only new commits created from now on")
	subscribeCommit.Flags().BoolVar(&all, "all", false, "return all types of commits, including aliases")
	subscribeCommit.Flags().StringVar(&originStr, "origin", "", "only return commits of a specific type")
	subscribeCommit.Flags().AddFlagSet(outputFlags)
	subscribeCommit.Flags().AddFlagSet(timestampFlags)
	subscribeCommit.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(subscribeCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(subscribeCommit, "subscribe commit", commits))

	squashCommit := &cobra.Command{
		Use:   "{{alias}} <commit-id>",
		Short: "Squash the sub-commits of a commit.",
		Long: `Squash the sub-commits of a commit.  The data in the sub-commits will remain in their child commits.
The squash will fail if it includes a commit with no children`,

		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			return txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				return c.SquashCommitSet(args[0])
			})
		}),
	}
	shell.RegisterCompletionFunc(squashCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(squashCommit, "squash commit", commits))

	deleteCommit := &cobra.Command{
		Use:   "{{alias}} <commit-id>",
		Short: "Delete the sub-commits of a commit.",
		Long: `Delete the sub-commits of a commit.  The data in the sub-commits will be lost.
This operation is only supported if none of the sub-commits have children.`,

		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			return txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				return c.DropCommitSet(args[0])
			})
		}),
	}
	shell.RegisterCompletionFunc(deleteCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteCommit, "delete commit", commits))

	branchDocs := &cobra.Command{
		Short: "Docs for branches.",
		Long: `A branch in Pachyderm records provenance relationships between data in different repos,
as well as being as an alias for a commit in its repo.

The branch reference will "float" to always refer to the latest commit on the
branch, known as the HEAD commit. All commits are on exactly one branch.

Any pachctl command that can take a commit, can take a branch name instead.`,
	}
	commands = append(commands, cmdutil.CreateDocsAliases(branchDocs, "branch", " branch$", branches))

	var branchProvenance cmdutil.RepeatedStringArg
	var head string
	trigger := &pfs.Trigger{}
	createBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Create a new branch, or update an existing branch, on a repo.",
		Long:  "Create a new branch, or update an existing branch, on a repo, starting a commit on the branch will also create it, so there's often no need to call this.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(project, args[0])
			if err != nil {
				return err
			}
			provenance, err := cmdutil.ParseBranches(project, branchProvenance)
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
					headCommit, err = cmdutil.ParseCommit(project, head)
					if err != nil {
						return err
					}
				} else {
					// treat head as the commitID or branch name
					headCommit = branch.Repo.NewCommit("", head)
				}
			}

			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			return txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
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
	createBranch.Flags().VarP(&branchProvenance, "provenance", "p", "The provenance for the branch. format: <repo>@<branch>")
	createBranch.MarkFlagCustom("provenance", "__pachctl_get_repo_commit")
	createBranch.Flags().StringVarP(&head, "head", "", "", "The head of the newly created branch. Either pass the commit with format: <branch-or-commit>, or fully-qualified as <repo>@<branch>=<id>")
	createBranch.MarkFlagCustom("head", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	createBranch.Flags().StringVarP(&trigger.Branch, "trigger", "t", "", "The branch to trigger this branch on.")
	createBranch.Flags().StringVar(&trigger.CronSpec, "trigger-cron", "", "The cron spec to use in triggering.")
	createBranch.Flags().StringVar(&trigger.Size_, "trigger-size", "", "The data size to use in triggering.")
	createBranch.Flags().Int64Var(&trigger.Commits, "trigger-commits", 0, "The number of commits to use in triggering.")
	createBranch.Flags().BoolVar(&trigger.All, "trigger-all", false, "Only trigger when all conditions are met, rather than when any are met.")
	createBranch.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	commands = append(commands, cmdutil.CreateAliases(createBranch, "create branch", branches))

	inspectBranch := &cobra.Command{
		Use:   "{{alias}}  <repo>@<branch>",
		Short: "Return info about a branch.",
		Long:  "Return info about a branch.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			branch, err := cmdutil.ParseBranch(project, args[0])
			if err != nil {
				return err
			}

			branchInfo, err := c.PfsAPIClient.InspectBranch(c.Ctx(), &pfs.InspectBranchRequest{Branch: branch})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if branchInfo == nil {
				return errors.Errorf("branch %s not found", args[0])
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(branchInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			return pretty.PrintDetailedBranchInfo(branchInfo)
		}),
	}
	inspectBranch.Flags().AddFlagSet(outputFlags)
	inspectBranch.Flags().AddFlagSet(timestampFlags)
	inspectBranch.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(inspectBranch, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectBranch, "inspect branch", branches))

	listBranch := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return all branches on a repo.",
		Long:  "Return all branches on a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			branchClient, err := c.PfsAPIClient.ListBranch(c.Ctx(), &pfs.ListBranchRequest{Repo: cmdutil.ParseRepo(project, args[0])})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				err := grpcutil.ForEach[*pfs.BranchInfo](branchClient, func(branch *pfs.BranchInfo) error {
					return errors.EnsureStack(encoder.EncodeProto(branch))
				})
				return grpcutil.ScrubGRPC(err)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}

			writer := tabwriter.NewWriter(os.Stdout, pretty.BranchHeader)
			if err := grpcutil.ForEach[*pfs.BranchInfo](branchClient, func(branch *pfs.BranchInfo) error {
				pretty.PrintBranch(writer, branch)
				return nil
			}); err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return writer.Flush()
		}),
	}
	listBranch.Flags().AddFlagSet(outputFlags)
	listBranch.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(listBranch, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(listBranch, "list branch", branches))

	deleteBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Delete a branch",
		Long:  "Delete a branch, while leaving the commits intact",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(project, args[0])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			return txncmds.WithActiveTransaction(c, func(c *client.APIClient) error {
				_, err := c.PfsAPIClient.DeleteBranch(c.Ctx(), &pfs.DeleteBranchRequest{Branch: branch, Force: force})
				return errors.EnsureStack(err)
			})
		}),
	}
	deleteBranch.Flags().BoolVarP(&force, "force", "f", false, "remove the branch regardless of errors; use with care")
	deleteBranch.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(deleteBranch, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteBranch, "delete branch", branches))

	FindCommits := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs> [flags]",
		Short: "find commits with reference to <filePath> within a branch starting from <repo@commitID>",
		Long:  "find commits with reference to <filePath> within a branch starting from <repo@commitID>",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			commit := file.Commit
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			ctx := c.Ctx()
			var cf context.CancelFunc
			if timeout != time.Duration(0) {
				ctx, cf = context.WithTimeout(c.Ctx(), timeout)
				defer cf()
			}
			req := &pfs.FindCommitsRequest{
				Start:    commit,
				FilePath: file.Path,
				Limit:    limit,
			}
			if jsonOutput {
				resp, err := c.FindCommits(req)
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				jsonResp, err := json.Marshal(resp)
				if err != nil {
					return grpcutil.ScrubGRPC(err)
				}
				_, err = fmt.Fprintf(os.Stdout, "%s\n", string(jsonResp))
				return grpcutil.ScrubGRPC(err)
			}
			findCommitClient, err := c.PfsAPIClient.FindCommits(ctx, req)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return grpcutil.ScrubGRPC(pretty.PrintFindCommits(findCommitClient))
		}),
	}
	FindCommits.Flags().BoolVar(&jsonOutput, "json", jsonOutput, "print the response in json")
	FindCommits.Flags().Uint32Var(&limit, "limit", limit, "Number of matching commits to return")
	FindCommits.Flags().DurationVar(&timeout, "timeout", timeout, "Search duration timeout")
	FindCommits.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(FindCommits, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(FindCommits, "find commit", commits))

	projectDocs := &cobra.Command{
		Short: "Docs for projects.",
		Long: `Projects are the top level organizational objects in Pachyderm.

Projects contain pachyderm objects such as Repos and Pipelines.`,
	}
	commands = append(commands, cmdutil.CreateDocsAliases(projectDocs, "project", " project", projects))

	createProject := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Create a new project.",
		Long:  "Create a new project.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			_, err = c.PfsAPIClient.CreateProject(
				c.Ctx(),
				&pfs.CreateProjectRequest{
					Project:     &pfs.Project{Name: args[0]},
					Description: description,
				})
			return grpcutil.ScrubGRPC(err)

		}),
	}
	createProject.Flags().StringVarP(&description, "description", "d", "", "The description of the newly-created project.")
	commands = append(commands, cmdutil.CreateAliases(createProject, "create project", projects))

	updateProject := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Update a project.",
		Long:  "Update a project.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			_, err = c.PfsAPIClient.CreateProject(
				c.Ctx(),
				&pfs.CreateProjectRequest{
					Project:     &pfs.Project{Name: args[0]},
					Description: description,
					Update:      true,
				})
			return grpcutil.ScrubGRPC(err)

		}),
	}
	updateProject.Flags().StringVarP(&description, "description", "d", "", "The description of the updated project.")
	shell.RegisterCompletionFunc(updateProject, shell.ProjectCompletion)
	commands = append(commands, cmdutil.CreateAliases(updateProject, "update project", projects))

	inspectProject := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Inspect a project.",
		Long:  "Inspect a project.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			pi, err := c.PfsAPIClient.InspectProject(
				c.Ctx(),
				&pfs.InspectProjectRequest{
					Project: &pfs.Project{Name: args[0]},
				})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(pi))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			return pretty.PrintDetailedProjectInfo(pi)
		}),
	}
	inspectProject.Flags().AddFlagSet(outputFlags)
	shell.RegisterCompletionFunc(inspectProject, shell.ProjectCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectProject, "inspect project", projects))

	listProject := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return all projects.",
		Long:  "Return all projects.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			pis, err := c.ListProject()
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				for _, pi := range pis {
					if err := encoder.EncodeProto(pi); err != nil {
						return errors.EnsureStack(err)
					}
				}
				return grpcutil.ScrubGRPC(err)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			header := pretty.ProjectHeader
			if len(pis) > 0 && pis[0].AuthInfo != nil {
				header = pretty.ProjectAuthHeader
			}
			writer := tabwriter.NewWriter(os.Stdout, header)
			for _, pi := range pis {
				pretty.PrintProjectInfo(writer, pi, &pfs.Project{Name: pachCtx.Project})
			}
			return writer.Flush()
		}),
	}
	listProject.Flags().AddFlagSet(outputFlags)
	commands = append(commands, cmdutil.CreateAliases(listProject, "list project", projects))

	deleteProject := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Delete a project.",
		Long:  "Delete a project.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			project := args[0]
			pipelineResp, err := c.PpsAPIClient.ListPipeline(
				c.Ctx(),
				&pps.ListPipelineRequest{
					Projects: []*pfs.Project{{Name: project}},
				},
			)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			pp, err := grpcutil.Collect[*pps.PipelineInfo](pipelineResp, 1000)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if len(pp) > 0 {
				for _, p := range pp {
					fmt.Printf("This will delete pipeline %s\n?", p.Pipeline)
				}
				if ok, err := cmdutil.InteractiveConfirm(); err != nil {
					return err
				} else if !ok {
					return errors.Errorf("cannot delete project with %d pipelines", len(pp))
				}
				for _, p := range pp {
					if _, err := c.PpsAPIClient.DeletePipeline(c.Ctx(), &pps.DeletePipelineRequest{Pipeline: p.Pipeline}); err != nil {
						return grpcutil.ScrubGRPC(err)
					}
				}
			}
			repoResp, err := c.PfsAPIClient.ListRepo(
				c.Ctx(),
				&pfs.ListRepoRequest{
					Projects: []*pfs.Project{{Name: project}},
					Type:     pfs.UserRepoType,
				},
			)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			rr, err := grpcutil.Collect[*pfs.RepoInfo](repoResp, 1000)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if len(rr) > 0 {
				for _, r := range rr {
					fmt.Printf("This will delete repo %s\n", r.Repo)
				}
				if ok, err := cmdutil.InteractiveConfirm(); err != nil {
					return err
				} else if !ok {
					return errors.Errorf("cannot delete project with %d repos", len(rr))
				}
				for _, r := range rr {
					if _, err := c.PfsAPIClient.DeleteRepo(c.Ctx(), &pfs.DeleteRepoRequest{Repo: r.Repo}); err != nil {
						return grpcutil.ScrubGRPC(err)
					}
				}
			}
			_, err = c.PfsAPIClient.DeleteProject(
				c.Ctx(),
				&pfs.DeleteProjectRequest{
					Project: &pfs.Project{Name: args[0]},
					Force:   force,
				})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if args[0] == pachCtx.Project {
				fmt.Fprintf(os.Stderr, "warning: deleted current project %s; update context by running:\n   pachctl config update context --project PROJECT\n", pachCtx.Project)
			}
			return nil
		}),
	}
	shell.RegisterCompletionFunc(deleteProject, shell.ProjectCompletion)
	deleteProject.Flags().BoolVarP(&force, "force", "f", false, "remove the project regardless of errors; use with care")
	commands = append(commands, cmdutil.CreateAliases(deleteProject, "delete project", projects))

	fileDocs := &cobra.Command{
		Short: "Docs for files.",
		Long: `Files are the lowest level data objects in Pachyderm.

Files can be of any type (e.g. csv, binary, images, etc) or size and can be
written to started (but not finished) commits with 'put file'. Files can be read
from commits with 'get file'.`,
	}
	commands = append(commands, cmdutil.CreateDocsAliases(fileDocs, "file", " file$", files))

	var filePaths []string
	var inputFile string
	var recursive bool
	var parallelism int
	var appendFile bool
	var compress bool
	var enableProgress bool
	var fullPath bool
	var untar bool
	putFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>[:<path/to/file>]",
		Short: "Put a file into the filesystem.",
		Long:  "Put a file into the filesystem.  This command supports a number of ways to insert data into PFS.",
		Example: `
# Put data from stdin at repo@branch:/path
$ echo "data" | {{alias}} repo@branch:/path

# Put a file from the local filesystem at repo@branch:/file
$ {{alias}} repo@branch -f file

# Put a file from the local filesystem at repo@branch:/path
$ {{alias}} repo@branch:/path -f file

# Put the contents of a directory at repo@branch:/dir/file
$ {{alias}} -r repo@branch -f dir

# Put the contents of a directory at repo@branch:/path/file (without /dir)
$ {{alias}} -r repo@branch:/path -f dir

# Put the data from a URL at repo@branch:/example.png
$ {{alias}} repo@branch -f http://host/example.png

# Put the data from a URL at repo@branch:/dir/example.png
$ {{alias}} repo@branch:/dir -f http://host/example.png

# Put the data from an S3 bucket at repo@branch:/s3_object
$ {{alias}} repo@branch -r -f s3://my_bucket

# Put several files or URLs that are listed in file.
# Files and URLs should be newline delimited.
$ {{alias}} repo@branch -i file

# Put several files or URLs that are listed at URL.
# NOTE this URL can reference local files, so it could cause you to put sensitive
# files into your Pachyderm cluster.
$ {{alias}} repo@branch -i http://host/path`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			if !enableProgress {
				progress.Disable()
			}
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			opts := []client.Option{client.WithMaxConcurrentStreams(parallelism)}
			if compress {
				opts = append(opts, client.WithGZIPCompression())
			}
			c, err := newClient(mainCtx, pachctlCfg, opts...)
			if err != nil {
				return err
			}
			defer c.Close()
			defer progress.Wait()

			// check whether or not the repo exists before attempting to upload
			if _, err = c.InspectRepo(file.Commit.Branch.Repo.Project.GetName(), file.Commit.Branch.Repo.Name); err != nil {
				if errutil.IsNotFoundError(err) {
					return err
				}
				return errors.Wrapf(err, "could not inspect repo %s", err, file.Commit.Branch.Repo.Name)
			}

			// TODO: Rethink put file parallelism for 2.0.
			// Doing parallel uploads at the file level for small files will be bad, but we still want a clear way to parallelize large file uploads.
			//limiter := limit.New(int(parallelism))
			var sources []string
			if inputFile != "" {
				// User has provided a file listing sources, one per line. Read sources
				var r io.Reader
				if inputFile == "-" {
					r = os.Stdin
				} else if url, err := url.Parse(inputFile); err == nil && url.Scheme != "" {
					resp, err := http.Get(url.String())
					if err != nil {
						return errors.EnsureStack(err)
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
						return errors.EnsureStack(err)
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

			return c.WithModifyFileClient(file.Commit, func(mf client.ModifyFile) error {
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
						if err := putFileHelper(mf, joinPaths("", target), source, recursive, appendFile, untar); err != nil {
							return err
						}
					} else if len(sources) == 1 {
						// We have a single source and the user has specified a path,
						// we use the path and ignore source (in terms of naming the file).
						if err := putFileHelper(mf, file.Path, source, recursive, appendFile, untar); err != nil {
							return err
						}
					} else {
						// We have multiple sources and the user has specified a path,
						// we use that path as a prefix for the filepaths.
						target := source
						if !fullPath {
							target = filepath.Base(source)
						}
						if err := putFileHelper(mf, joinPaths(file.Path, target), source, recursive, appendFile, untar); err != nil {
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
	putFile.Flags().BoolVar(&untar, "untar", false, "If true, file(s) with the extension .tar are untarred and put as a separate file for each file within the tar stream(s). gzipped (.tar.gz or .tgz) tar file(s) are handled as well")
	putFile.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
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
	commands = append(commands, cmdutil.CreateAliases(putFile, "put file", files))

	var srcProject, destProject string
	copyFile := &cobra.Command{
		Use:   "{{alias}} <src-repo>@<src-branch-or-commit>:<src-path> <dst-repo>@<dst-branch-or-commit>:<dst-path>",
		Short: "Copy files between pfs paths.",
		Long:  "Copy files between pfs paths.",
		Example: `
# copy between repos within the current project defined by the pachyderm context
# defaults to the "default" project
$ {{alias}} srcRepo@master:/file destRepo@master:/file

# copy within a specified project
$ {{alias}} srcRepo@master:/file destRepo@master:/file --project myProject

# copy from the current project to a different project
# here, srcRepo is in the current project, while destRepo is in myProject
$ {{alias}} srcRepo@master:/file destRepo@master:/file --dest-project myProject

# copy from a different project to the current project
# here, srcRepo is in myProject, while destRepo is in the current project
$ {{alias}} srcRepo@master:/file destRepo@master:/file --src-project myProject

# copy between repos across two different projects
# here, srcRepo is in project1, while destRepo is in project2
$ {{alias}} srcRepo@master:/file destRepo@master:/file --src-project project1 --dest-project project2`,
		Run: cmdutil.RunFixedArgs(2, func(args []string) (retErr error) {
			if srcProject == "" {
				srcProject = project
			}
			srcFile, err := cmdutil.ParseFile(srcProject, args[0])
			if err != nil {
				return err
			}
			if destProject == "" {
				destProject = project
			}
			destFile, err := cmdutil.ParseFile(destProject, args[1])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false, client.WithMaxConcurrentStreams(parallelism))
			if err != nil {
				return err
			}
			defer c.Close()

			var opts []client.CopyFileOption
			if appendFile {
				opts = append(opts, client.WithAppendCopyFile())
			}
			return c.CopyFile(
				destFile.Commit, destFile.Path,
				srcFile.Commit, srcFile.Path,
				opts...,
			)
		}),
	}
	copyFile.Flags().BoolVarP(&appendFile, "append", "a", false, "Append to the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.")
	copyFile.Flags().StringVar(&project, "project", project, "Project in which both source and destination repos are located.")
	copyFile.Flags().StringVar(&srcProject, "src-project", "", "Project in which the source repo is located. This overrides --project.")
	copyFile.Flags().StringVar(&destProject, "dest-project", "", "Project in which the destination repo is located. This overrides --project.")
	shell.RegisterCompletionFunc(copyFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(copyFile, "copy file", files))

	var outputPath string
	var offsetBytes int64
	var retry bool
	getFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return the contents of a file.",
		Long:  "Return the contents of a file.",
		Example: `
# get a single file "XXX" on branch "master" in repo "foo"
$ {{alias}} foo@master:XXX

# get file "XXX" in the parent of the current head of branch "master"
# in repo "foo"
$ {{alias}} foo@master^:XXX

# get file "XXX" in the grandparent of the current head of branch "master"
# in repo "foo"
$ {{alias}} foo@master^2:XXX

# get file "test[].txt" on branch "master" in repo "foo"
# the path is interpreted as a glob pattern: quote and protect regex characters
$ {{alias}} 'foo@master:/test\[\].txt'

# get all files under the directory "XXX" on branch "master" in repo "foo"
$ {{alias}} foo@master:XXX -r
`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			if !enableProgress {
				progress.Disable()
			}
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			c, err := newClient(mainCtx, pachctlCfg)
			if err != nil {
				return err
			}
			defer c.Close()
			defer progress.Wait()
			// TODO: Decide what progress should look like in the recursive case. The files are downloaded in a batch in 2.x.
			if recursive {
				if outputPath == "" {
					return errors.Errorf("an output path needs to be specified when using the --recursive flag")
				}
				// Check that the path matches one directory / file.
				fi, err := c.InspectFile(file.Commit, file.Path)
				if err != nil {
					return err
				}
				r, err := c.GetFileTAR(file.Commit, file.Path)
				if err != nil {
					return err
				}
				return tarutil.Import(outputPath, r, func(hdr *tar.Header) error {
					hdr.Name = strings.TrimPrefix(hdr.Name, fi.File.Path)
					return nil
				})
			}
			var w io.Writer
			// If an output path is given, print the output to stdout
			if outputPath == "" {
				w = os.Stdout
			} else {
				if url, err := url.Parse(outputPath); err == nil && url.Scheme != "" {
					return c.GetFileURL(file.Commit, file.Path, url.String())
				}
				fi, err := c.InspectFile(file.Commit, file.Path)
				if err != nil {
					return err
				}
				var f *progress.File
				if ofi, err := os.Stat(outputPath); retry && err == nil {
					// when retrying, just write the unwritten bytes
					if offsetBytes == 0 {
						offsetBytes = ofi.Size()
					}
					f, err = progress.OpenAppend(outputPath, int64(fi.SizeBytes)-offsetBytes)
					if err != nil {
						return err
					}
				} else {
					f, err = progress.Create(outputPath, int64(fi.SizeBytes)-offsetBytes)
					if err != nil {
						return err
					}
				}
				defer f.Close()
				w = f
			}
			if err := c.GetFile(file.Commit, file.Path, w, client.WithOffset(offsetBytes)); err != nil {
				msg := err.Error()
				if strings.Contains(msg, pfsserver.GetFileTARSuggestion) {
					err = errors.New(strings.ReplaceAll(msg, pfsserver.GetFileTARSuggestion, "Try again with the -r flag"))
				}
				return errors.Wrapf(err, "couldn't download %s from %s", file.Path, file.Commit)
			}
			return nil
		}),
	}
	getFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Download multiple files, or recursively download a directory.")
	getFile.Flags().StringVarP(&outputPath, "output", "o", "", "The path where data will be downloaded.")
	getFile.Flags().BoolVar(&enableProgress, "progress", isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()), "{true|false} Whether or not to print the progress bars.")
	getFile.Flags().Int64Var(&offsetBytes, "offset", 0, "The number of bytes in the file to skip ahead when reading.")
	getFile.Flags().BoolVar(&retry, "retry", false, "{true|false} Whether to append the missing bytes to an existing file. No-op if the file doesn't exist.")
	getFile.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(getFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(getFile, "get file", files))

	inspectFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			fileInfo, err := c.InspectFile(file.Commit, file.Path)
			if err != nil {
				return err
			}
			if fileInfo == nil {
				return errors.Errorf("file %s not found", file.Path)
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(fileInfo))
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			return pretty.PrintDetailedFileInfo(fileInfo)
		}),
	}
	inspectFile.Flags().AddFlagSet(outputFlags)
	inspectFile.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(inspectFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectFile, "inspect file", files))

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

# list file under directory "dir[1]" on branch "master" in repo "foo"
# : quote and protect regex characters
$ {{alias}} 'foo@master:dir\[1\]'`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				return c.ListFile(file.Commit, file.Path, func(fi *pfs.FileInfo) error {
					return errors.EnsureStack(encoder.EncodeProto(fi))
				})
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			header := pretty.FileHeader
			writer := tabwriter.NewWriter(os.Stdout, header)
			if err := c.ListFile(file.Commit, file.Path, func(fi *pfs.FileInfo) error {
				pretty.PrintFileInfo(writer, fi, fullTimestamps, false)
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	listFile.Flags().AddFlagSet(outputFlags)
	listFile.Flags().AddFlagSet(timestampFlags)
	listFile.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(listFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(listFile, "list file", files))

	globFile := &cobra.Command{
		Use:   `{{alias}} "<repo>@<branch-or-commit>:<pattern>"`,
		Short: "Return files that match a glob pattern in a commit.",
		Long:  "Return files that match a glob pattern in a commit (that is, match a glob pattern in a repo at the state represented by a commit). Glob patterns are documented [here](https://golang.org/pkg/path/filepath/#Match).",
		Example: `
# Return files in repo "foo" on branch "master" that start
# with the character "A".  Note how the double quotation marks around the
# parameter are necessary because otherwise your shell might interpret the "*".
$ {{alias}} "foo@master:A*"

# Return files in repo "foo" on branch "master" under directory "data".
$ {{alias}} "foo@master:data/*"

# If you only want to view all files on a given repo branch, use "list file -f <repo>@<branch>" instead.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			fileInfos, err := c.GlobFileAll(file.Commit, file.Path)
			if err != nil {
				return err
			}
			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				for _, fileInfo := range fileInfos {
					if err := encoder.EncodeProto(fileInfo); err != nil {
						return errors.EnsureStack(err)
					}
				}
				return nil
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo, fullTimestamps, false)
			}
			return writer.Flush()
		}),
	}
	globFile.Flags().AddFlagSet(outputFlags)
	globFile.Flags().AddFlagSet(timestampFlags)
	globFile.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(globFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(globFile, "glob file", files))

	var shallow bool
	var nameOnly bool
	var diffCmdArg string
	var oldProject string
	diffFile := &cobra.Command{
		Use:   "{{alias}} <new-repo>@<new-branch-or-commit>:<new-path> [<old-repo>@<old-branch-or-commit>:<old-path>]",
		Short: "Return a diff of two file trees stored in Pachyderm",
		Long:  "Return a diff of two file trees stored in Pachyderm",
		Example: `
# Return the diff of the file "path" of the repo "foo" between the head of the
# "master" branch and its parent.
$ {{alias}} foo@master:path

# Return the diff between the master branches of repos foo and bar at paths
# path1 and path2, respectively.
$ {{alias}} foo@master:path1 bar@master:path2`,
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			newFile, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			if oldProject == "" {
				oldProject = project
			}
			oldFile := &pfs.File{}
			if len(args) == 2 {
				oldFile, err = cmdutil.ParseFile(oldProject, args[1])
				if err != nil {
					return err
				}
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			return pager.Page(noPager, os.Stdout, func(w io.Writer) (retErr error) {
				var writer *tabwriter.Writer
				if nameOnly {
					writer = tabwriter.NewWriter(w, pretty.DiffFileHeader)
					defer func() {
						if err := writer.Flush(); err != nil && retErr == nil {
							retErr = err
						}
					}()
				}

				newFiles, oldFiles, err := c.DiffFileAll(
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
					if nFI != nil && nFI.FileType == pfs.FileType_FILE {
						nPath, err = dlFile(c, nFI.File)
						if err != nil {
							return err
						}
						defer func() {
							if err := os.RemoveAll(nPath); err != nil && retErr == nil {
								retErr = err
							}
						}()
					}
					if oFI != nil && oFI.FileType == pfs.FileType_FILE {
						oPath, err = dlFile(c, oFI.File)
						defer func() {
							if err := os.RemoveAll(oPath); err != nil && retErr == nil {
								retErr = err
							}
						}()
					}
					cmd := exec.Command(diffCmd[0], append(diffCmd[1:], oPath, nPath)...)
					cmd.Stdout = w
					cmd.Stderr = os.Stderr
					// Diff returns exit code 1 when it finds differences
					// between the files, so we catch it.
					if err := cmd.Run(); err != nil && cmd.ProcessState.ExitCode() != 1 {
						return errors.EnsureStack(err)
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
	diffFile.Flags().StringVar(&project, "project", project, "Project in which first repo is located.")
	diffFile.Flags().StringVar(&oldProject, "old-project", "", "Project in which second, older repo is located.")
	shell.RegisterCompletionFunc(diffFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(diffFile, "diff file", files))

	deleteFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(project, args[0])
			if err != nil {
				return err
			}
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()

			var opts []client.DeleteFileOption
			if recursive {
				opts = append(opts, client.WithRecursiveDeleteFile())
			}
			return c.DeleteFile(file.Commit, file.Path, opts...)
		}),
	}
	deleteFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively delete the files in a directory.")
	deleteFile.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	shell.RegisterCompletionFunc(deleteFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteFile, "delete file", files))

	objectDocs := &cobra.Command{
		Short: "Docs for objects.",
		Long: `Objects are content-addressed blobs of data that are directly stored in the backend object store.

Objects are a low-level resource and should not be accessed directly by most users.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(objectDocs, "object", " object$"))

	var fix bool
	var zombie string
	var zombieAll bool
	fsck := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Run a file system consistency check on pfs.",
		Long:  "Run a file system consistency check on the pachyderm file system, ensuring the correct provenance relationships are satisfied.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer c.Close()
			foundErrors := false
			var opts []client.FsckOption
			if zombieAll {
				if zombie != "" {
					return errors.New("either check all pipelines for zombie files or provide a single commit")
				}
				opts = append(opts, client.WithZombieCheckAll())
			} else if zombie != "" {
				commit, err := cmdutil.ParseCommit(project, zombie)
				if err != nil {
					return err
				}
				if commit.ID == "" && commit.Branch.Name == "" {
					return errors.Errorf("provide a specific commit or branch for zombie detection on %s", commit.Branch.Repo)
				}
				opts = append(opts, client.WithZombieCheckTarget(commit))
			}

			if err = c.Fsck(fix, func(resp *pfs.FsckResponse) error {
				if resp.Error != "" {
					foundErrors = true
					fmt.Printf("Error: %s\n", resp.Error)
				} else {
					fmt.Printf("Fix applied: %v", resp.Fix)
				}
				return nil
			}, opts...); err != nil {
				return err
			}
			if !foundErrors {
				fmt.Println("No errors found.")
			}
			return nil
		}),
	}
	fsck.Flags().BoolVarP(&fix, "fix", "f", false, "Attempt to fix as many issues as possible.")
	fsck.Flags().BoolVar(&zombieAll, "zombie-all", false, "Check all pipelines for zombie files: files corresponding to old inputs that were not properly deleted")
	fsck.Flags().StringVar(&zombie, "zombie", "", "A single commit to check for zombie files")
	fsck.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	commands = append(commands, cmdutil.CreateAlias(fsck, "fsck"))

	var branchStr string
	var seed int64
	var stateID string
	runLoadTest := &cobra.Command{
		Use:     "{{alias}} <spec-file>",
		Short:   "Run a PFS load test.",
		Long:    "Run a PFS load test.",
		Example: pfsload.LoadSpecification,
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer func() {
				if err := c.Close(); retErr == nil {
					retErr = err
				}
			}()
			if len(args) == 0 {
				resp, err := c.PfsAPIClient.RunLoadTestDefault(c.Ctx(), &types.Empty{})
				if err != nil {
					return errors.EnsureStack(err)
				}
				resp.Spec = ""
				if err := cmdutil.Encoder(output, os.Stdout).EncodeProto(resp); err != nil {
					return errors.EnsureStack(err)
				}
				fmt.Println()
				return nil
			}
			err = filepath.Walk(args[0], func(file string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fi.IsDir() {
					return nil
				}
				spec, err := os.ReadFile(file)
				if err != nil {
					return errors.EnsureStack(err)
				}
				var branch *pfs.Branch
				if branchStr != "" {
					branch, err = cmdutil.ParseBranch(project, branchStr)
					if err != nil {
						return err
					}
				}
				resp, err := c.PfsAPIClient.RunLoadTest(c.Ctx(), &pfs.RunLoadTestRequest{
					Spec:    string(spec),
					Branch:  branch,
					Seed:    seed,
					StateId: stateID,
				})
				if err != nil {
					return errors.EnsureStack(err)
				}
				resp.Spec = ""
				if err := cmdutil.Encoder(output, os.Stdout).EncodeProto(resp); err != nil {
					return errors.EnsureStack(err)
				}
				fmt.Println()
				return nil
			})
			return errors.EnsureStack(err)
		}),
	}
	runLoadTest.Flags().StringVarP(&branchStr, "branch", "b", "", "The branch to use for generating the load.")
	runLoadTest.Flags().Int64VarP(&seed, "seed", "s", 0, "The seed to use for generating the load.")
	runLoadTest.Flags().StringVar(&project, "project", project, "Project in which repo is located.")
	runLoadTest.Flags().StringVar(&stateID, "state-id", "", "The ID of the base state to use for the load.")
	commands = append(commands, cmdutil.CreateAlias(runLoadTest, "run pfs-load-test"))

	// Add the mount commands (which aren't available on Windows, so they're in
	// their own file)
	commands = append(commands, mountCmds(mainCtx, pachctlCfg)...)

	return commands
}

func putFileHelper(mf client.ModifyFile, path, source string, recursive, appendFile, untar bool) (retErr error) {
	// Resolve the path and convert to unix path in case we're on windows.
	path = filepath.ToSlash(filepath.Clean(path))
	var opts []client.PutFileOption
	if appendFile {
		opts = append(opts, client.WithAppendPutFile())
	}
	// try parsing the filename as a url, if it is one do a PutFileURL
	if url, err := url.Parse(source); err == nil && url.Scheme != "" {
		return errors.EnsureStack(mf.PutFileURL(path, url.String(), recursive, opts...))
	}
	if source == "-" {
		if recursive {
			return errors.New("cannot set -r and read from stdin (must also set -f or -i)")
		}
		stdin := progress.Stdin()
		defer stdin.Finish()
		return errors.EnsureStack(mf.PutFile(path, stdin, opts...))
	}
	// Resolve the source and convert to unix path in case we're on windows.
	source = filepath.ToSlash(filepath.Clean(source))
	if recursive {
		err := filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
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
			return putFileHelper(mf, childDest, filePath, false, appendFile, untar)
		})
		return errors.EnsureStack(err)
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
	if untar {
		switch {
		case strings.HasSuffix(source, ".tar"):
			return errors.EnsureStack(mf.PutFileTAR(f, opts...))
		case strings.HasSuffix(source, ".tar.gz"), strings.HasSuffix(source, ".tgz"):
			r, err := gzip.NewReader(f)
			if err != nil {
				return errors.EnsureStack(err)
			}
			defer func() {
				if err := r.Close(); retErr == nil {
					retErr = err
				}
			}()
			return errors.EnsureStack(mf.PutFileTAR(r, opts...))
		}
	}
	return errors.EnsureStack(mf.PutFile(path, f, opts...))
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
		return "", errors.EnsureStack(err)
	}
	file, err := os.CreateTemp(tempDir, filepath.Base(f.Path+"_"))
	if err != nil {
		return "", errors.EnsureStack(err)
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

func newClient(ctx context.Context, pachctlCfg *pachctl.Config, options ...client.Option) (*client.APIClient, error) {
	if inWorkerStr, ok := os.LookupEnv("PACH_IN_WORKER"); ok {
		inWorker, err := strconv.ParseBool(inWorkerStr)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't parse PACH_IN_WORKER")
		}
		if inWorker {
			return pachctlCfg.NewInWorker(ctx, options...)
		}
	}
	return pachctlCfg.NewOnUserMachine(ctx, false, options...)
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

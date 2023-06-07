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
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

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

	// REPO COMMANDS

	repoDocs := &cobra.Command{
		Short: "Docs for repos.",
		Long: "A repo (repository) is the top-level data object in Pachyderm and contains a collection of files, directories, and commits that are versioned-controlled. Repos exist under a Project, which is the top-level organizational object in Pachyderm. Files stored in an input repo can be of any type or size. " +
			"After you create a repo using `pachctl create repo`, you can put files into it using `pachctl put file`." +
			"To transform the files in a repo, create a pipeline using `pachctl create pipeline` and pass in a specification that references your repo, user code, and other configurations. " +
			"See Pipeline Specification documentation for more information: https://docs.pachyderm.com/latest/build-dags/pipeline-spec/",
	}
	commands = append(commands, cmdutil.CreateDocsAliases(repoDocs, "repo", " repo$", repos))

	var description string
	var allProjects bool
	project := pachCtx.Project
	createRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Create a new repo in your active project.",
		Long: "This command creates a repo in the project that is set to your active context (initially the `default` project).\n" +
			"\n" +
			"\t- To specify which project to create the repo in, use the `--project` flag \n" +
			"\t- To add a description to the project, use the `--description` flag  \n",
		Example: "\t- {{alias}} foo ➔ /<active-project>/foo \n" +
			"\t- {{alias}} bar --description 'my new repo' ➔ /<active-project>/bar \n" +
			"\t- {{alias}} baz --project myproject ➔ /myproject/baz \n",

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
	createRepo.Flags().StringVarP(&description, "description", "d", "", "Set a repo description.")
	createRepo.Flags().StringVar(&project, "project", project, "Specify an existing project (by name) for where the new repo will be located.")
	commands = append(commands, cmdutil.CreateAliases(createRepo, "create repo", repos))

	updateRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Update a repo.",
		Long: "This command enables you to update the description of an existing repo. \n" +
			"\n" +
			"\t- To specify which project to update the repo in, use the `--project` flag \n" +
			"\t- To update the description of a repo, use the `--description` flag \n" +
			"\n" +
			"If you are looking to update the pipelines in your repo, see `pachctl update pipeline` instead.",
		Example: "\t- {{alias}} foo --description 'my updated repo description'\n" +
			"\t- {{alias}} foo --project bar --description 'my updated repo description'\n",

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
	updateRepo.Flags().StringVarP(&description, "description", "d", "", "Set a repo description.")
	updateRepo.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo is located.")
	shell.RegisterCompletionFunc(updateRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(updateRepo, "update repo", repos))

	inspectRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return info about a repo.",
		Long: "This command returns details of the repo such as: `Name`, `Description`, `Created`, and `Size of HEAD on Master`. By default, PachCTL checks for a matching repo in the project that is set to your active context (initially the `default` project).\n" +
			"\n" +
			"\t- To specify the project containing the repo you want to inspect, use the `--project` flag \n",
		Example: "\t- {{alias}} foo  \n" +
			"\t- {{alias}} foo --project myproject",
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
	inspectRepo.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo is located.")
	shell.RegisterCompletionFunc(inspectRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectRepo, "inspect repo", repos))

	var all bool
	var repoType string
	listRepo := &cobra.Command{
		Short: "Return a list of repos.",
		Long: "This command returns a list of repos that you have permissions to view. By default, it does not show system repos like pipeline metadata. \n" +
			"\n" +
			"\t- To view all input repos across projects, use `-A` flag \n" +
			"\t- To view all repos in a specific project, including system repos, use the `--all` flag \n" +
			"\t- To view repos of a specific type, use the `--type` flag; options include `USER`, `META`, & `SPEC` \n" +
			"\t- To view repos of a specific type across projects, use the `--type` and `-A` flags \n" +
			"\t- To view repos of a specific type in a specific project, use the `--type` and `--project` flags \n" +
			"\n" +
			"For information on roles and permissions, see the documentation: https://docs.pachyderm.com/latest/set-up/authorization/permissions/",
		Example: "\t- pachctl list repos \n" +
			"\t- {{alias}} -A \n" +
			"\t- {{alias}} --all \n" +
			"\t- {{alias}} --type user \n" +
			"\t- {{alias}} --type user --all \n" +
			"\t- {{alias}} --type user --all --project default \n" +
			"\t- {{alias}} --type user --all --project default --raw",
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
	listRepo.Flags().BoolVar(&all, "all", false, "Specify all repo types should be returned, including hidden system repos.")
	listRepo.Flags().StringVar(&repoType, "type", pfs.UserRepoType, "Set a repo type for scoped results.")
	listRepo.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repos are located.")
	listRepo.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "Specify all repos across all projects should be returned.")
	commands = append(commands, cmdutil.CreateAliases(listRepo, "list repo", repos))

	var force bool
	deleteRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Delete a repo.",
		Long: "This command deletes a repo. If this is a shared resource, it will be deleted for other users as well. \n" +
			"\n" +
			"\t- To force delete a repo, use the `--force` flag; use with caution \n" +
			"\t- To delete all repos across all projects, use the `--all` flag \n" +
			"\t- To delete a repo of a specific type, use the `--type` flag; options include `USER`, `META`, & `SPEC` \n" +
			"\t- To delete all repos of a specific type across all projects, use the `--all` and `--type` flags \n" +
			"\t- To delete all repos of a specific type in a specific project, use the `--all`, `--type`, and `--project` flags \n" +
			"\n",
		Example: "\t- {{alias}} foo \n" +
			"\t- {{alias}} foo --force \n" +
			"\t- {{alias}} --type user \n" +
			"\t- {{alias}} --all \n" +
			"\t- {{alias}} --all --type user \n" +
			"\t- {{alias}} --all --type user --project default",

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
	deleteRepo.Flags().BoolVarP(&force, "force", "f", false, "Force repo deletion, regardless of errors; use with caution.")
	deleteRepo.Flags().BoolVar(&all, "all", false, "remove all repos")
	deleteRepo.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the to-be-deleted repo is located.")
	deleteRepo.Flags().BoolVarP(&allProjects, "all-projects", "A", false, "Delete repo(s) across all projects; only valid with --all.")
	shell.RegisterCompletionFunc(deleteRepo, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteRepo, "delete repo", repos))

	// COMMIT COMMANDS

	commitDocs := &cobra.Command{
		Short: "Docs for commits.",
		Long: "Commits are atomic transactions on the content of a repo.\n" +
			"\n" +
			"Creating a commit is a multi-step process: \n" +
			"\t1. Start a new commit with `pachctl start commit`\n" +
			"\t2. Write files to the commit via `pachctl put file`\n" +
			"\t3. Finish the new commit with `pachctl finish commit`\n" +
			"\n" +
			"Commits that have been started but not finished are NOT durable storage.\n" +
			"Commits become reliable (and immutable) when they are finished.\n" +
			"Commits can be created with another commit as a parent.",
	}
	commands = append(commands, cmdutil.CreateDocsAliases(commitDocs, "commit", " commit$", commits))

	var parent string
	startCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Start a new commit.",
		Long: "This command starts a new commit with parent-commit as the parent on the given branch; if the branch does not exist, it will be created. \n" +
			"\n" +
			"\t- To specify a parent commit, use the `--parent` flag \n" +
			"\t- To add a message to the commit, use the `--message` or `--description` flag \n" +
			"\t- To specify which project the repo is in, use the `--project` flag \n",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master -p 0001a0100b1c10d01111e001fg00h00i \n" +
			"\t- {{alias}} foo@master --message 'my commit message' \n" +
			"\t- {{alias}} foo@master --description 'my commit description' \n" +
			"\t- {{alias}} foo@master --project bar",

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
				fmt.Println(commit.Id)
			}
			return grpcutil.ScrubGRPC(err)
		}),
	}
	startCommit.Flags().StringVarP(&parent, "parent", "p", "", "Set the parent (by id) of the new commit; only needed when branch is not specified using the @ syntax.")
	startCommit.MarkFlagCustom("parent", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	startCommit.Flags().StringVarP(&description, "message", "m", "", "Set a description for the commit's contents (synonym for --description).")
	startCommit.Flags().StringVar(&description, "description", "d", "Set a description of this commit's contents (synonym for --message).")
	startCommit.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo for this commit is located.")
	shell.RegisterCompletionFunc(startCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(startCommit, "start commit", commits))

	finishCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Finish a started commit.",
		Long: "This command finishes a started commit. \n" +
			"\n" +
			"\t- To force finish a commit, use the `--force` flag \n" +
			"\t- To add a message to the commit, use the `--message` or `--description` flag \n" +
			"\t- To specify which project the repo is in, use the `--project` flag \n",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master --force \n" +
			"\t- {{alias}} foo@master --project bar" +
			"\t- {{alias}} foo@master --message 'my commit message' \n" +
			"\t- {{alias}} foo@master --description 'my commit description' --project bar \n",

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
	finishCommit.Flags().StringVarP(&description, "message", "m", "", "Set a description of this commit's contents; overwrites existing commit description (synonym for --description).")
	finishCommit.Flags().StringVar(&description, "description", "", "Set a description of this commit's contents; overwrites existing commit description (synonym for --message).")
	finishCommit.Flags().BoolVarP(&force, "force", "f", false, "Force finish commit, even if it has provenance, which could break jobs; prefer 'pachctl stop stop job'")
	finishCommit.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo for this commit is located.")
	shell.RegisterCompletionFunc(finishCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(finishCommit, "finish commit", commits))

	inspectCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Return info about a commit.",
		Long: "This command returns information about the commit, such as the commit location (`branch@commit-id`), originating branch, start/finish times, and size. \n" +
			"\n" +
			"\t- To view the raw details of the commit in JSON format, use the `--raw` flag \n" +
			"\t- To specify which project the repo is in, use the `--project` flag \n",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master --project bar \n" +
			"\t- {{alias}} foo@master --raw \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i --project bar --raw \n",

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
	inspectCommit.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo for this commit is located.")
	shell.RegisterCompletionFunc(inspectCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectCommit, "inspect commit", commits))

	var from string
	var number int64
	var originStr string
	var expand bool
	listCommit := &cobra.Command{
		Use:   "{{alias}} [<commit-id>|<repo>[@<branch-or-commit>]]",
		Short: "Return a list of commits.",
		Long: "This command returns a list of commits, either across the entire pachyderm cluster or restricted to a single repo. \n" +
			"\n" +
			"\t- To specify which project the repo is in, use the `--project` flag \n" +
			"\t- To specify the number of commits to return, use the `--number` flag \n" +
			"\t- To list all commits that have come after a certain commit, use the `--from` flag \n" +
			"\t- To specify the origin of the commit, use the `--origin` flag; options include `AUTO`, `FSCK`, & `USER`. Requires at least repo name in command \n" +
			"\t- To expand the commit to include all of its sub-commits, use the `--expand` flag \n",
		Example: "\t- {{alias}} foo ➔ returns all commits in repo foo \n" +
			"\t- {{alias}} foo@master ➔ returns all commits in repo foo on branch master \n" +
			"\t- {{alias}} foo@master --number 10 ➔ returns the last 10 commits in repo foo on branch master \n" +
			"\t- {{alias}} foo@master --from 0001a0100b1c10d01111e001fg00h00i ➔ returns all commits in repo foo on branch master since <commit> \n" +
			"\t- {{alias}} foo@master --origin user ➔ returns all commits in repo foo on branch master originating from \n" +
			"\t- {{alias}} 0001a0100b1c10d01111e001fg00h00i ➔ returns all commits with ID <commit-id> \n" +
			"\t- {{alias}} 0001a0100b1c10d01111e001fg00h00i --expand ➔ returns all sub-commits on new lines, along with columns of more information \n" +
			"\t- {{alias}} foo@master --raw -o yaml ➔ returns all commits in repo foo on branch master in YAML format \n",
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

				if toCommit.Id == "" && toCommit.Branch.Name == "" {
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
	listCommit.Flags().StringVarP(&from, "from", "f", "", "Set the starting point of the commit range to list.")
	listCommit.Flags().Int64VarP(&number, "number", "n", 0, "Set the limit of returned results; if zero, list all commits.")
	listCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	listCommit.Flags().BoolVar(&all, "all", false, "Specify all types of commits (AUTO, FSCK, USER) should be returned; default only includes USER.")
	listCommit.Flags().BoolVarP(&expand, "expand", "x", false, "Specify results should return one line for each sub-commit and include more columns.")
	listCommit.Flags().StringVar(&originStr, "origin", "", "Specify the type of commit to scope returned results by; options include AUTO, FSCK, & USER.")
	listCommit.Flags().AddFlagSet(outputFlags)
	listCommit.Flags().AddFlagSet(timestampFlags)
	listCommit.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the commit.")
	shell.RegisterCompletionFunc(listCommit, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(listCommit, "list commit", commits))

	waitCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Wait for the specified commit to finish and return it.",
		Long:  "This command waits for the specified commit to finish before returning it, allowing you to track your commits downstream as they are produced. Each line is printed as soon as a new (sub) commit of your global commit finishes.",
		Example: "\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i --project bar \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i --project bar --raw -o yaml \n",

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

			commitInfo, err := c.WaitCommit(commit.Branch.Repo.Project.GetName(), commit.Branch.Repo.Name, commit.Branch.Name, commit.Id)
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
	waitCommit.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the commit.")
	commands = append(commands, cmdutil.CreateAliases(waitCommit, "wait commit", commits))

	var newCommits bool
	subscribeCommit := &cobra.Command{
		Use:   "{{alias}} <repo>[@<branch>]",
		Short: "Print commits as they are created (finished).",
		Long: "This command prints commits as they are created in the specified repo and branch. By default, all existing commits on the specified branch are returned first.  A commit is only considered created when it's been finished." +
			"\n" +
			"\t- To only see commits created after a certain commit, use the `--from` flag. \n" +
			"\t- To only see new commits created from now on, use the `--new` flag. \n" +
			"\t- To see all commit types, use the `--all` flag.\n" +
			"\t- To only see commits of a specific type, use the `--origin` flag. \n",
		Example: "\t {{alias}} foo@master ➔ subscribe to commits in the foo repo on the master branch \n" +
			"\t- {{alias}} foo@bar --from 0001a0100b1c10d01111e001fg00h00i ➔ starting at <commit-id>, subscribe to commits in the foo repo on the master branch \n" +
			"\t- {{alias}} foo@bar --new ➔ subscribe to commits in the foo repo on the master branch, but only for new commits created from now on \n",
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
	subscribeCommit.Flags().StringVar(&from, "from", "", "Subscribe to and return all commits since the specified commit.")
	subscribeCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	subscribeCommit.Flags().BoolVar(&newCommits, "new", false, "Subscribe to and return only new commits created from now on.")
	subscribeCommit.Flags().BoolVar(&all, "all", false, "Specify results should return all types of commits (AUTO, FSCK, USER)")
	subscribeCommit.Flags().StringVar(&originStr, "origin", "", "Specify results should only return commits of a specific type; options include AUTO, FSCK, & USER.")
	subscribeCommit.Flags().AddFlagSet(outputFlags)
	subscribeCommit.Flags().AddFlagSet(timestampFlags)
	subscribeCommit.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing the commit.")
	shell.RegisterCompletionFunc(subscribeCommit, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(subscribeCommit, "subscribe commit", commits))

	squashCommit := &cobra.Command{
		Use:     "{{alias}} <commit-id>",
		Short:   "Squash the sub-commits of a commit.",
		Long:    "This command squashes the sub-commits of a commit.  The data in the sub-commits will remain in their child commits. The squash will fail if it includes a commit with no children",
		Example: "\t- {{alias}} 0001a0100b1c10d01111e001fg00h00i \n",

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
		Long: "This command deletes the sub-commits of a commit; data in sub-commits will be lost, so use with caution. " +
			"This operation is only supported if none of the sub-commits have children. ",
		Example: "\t- {{alias}} 0001a0100b1c10d01111e001fg00h00i",
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

	// BRANCH COMMANDS

	branchDocs := &cobra.Command{
		Short: "Docs for branches.",
		Long: "A branch in Pachyderm records provenance relationships between data in different repos, as well as being as an alias for a commit in its repo. " +
			"The branch reference will float to always refer to the latest commit on the branch, known as the HEAD commit. All commits are on exactly one branch. " +
			"Any pachctl command that can take a commit, can take a branch name instead.",
	}
	commands = append(commands, cmdutil.CreateDocsAliases(branchDocs, "branch", " branch$", branches))

	var branchProvenance cmdutil.RepeatedStringArg
	var head string
	trigger := &pfs.Trigger{}
	createBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Create a new branch, or update an existing branch, on a repo.",
		Long: "This command creates or updates a branch on a repo. \n" +
			"\n" +
			"\t- To create a branch for a repo in a particular project, use the `--project` flag; this requires the repo to already exist in that project \n" +
			"\t- To attach an existing commit as the head commit of the new branch, use the `--head` flag \n" +
			"\t- To set a trigger, use the `--trigger` flag, pass in the branch (from same repo, without `repo@`), and set conditions using any of the `--trigger` options \n" +
			"\t- To require all defined triggering conditions to be met, use the `--trigger-all` flag; otherwise, each condition can execute the trigger \n" +
			"\t- To attach provenance to the new branch, use the `--provenance` flag. You can inspect provenance using `pachctl inspect branch foo@bar` \n" +
			"\n" +
			"Note: Starting a commit on the branch also creates it, so there's often no need to call this.",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master --project bar \n" +
			"\t- {{alias}} foo@master --head 0001a0100b1c10d01111e001fg00h00i \n" +
			"\t- {{alias}} foo@master=0001a0100b1c10d01111e001fg00h00i \n" +
			"\t- {{alias}} foo@master --provenance=foo@branch1,foo@branch2 \n" +
			"\t- {{alias}} foo@master --trigger staging \n" +
			"\t- {{alias}} foo@master --trigger staging --trigger-size=100M \n" +
			"\t- {{alias}} foo@master --trigger staging --trigger-cron='@every 1h' \n" +
			"\t- {{alias}} foo@master --trigger staging --trigger-commits=10' \n" +
			"\t- {{alias}} foo@master --trigger staging --trigger-size=100M --trigger-cron='@every 1h --trigger-commits=10 --trigger-all \n",
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
			if (trigger.CronSpec != "" || trigger.Size != "" || trigger.Commits != 0) && trigger.Branch == "" {
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
	createBranch.Flags().VarP(&branchProvenance, "provenance", "p", "Set the provenance for the branch. format: <repo>@<branch>")
	createBranch.MarkFlagCustom("provenance", "__pachctl_get_repo_commit")
	createBranch.Flags().StringVarP(&head, "head", "", "", "Set the head of the newly created branch using <branch-or-commit> or <repo>@<branch>=<id>")
	createBranch.MarkFlagCustom("head", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	createBranch.Flags().StringVarP(&trigger.Branch, "trigger", "t", "", "Specify the branch name that triggers this branch.")
	createBranch.Flags().StringVar(&trigger.CronSpec, "trigger-cron", "", "Set a cron spec interval as a condition for the trigger.")
	createBranch.Flags().StringVar(&trigger.Size, "trigger-size", "", "Set data size as a condition for the trigger.")
	createBranch.Flags().Int64Var(&trigger.Commits, "trigger-commits", 0, "Set the number of commits as a condition for the trigger.")
	createBranch.Flags().BoolVar(&trigger.All, "trigger-all", false, "Specify that all set conditions must be met for the trigger.")
	createBranch.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo for this branch is located.")
	commands = append(commands, cmdutil.CreateAliases(createBranch, "create branch", branches))

	inspectBranch := &cobra.Command{
		Use:   "{{alias}}  <repo>@<branch>",
		Short: "Return info about a branch.",
		Long: "This command returns info about a branch, such as its `Name`, `Head Commit`, and `Trigger`. \n" +
			"\n" +
			"\t- To inspect a branch from a repo in another project, use the `--project` flag \n" +
			"\t- To get additional details about the branch, use the `--raw` flag \n",
		Example: "\t- {{alias}} foo@master  \n" +
			"\t- {{alias}} foo@master --project bar \n" +
			"\t- {{alias}} foo@master --raw \n",
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
	inspectBranch.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing branch's repo.")
	shell.RegisterCompletionFunc(inspectBranch, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectBranch, "inspect branch", branches))

	listBranch := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return all branches on a repo.",
		Long: "This command returns all branches on a repo. \n" +
			"\n" +
			"\t- To list branches from a repo in another project, use the `--project` flag \n" +
			"\t- To get additional details about the branches, use the `--raw` flag \n",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master --project bar \n" +
			"\t- {{alias}} foo@master --raw \n" +
			"\t- {{alias}} foo@master --raw -o yaml \n",
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
	listBranch.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing branch's repo.")
	shell.RegisterCompletionFunc(listBranch, shell.RepoCompletion)
	commands = append(commands, cmdutil.CreateAliases(listBranch, "list branch", branches))

	deleteBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Delete a branch",
		Long: "This command deletes a branch while leaving its commits intact. \n" +
			"\n" +
			"\t- To delete a branch from a repo in another project, use the `--project` flag \n" +
			"\t- To delete a branch regardless of errors, use the `--force` flag \n",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master --project bar \n" +
			"\t- {{alias}} foo@master --force \n" +
			"\t- {{alias}} foo@master --project bar --force \n",
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
	deleteBranch.Flags().BoolVarP(&force, "force", "f", false, "Force branch deletion regardless of errors; use with caution.")
	deleteBranch.Flags().StringVar(&project, "project", project, "Specify the project (by name) containing branch's repo.")
	shell.RegisterCompletionFunc(deleteBranch, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteBranch, "delete branch", branches))

	// COMMIT COMMANDS

	FindCommits := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs> [flags]",
		Short: "Find commits with reference to <filePath> within a branch starting from <repo@commitID>",
		Long: "This command returns a list of commits using a reference to their `file/path` within a branch, starting from `repo@<commitID>`. \n" +
			"\n" +
			"\t- To find commits from a repo in another project, use the `--project` flag \n" +
			"\t- To set a limit on the number of returned commits, use the `--limits` flag \n" +
			"\t- To set a timeout for your commit search, use the `--timeout` flag \n" +
			"\t- To print the results as json, use the `--json` flag \n",
		Example: "\t- {{alias}} foo@master:file \n" +
			"\t- {{alias}} foo@master:file --project bar \n" +
			"\t- {{alias}} foo@master:file --limit 10 \n" +
			"\t- {{alias}} foo@master:file --timeout 10s \n" +
			"\t- {{alias}} foo@master:file --json \n" +
			"\t- {{alias}} foo@master:file --project bar --json --limit 100 --timeout 20s \n",

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
	FindCommits.Flags().BoolVar(&jsonOutput, "json", jsonOutput, "Print the response in json.")
	FindCommits.Flags().Uint32Var(&limit, "limit", limit, "Set the number of matching commits to return.")
	FindCommits.Flags().DurationVar(&timeout, "timeout", timeout, "Set the search duration timeout.")
	FindCommits.Flags().StringVar(&project, "project", project, "Specify the project (by name) in which commits are located.")
	shell.RegisterCompletionFunc(FindCommits, shell.BranchCompletion)
	commands = append(commands, cmdutil.CreateAliases(FindCommits, "find commit", commits))

	// PROJECT COMMANDS

	projectDocs := &cobra.Command{
		Short: "Docs for projects.",
		Long: "Projects are the top-level organizational objects in Pachyderm. " +
			"Projects contain pachyderm data objects such as Repos and Pipelines.",
	}
	commands = append(commands, cmdutil.CreateDocsAliases(projectDocs, "project", " project", projects))

	createProject := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Create a new project.",
		Long: "This command creates a new project. \n" +
			"\n" +
			"\t- To set a description for the project, use the `--description` flag \n",
		Example: "\t- {{alias}} foo-project \n" +
			"\t- {{alias}} foo-project --description 'This is a project for foo.' \n",

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
	createProject.Flags().StringVarP(&description, "description", "d", "", "Set a description for the newly-created project.")
	commands = append(commands, cmdutil.CreateAliases(createProject, "create project", projects))

	updateProject := &cobra.Command{
		Use:     "{{alias}} <project>",
		Short:   "Update a project.",
		Long:    "This command updates a project's description.",
		Example: "\t- {{alias}} foo-project --description 'This is a project for foo.' \n",
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
	updateProject.Flags().StringVarP(&description, "description", "d", "", "Set a new description of the updated project.")
	shell.RegisterCompletionFunc(updateProject, shell.ProjectCompletion)
	commands = append(commands, cmdutil.CreateAliases(updateProject, "update project", projects))

	inspectProject := &cobra.Command{
		Use:   "{{alias}} <project>",
		Short: "Inspect a project.",
		Long: "This command inspects a project and returns information like its `Name` and `Created at` time. \n" +
			"\n" +
			"\t- To return additional details, use the `--raw` flag \n",

		Example: "\t- {{alias}} foo-project \n" +
			"\t- {{alias}} foo-project --raw \n" +
			"\t- {{alias}} foo-project --output=yaml \n",

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
		Long:  "This command returns all projects.",
		Example: "\t- {{alias}} \n" +
			"\t- {{alias}} --raw \n",
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
		Long:  "This command deletes a project.",
		Example: "\t- {{alias}} foo-project \n" +
			"\t- {{alias}} foo-project --force \n",
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
	deleteProject.Flags().BoolVarP(&force, "force", "f", false, "Force delete the project regardless of errors; use with caution.")
	commands = append(commands, cmdutil.CreateAliases(deleteProject, "delete project", projects))

	// FILE COMMANDS

	fileDocs := &cobra.Command{
		Short: "Docs for files.",
		Long: "Files are the lowest-level data objects in Pachyderm and can be of any type (e.g. csv, binary, images, etc) or size. \n" +
			"Files can be written to started--but not yet finished--commits with the `pachctl put file` command. To read a file from a commit, Files use the `pachctl get file` command.",
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
		Long: "This command puts a file into the filesystem.  This command supports a number of ways to insert data into PFS. \n" +
			"\n" +
			"Files, Directories, & URLs: \n" +
			"\t- To upload via local filesystem, use the `-f` flag \n" +
			"\t- To upload via URL, use the `-f` flag with a URL as the argument \n" +
			"\t- To upload via filepaths & urls within a file, use the `i` flag \n" +
			"\t- To upload to a specific path in the repo, use the `-f` flag and add the path to the `repo@branch:/path` \n" +
			"\t- To upload recursively from a directory, use the `-r` flag \n" +
			"\t- To upload tar files and have them automatically untarred, use the `-untar` flag \n" +
			"\n" +
			"Compression, Parallelization, Appends: \n" +
			"\t- To compress files before uploading, use the `-c` flag \n" +
			"\t- To define the maximum number of files that can be uploaded in parallel, use the `-p` flag \n" +
			"\t- To append to an existing file, use the `-a` flag \n" +
			"\n" +
			"Other: \n" +
			"\t- To enable progress bars, use the `-P` flag \n",

		Example: "\t- {{alias}} repo@master-f image.png \n" +
			"\t- {{alias}} repo@master:/logs/log-1.txt  \n" +
			"\t- {{alias}} -r repo@master -f my-directory \n" +
			"\t- {{alias}} -r repo@branch:/path -f my-directory \n" +
			"\t- {{alias}} repo@branch -f http://host/example.png \n" +
			"\t- {{alias}} repo@branch:/dir -f http://host/example.png \n" +
			"\t- {{alias}} repo@branch -r -f s3://my_bucket \n" +
			"\t- {{alias}} repo@branch -i file \n" +
			"\t- {{alias}} repo@branch -i http://host/path \n" +
			"\t- {{alias}} repo@branch -f -untar dir.tar \n" +
			"\t- {{alias}} repo@branch -f -c image.png \n",

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
	putFile.Flags().StringSliceVarP(&filePaths, "file", "f", []string{"-"}, "Specify the file to be put; it can be a local file or a URL.")
	putFile.Flags().StringVarP(&inputFile, "input-file", "i", "", "Specify file provided contains a list of files to be put (as paths or URLs).")
	putFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Specify files should be recursively put into a directory.")
	putFile.Flags().BoolVarP(&compress, "compress", "", false, "Specify data should be compressed during upload. This parameter might help you upload your uncompressed data, such as CSV files, to Pachyderm faster. Use 'compress' with caution, because if your data is already compressed, this parameter might slow down the upload speed instead of increasing.")
	putFile.Flags().IntVarP(&parallelism, "parallelism", "p", DefaultParallelism, "Set the maximum number of files that can be uploaded in parallel.")
	putFile.Flags().BoolVarP(&appendFile, "append", "a", false, "Specify file contents should be appended to existing content from previous commits or previous calls to 'pachctl put file' within this commit.")
	putFile.Flags().BoolVar(&enableProgress, "progress", isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()), "Print progress bars.")
	putFile.Flags().BoolVar(&fullPath, "full-path", false, "Specify entire path provided to -f should be the target filename in PFS; by default only the base of the path is used.")
	putFile.Flags().BoolVar(&untar, "untar", false, "Specify file(s) with the extension .tar should be untarred and put as a separate file for each file within the tar stream(s); gzipped (.tar.gz or .tgz) tar file(s) are handled as well")
	putFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo for uploading this file is located.")
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
		Long: "This command files between pfs paths. While using this command, take special note of which project is set to your active context by running `pachctl list projects` and checking for the `*` in the ACTIVE column. \n" +
			"\n" +
			"\t- To append to an existing file, use the --append flag.\n" +
			"\t- To specify the project where both the source and destination repos are located, use the --project flag. This is only necessary if the project in question is not set to your active context.\n" +
			"\t- To copy a file from one project to another, use the --src-project and --dest-project flags. Needing to use one (or both) depends on whether or not either project is set to your active context.\n",
		Example: "\t- {{alias}} foo@master:/file bar@master:/file \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i:/file bar@master:/file \n" +
			"\t- {{alias}} foo@master:/file bar@master:/file --project ProjectContainingFooAndBar \n" +
			"\t- {{alias}} foo@master:/file bar@master:/file --dest-project ProjectContainingBar \n" +
			"\t- {{alias}} foo@master:/file bar@master:/file --src-project ProjectContainingFoo \n" +
			"\t- {{alias}} foo@master:/file bar@master:/file --src-project ProjectContainingFoo --dest-project ProjectContainingBar",
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
	copyFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where both source and destination repos are located.")
	copyFile.Flags().StringVar(&srcProject, "src-project", "", "Specify the project (by name) where the source repo is located; this overrides --project.")
	copyFile.Flags().StringVar(&destProject, "dest-project", "", "Specify the project (by name) where the destination repo is located; this overrides --project.")
	shell.RegisterCompletionFunc(copyFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(copyFile, "copy file", files))

	var outputPath string
	var offsetBytes int64
	var retry bool
	getFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return the contents of a file.",
		Long: "This command returns the contents of a file. " +
			"While using this command, take special note of how you can use ancestry syntax (e.g., appending`^2` or `.-1` to `repo@branch`) to retrieve the contents of a file from a previous commit. \n" +
			"\n" +
			"\t- To specify the project where the repo is located, use the --project flag \n" +
			"\t- To specify the output path, use the --output flag \n" +
			"\t- To specify the number of bytes to offset the read by, use the --offset-bytes flag \n" +
			"\t- To retry the operation if it fails, use the --retry flag \n",

		Example: "\t- {{alias}} foo@master:image.png \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i:image.png \n" +
			"\t- {{alias}} foo@master:/directory -r \n" +
			"\t- {{alias}} foo@master:image.png --output /path/to/image.png \n" +
			"\t- {{alias}} foo@master:/logs/log.txt--offset-bytes 100 \n" +
			"\t- {{alias}} foo@master:image.png --retry \n" +
			"\t- {{alias}} foo@master:/logs/log.txt --output /path/to/image.png --offset-bytes 100 --retry \n" +
			"\t- {{alias}} foo@master^:chart.png \n" +
			"\t- {{alias}} foo@master^2:chart.png  \n" +
			"\t- {{alias}} foo@master.1:chart.png  \n" +
			"\t- {{alias}} foo@master.-1:chart.png  \n" +
			"\t- " + `{{alias}} 'foo@master:/test\[\].txt'` + "\n",
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
	getFile.Flags().StringVarP(&outputPath, "output", "o", "", "Set the path where data will be downloaded.")
	getFile.Flags().BoolVar(&enableProgress, "progress", isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()), "{true|false} Whether or not to print the progress bars.")
	getFile.Flags().Int64Var(&offsetBytes, "offset", 0, "Set the number of bytes in the file to skip ahead when reading.")
	getFile.Flags().BoolVar(&retry, "retry", false, "{true|false} Whether to append the missing bytes to an existing file. No-op if the file doesn't exist.")
	getFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the file's repo is located.")
	shell.RegisterCompletionFunc(getFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(getFile, "get file", files))

	inspectFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return info about a file.",
		Long: "This command returns info about a file." +
			"While using this command, take special note of how you can use ancestry syntax (e.g., appending`^2` or `.-1` to `repo@branch`) to inspect the contents of a file from a previous commit. \n" +
			"\t- To specify the project where the repo is located, use the --project flag \n",
		Example: "\t- {{alias}} repo@master:/logs/log.txt \n" +
			"\t- {{alias}} repo@0001a0100b1c10d01111e001fg00h00i:/logs/log.txt \n" +
			"\t- {{alias}} repo@master:/logs/log.txt^2 \n" +
			"\t- {{alias}} repo@master:/logs/log.txt.-1 \n" +
			"\t- {{alias}} repo@master:/logs/log.txt^2  --project foo \n",
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
	inspectFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the file's repo is located.")
	shell.RegisterCompletionFunc(inspectFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(inspectFile, "inspect file", files))

	listFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>[:<path/in/pfs>]",
		Short: "Return the files in a directory.",
		Long: "This command returns the files in a directory. " +
			"While using this command, take special note of how you can use ancestry syntax (e.g., appending`^2` or `.-1` to `repo@branch`) to inspect the contents of a file from a previous commit. \n" +
			"\t- To specify the project where the repo is located, use the --project flag \n",
		Example: "\t- {{alias}} foo@master \n" +
			"\t- {{alias}} foo@master:dir \n" +
			"\t- {{alias}} foo@master^ \n" +
			"\t- {{alias}} foo@master^2 \n" +
			"\t- {{alias}} repo@master.-2  --project foo \n" +
			"\t- " + `{{alias}} 'foo@master:dir\[1\]'` + "\n",
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
	listFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where repo is located.")
	shell.RegisterCompletionFunc(listFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(listFile, "list file", files))

	globFile := &cobra.Command{
		Use:   `{{alias}} "<repo>@<branch-or-commit>:<pattern>"`,
		Short: "Return files that match a glob pattern in a commit.",
		Long: "This command returns files that match a glob pattern in a commit (that is, match a glob pattern in a repo at the state represented by a commit). " +
			"Glob patterns are documented [here](https://golang.org/pkg/path/filepath/#Match). \n" +
			"\n" +
			"\t- To specify the project where the repo is located, use the `--project flag` \n",
		Example: "\t- " + `{{alias}} "foo@master:A*"` + "\n" +
			"\t- " + `{{alias}} "foo@0001a0100b1c10d01111e001fg00h00i:data/*"` + "\n" +
			"\t- " + `{{alias}} "foo@master:data/*"`,
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
	globFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo with potential matching file(s) is located.")
	shell.RegisterCompletionFunc(globFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(globFile, "glob file", files))

	var shallow bool
	var nameOnly bool
	var diffCmdArg string
	var oldProject string
	diffFile := &cobra.Command{
		Use:   "{{alias}} <new-repo>@<new-branch-or-commit>:<new-path> [<old-repo>@<old-branch-or-commit>:<old-path>]",
		Short: "Return a diff of two file trees stored in Pachyderm",
		Long: "This command returns a diff of two file trees stored in Pachyderm. The file trees are specified by two files, one from the new tree and one from the old tree." +
			" \n" +
			"\t- To specify the project where the repos are located, use the `--project flag` \n" +
			"\t- To specify the project where the second older repo is located, use the `--old-project flag` \n" +
			"\t- To prevent descending into sub-directories, use the `--shallow flag`\n" +
			"\t- To use an alternative (non-git) diff command, use the `--diff-command flag` \n" +
			"\t- To get only the names of changed files, use the `--name-only flag` \n",

		Example: "\t- {{alias}} foo@master:/logs/log.txt \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i:log.txt \n" +
			"\t- {{alias}} foo@master:path1 bar@master:path2",
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
	diffFile.Flags().BoolVarP(&shallow, "shallow", "s", false, "Specify results should not to descend into sub directories.")
	diffFile.Flags().BoolVar(&nameOnly, "name-only", false, "Specify results should only return the names of changed files.")
	diffFile.Flags().StringVar(&diffCmdArg, "diff-command", "", "Set a git-alternative program to diff files.")
	diffFile.Flags().AddFlagSet(timestampFlags)
	diffFile.Flags().AddFlagSet(pagerFlags)
	diffFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the first repo is located.")
	diffFile.Flags().StringVar(&oldProject, "old-project", "", "Specify the project (by name) where the second, older repo is located.")
	shell.RegisterCompletionFunc(diffFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(diffFile, "diff file", files))

	deleteFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Delete a file.",
		Long:  "This command deletes a file.",
		Example: "\t- {{alias}} foo@bar:image.png \n" +
			"\t- {{alias}} foo@0001a0100b1c10d01111e001fg00h00i:/images/image.png \n" +
			"\t- {{alias}} -r foo@master:/images \n" +
			"\t- {{alias}} -r foo@master:/images --project projectContainingFoo \n",
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
	deleteFile.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the file's repo is located.")
	shell.RegisterCompletionFunc(deleteFile, shell.FileCompletion)
	commands = append(commands, cmdutil.CreateAliases(deleteFile, "delete file", files))

	// OBJECT COMMANDS

	objectDocs := &cobra.Command{
		Short: "Docs for objects.",
		Long: "Objects are content-addressed blobs of data that are directly stored in the backend object store. \n" +
			"\n" +
			"Objects are a low-level resource and should not be accessed directly by most users.",
	}
	commands = append(commands, cmdutil.CreateDocsAlias(objectDocs, "object", " object$"))

	var fix bool
	var zombie string
	var zombieAll bool
	fsck := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Run a file system consistency check on PFS.",
		Long:  "This command runs a file system consistency check on the Pachyderm file system, ensuring the correct provenance relationships are satisfied.",
		Example: "\t- {{alias}} \n" +
			"\t- {{alias}} --fix \n" +
			"\t- {{alias}} --zombie-all \n" +
			"\t- {{alias}} --zombie foo@bar \n" +
			"\t- {{alias}} --zombie foo@bar --project projectContainingFoo \n",
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
				if commit.Id == "" && commit.Branch.Name == "" {
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
	fsck.Flags().BoolVar(&zombieAll, "zombie-all", false, "Check all pipelines for zombie files: files corresponding to old inputs that were not properly deleted.")
	fsck.Flags().StringVar(&zombie, "zombie", "", "Set a single commit (by id) to check for zombie files")
	fsck.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo of the branch/commit is located.")
	commands = append(commands, cmdutil.CreateAlias(fsck, "fsck"))

	var branchStr string
	var seed int64
	var stateID string
	runLoadTest := &cobra.Command{
		Use:     "{{alias}} <spec-file>",
		Short:   "Run a PFS load test.",
		Long:    "This command runs a PFS load test.",
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
				resp, err := c.PfsAPIClient.RunLoadTestDefault(c.Ctx(), &emptypb.Empty{})
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
	runLoadTest.Flags().StringVarP(&branchStr, "branch", "b", "", "Specify the branch to use for generating the load.")
	runLoadTest.Flags().Int64VarP(&seed, "seed", "s", 0, "Set the seed to use for generating the load.")
	runLoadTest.Flags().StringVar(&project, "project", project, "Specify the project (by name) where the repo is located.")
	runLoadTest.Flags().StringVar(&stateID, "state-id", "", "Set the ID of the base state to use for the load.")
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

package cmds

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	gosync "sync"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/tabwriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// DefaultParallelism is the default parallelism used by 'get file' and 'put file'.
	DefaultParallelism = 10
)

// Cmds returns a slice containing pfs commands.
func Cmds(noMetrics *bool, noPortForwarding *bool) []*cobra.Command {
	var commands []*cobra.Command

	raw := false
	rawFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	rawFlags.BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")

	fullTimestamps := false
	fullTimestampsFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	fullTimestampsFlags.BoolVar(&fullTimestamps, "full-timestamps", false, "Return absolute timestamps (as opposed to the default, relative timestamps).")

	marshaller := &jsonpb.Marshaler{Indent: "  "}

	repoDocs := &cobra.Command{
		Short: "Docs for repos.",
		Long: `Repos, short for repository, are the top level data objects in Pachyderm.

Repos contain version-controlled directories and files. Files can be of any size
or type (e.g. csv, binary, images, etc).`,
	}
	cmdutil.SetDocsUsage(repoDocs)
	commands = append(commands, cmdutil.CreateAlias(repoDocs, "repo"))

	var description string
	createRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()
			_, err = c.PfsAPIClient.CreateRepo(
				c.Ctx(),
				&pfsclient.CreateRepoRequest{
					Repo:        client.NewRepo(args[0]),
					Description: description,
				},
			)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createRepo.Flags().StringVarP(&description, "description", "d", "", "A description of the repo.")
	commands = append(commands, cmdutil.CreateAlias(createRepo, "create repo"))

	updateRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Update a repo.",
		Long:  "Update a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()
			_, err = c.PfsAPIClient.CreateRepo(
				c.Ctx(),
				&pfsclient.CreateRepoRequest{
					Repo:        client.NewRepo(args[0]),
					Description: description,
					Update:      true,
				},
			)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateRepo.Flags().StringVarP(&description, "description", "d", "", "A description of the repo.")
	commands = append(commands, cmdutil.CreateAlias(updateRepo, "update repo"))

	inspectRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()
			repoInfo, err := c.InspectRepo(args[0])
			if err != nil {
				return err
			}
			if repoInfo == nil {
				return fmt.Errorf("repo %s not found", args[0])
			}
			if raw {
				return marshaller.Marshal(os.Stdout, repoInfo)
			}
			ri := &pretty.PrintableRepoInfo{
				RepoInfo:       repoInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedRepoInfo(ri)
		}),
	}
	inspectRepo.Flags().AddFlagSet(rawFlags)
	inspectRepo.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectRepo, "inspect repo"))

	listRepo := &cobra.Command{
		Short: "Return all repos.",
		Long:  "Return all repos.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()
			repoInfos, err := c.ListRepo()
			if err != nil {
				return err
			}
			if raw {
				for _, repoInfo := range repoInfos {
					if err := marshaller.Marshal(os.Stdout, repoInfo); err != nil {
						return err
					}
				}
				return nil
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
	listRepo.Flags().AddFlagSet(rawFlags)
	listRepo.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(listRepo, "list repo"))

	var force bool
	var all bool
	deleteRepo := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			if len(args) > 0 && all {
				return fmt.Errorf("cannot use the --all flag with an argument")
			}
			if len(args) == 0 && !all {
				return fmt.Errorf("either a repo name or the --all flag needs to be provided")
			}
			if all {
				_, err = client.PfsAPIClient.DeleteRepo(client.Ctx(),
					&pfsclient.DeleteRepoRequest{
						Force: force,
						All:   all,
					})
			} else {
				err = client.DeleteRepo(args[0], force)
			}
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			return nil
		}),
	}
	deleteRepo.Flags().BoolVarP(&force, "force", "f", false, "remove the repo regardless of errors; use with care")
	deleteRepo.Flags().BoolVar(&all, "all", false, "remove all repos")
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
	cmdutil.SetDocsUsage(commitDocs)
	commands = append(commands, cmdutil.CreateAlias(commitDocs, "commit"))

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
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}
			cli, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer cli.Close()
			commit, err := cli.PfsAPIClient.StartCommit(cli.Ctx(),
				&pfsclient.StartCommitRequest{
					Branch:      branch.Name,
					Parent:      client.NewCommit(branch.Repo.Name, parent),
					Description: description,
				})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Println(commit.ID)
			return nil
		}),
	}
	startCommit.Flags().StringVarP(&parent, "parent", "p", "", "The parent of the new commit, unneeded if branch is specified and you want to use the previous head of the branch as the parent.")
	startCommit.MarkFlagCustom("parent", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	startCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents")
	startCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")
	commands = append(commands, cmdutil.CreateAlias(startCommit, "start commit"))

	finishCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			commit, err := cmdutil.ParseCommit(args[0])
			if err != nil {
				return err
			}
			cli, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer cli.Close()
			if description != "" {
				_, err := cli.PfsAPIClient.FinishCommit(cli.Ctx(),
					&pfsclient.FinishCommitRequest{
						Commit:      commit,
						Description: description,
					})
				return grpcutil.ScrubGRPC(err)
			}
			return cli.FinishCommit(commit.Repo.Name, commit.ID)
		}),
	}
	finishCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents (overwrites any existing commit description)")
	finishCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")
	commands = append(commands, cmdutil.CreateAlias(finishCommit, "finish commit"))

	inspectCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			commit, err := cmdutil.ParseCommit(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()

			commitInfo, err := client.InspectCommit(commit.Repo.Name, commit.ID)
			if err != nil {
				return err
			}
			if commitInfo == nil {
				return fmt.Errorf("commit %s not found", commit.ID)
			}
			if raw {
				return marshaller.Marshal(os.Stdout, commitInfo)
			}
			ci := &pretty.PrintableCommitInfo{
				CommitInfo:     commitInfo,
				FullTimestamps: fullTimestamps,
			}
			return pretty.PrintDetailedCommitInfo(ci)
		}),
	}
	inspectCommit.Flags().AddFlagSet(rawFlags)
	inspectCommit.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectCommit, "inspect commit"))

	var from string
	var number int
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
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()

			var to string
			if len(args) == 2 {
				to = args[1]
			}
			if raw {
				return c.ListCommitF(args[0], to, from, uint64(number), func(ci *pfsclient.CommitInfo) error {
					return marshaller.Marshal(os.Stdout, ci)
				})
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.CommitHeader)
			if err := c.ListCommitF(args[0], to, from, uint64(number), func(ci *pfsclient.CommitInfo) error {
				pretty.PrintCommitInfo(writer, ci, fullTimestamps)
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	listCommit.Flags().StringVarP(&from, "from", "f", "", "list all commits since this commit")
	listCommit.Flags().IntVarP(&number, "number", "n", 0, "list only this many commits; if set to zero, list all commits")
	listCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	listCommit.Flags().AddFlagSet(rawFlags)
	listCommit.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(listCommit, "list commit"))

	printCommitIter := func(commitIter client.CommitInfoIterator) error {
		if raw {
			for {
				commitInfo, err := commitIter.Next()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				if err := marshaller.Marshal(os.Stdout, commitInfo); err != nil {
					return err
				}
			}
		}
		writer := tabwriter.NewWriter(os.Stdout, pretty.CommitHeader)
		for {
			commitInfo, err := commitIter.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			pretty.PrintCommitInfo(writer, commitInfo, fullTimestamps)
		}
		return writer.Flush()
	}

	var repos cmdutil.RepeatedStringArg
	flushCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit> ...",
		Short: "Wait for all commits caused by the specified commits to finish and return them.",
		Long:  "Wait for all commits caused by the specified commits to finish and return them.",
		Example: `
# return commits caused by foo@XXX and bar@YYY
$ {{alias}} foo@XXX bar@YYY

# return commits caused by foo@XXX leading to repos bar and baz
$ {{alias}} foo@XXX -r bar -r baz`,
		Run: cmdutil.Run(func(args []string) error {
			commits, err := cmdutil.ParseCommits(args)
			if err != nil {
				return err
			}

			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()

			var toRepos []*pfsclient.Repo
			for _, repoName := range repos {
				toRepos = append(toRepos, client.NewRepo(repoName))
			}

			commitIter, err := c.FlushCommit(commits, toRepos)
			if err != nil {
				return err
			}

			return printCommitIter(commitIter)
		}),
	}
	flushCommit.Flags().VarP(&repos, "repos", "r", "Wait only for commits leading to a specific set of repos")
	flushCommit.MarkFlagCustom("repos", "__pachctl_get_repo")
	flushCommit.Flags().AddFlagSet(rawFlags)
	flushCommit.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(flushCommit, "flush commit"))

	var newCommits bool
	subscribeCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch>",
		Short: "Print commits as they are created (finished).",
		Long:  "Print commits as they are created in the specified repo and branch.  By default, all existing commits on the specified branch are returned first.  A commit is only considered 'created' when it's been finished.",
		Example: `
# subscribe to commits in repo "test" on branch "master"
$ {{alias}} test@master

# subscribe to commits in repo "test" on branch "master", but only since commit XXX.
$ {{alias}} test@master --from XXX

# subscribe to commits in repo "test" on branch "master", but only for new commits created from now on.
$ {{alias}} test@master --new`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer c.Close()

			if newCommits && from != "" {
				return fmt.Errorf("--new and --from cannot be used together")
			}

			if newCommits {
				from = branch.Name
			}

			commitIter, err := c.SubscribeCommit(branch.Repo.Name, branch.Name, from, pfsclient.CommitState_STARTED)
			if err != nil {
				return err
			}

			return printCommitIter(commitIter)
		}),
	}
	subscribeCommit.Flags().StringVar(&from, "from", "", "subscribe to all commits since this commit")
	subscribeCommit.MarkFlagCustom("from", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	subscribeCommit.Flags().BoolVar(&newCommits, "new", false, "subscribe to only new commits created from now on")
	subscribeCommit.Flags().AddFlagSet(rawFlags)
	subscribeCommit.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(subscribeCommit, "subscribe commit"))

	deleteCommit := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Delete an input commit.",
		Long:  "Delete an input commit. An input is a commit which is not the output of a pipeline.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			commit, err := cmdutil.ParseCommit(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.DeleteCommit(commit.Repo.Name, commit.ID)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(deleteCommit, "delete commit"))

	branchDocs := &cobra.Command{
		Short: "Docs for branches.",
		Long: `A branch in Pachyderm is an alias for a Commit ID.

The branch reference will "float" to always refer to the latest commit on the
branch, known as the HEAD commit. Not all commits must be on a branch and
multiple branches can refer to the same commit.

Any pachctl command that can take a Commit ID, can take a branch name instead.`,
	}
	cmdutil.SetDocsUsage(branchDocs)
	commands = append(commands, cmdutil.CreateAlias(branchDocs, "branch"))

	var branchProvenance cmdutil.RepeatedStringArg
	var head string
	createBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Create a new branch, or update an existing branch, on a repo.",
		Long:  "Create a new branch, or update an existing branch, on a repo, starting a commit on the branch will also create it, so there's often no need to call this.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}
			provenance, err := cmdutil.ParseBranches(branchProvenance)
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.CreateBranch(branch.Repo.Name, branch.Name, head, provenance)
		}),
	}
	createBranch.Flags().VarP(&branchProvenance, "provenance", "p", "The provenance for the branch. format: <repo>@<branch-or-commit>")
	createBranch.MarkFlagCustom("provenance", "__pachctl_get_repo_commit")
	createBranch.Flags().StringVarP(&head, "head", "", "", "The head of the newly created branch.")
	createBranch.MarkFlagCustom("head", "__pachctl_get_commit $(__parse_repo ${nouns[0]})")
	commands = append(commands, cmdutil.CreateAlias(createBranch, "create branch"))

	listBranch := &cobra.Command{
		Use:   "{{alias}} <repo>",
		Short: "Return all branches on a repo.",
		Long:  "Return all branches on a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			branches, err := client.ListBranch(args[0])
			if err != nil {
				return err
			}
			if raw {
				for _, branch := range branches {
					if err := marshaller.Marshal(os.Stdout, branch); err != nil {
						return err
					}
				}
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.BranchHeader)
			for _, branch := range branches {
				pretty.PrintBranch(writer, branch)
			}
			return writer.Flush()
		}),
	}
	listBranch.Flags().AddFlagSet(rawFlags)
	commands = append(commands, cmdutil.CreateAlias(listBranch, "list branch"))

	deleteBranch := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>",
		Short: "Delete a branch",
		Long:  "Delete a branch, while leaving the commits intact",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			branch, err := cmdutil.ParseBranch(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.DeleteBranch(branch.Repo.Name, branch.Name, force)
		}),
	}
	deleteBranch.Flags().BoolVarP(&force, "force", "f", false, "remove the branch regardless of errors; use with care")
	commands = append(commands, cmdutil.CreateAlias(deleteBranch, "delete branch"))

	fileDocs := &cobra.Command{
		Short: "Docs for files.",
		Long: `Files are the lowest level data objects in Pachyderm.

Files can be of any type (e.g. csv, binary, images, etc) or size and can be
written to started (but not finished) commits with 'put file'. Files can be read
from commits with 'get file'.`,
	}
	cmdutil.SetDocsUsage(fileDocs)
	commands = append(commands, cmdutil.CreateAlias(fileDocs, "file"))

	var filePaths []string
	var recursive bool
	var inputFile string
	var parallelism int
	var split string
	var targetFileDatums uint
	var targetFileBytes uint
	var headerRecords uint
	var putFileCommit bool
	var overwrite bool
	putFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>[:<path/in/pfs>]",
		Short: "Put a file into the filesystem.",
		Long:  "Put a file into the filesystem.  This supports a number of ways to insert data into pfs.",
		Example: `
# Put data from stdin as repo/branch/path:
$ echo "data" | {{alias}} repo branch path

# Put data from stdin as repo/branch/path and start / finish a new commit on the branch.
$ echo "data" | {{alias}} -c repo branch path

# Put a file from the local filesystem as repo/branch/path:
$ {{alias}} repo branch path -f file

# Put a file from the local filesystem as repo/branch/file:
$ {{alias}} repo branch -f file

# Put the contents of a directory as repo/branch/path/dir/file:
$ {{alias}} -r repo branch path -f dir

# Put the contents of a directory as repo/branch/dir/file:
$ {{alias}} -r repo branch -f dir

# Put the contents of a directory as repo/branch/file, i.e. put files at the top level:
$ {{alias}} -r repo branch / -f dir

# Put the data from a URL as repo/branch/path:
$ {{alias}} repo branch path -f http://host/path

# Put the data from a URL as repo/branch/path:
$ {{alias}} repo branch -f http://host/path

# Put the data from an S3 bucket as repo/branch/s3_object:
$ {{alias}} repo branch -r -f s3://my_bucket

# Put several files or URLs that are listed in file.
# Files and URLs should be newline delimited.
$ {{alias}} repo branch -i file

# Put several files or URLs that are listed at URL.
# NOTE this URL can reference local files, so it could cause you to put sensitive
# files into your Pachyderm cluster.
$ {{alias}} repo branch -i http://host/path`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user", client.WithMaxConcurrentStreams(parallelism))
			if err != nil {
				return err
			}
			defer c.Close()
			pfc, err := c.NewPutFileClient()
			if err != nil {
				return err
			}
			defer func() {
				if err := pfc.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			if putFileCommit {
				fmt.Fprintf(os.Stderr, "flag --commit / -c is deprecated; as of 1.7.2, you will get the same behavior without it\n")
			}

			limiter := limit.New(int(parallelism))
			var sources []string
			if inputFile != "" {
				// User has provided a file listing sources, one per line. Read sources
				var r io.Reader
				if inputFile == "-" {
					r = os.Stdin
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

			// Arguments parsed; create putFileHelper and begin copying data
			var eg errgroup.Group
			filesPut := &gosync.Map{}
			for _, source := range sources {
				source := source
				if file.Path == "" {
					// The user has not specified a path so we use source as path.
					if source == "-" {
						return fmt.Errorf("must specify filename when reading data from stdin")
					}
					eg.Go(func() error {
						return putFileHelper(c, pfc, file.Commit.Repo.Name, file.Commit.ID, joinPaths("", source), source, recursive, overwrite, limiter, split, targetFileDatums, targetFileBytes, headerRecords, filesPut)
					})
				} else if len(sources) == 1 {
					// We have a single source and the user has specified a path,
					// we use the path and ignore source (in terms of naming the file).
					eg.Go(func() error {
						return putFileHelper(c, pfc, file.Commit.Repo.Name, file.Commit.ID, file.Path, source, recursive, overwrite, limiter, split, targetFileDatums, targetFileBytes, headerRecords, filesPut)
					})
				} else {
					// We have multiple sources and the user has specified a path,
					// we use that path as a prefix for the filepaths.
					eg.Go(func() error {
						return putFileHelper(c, pfc, file.Commit.Repo.Name, file.Commit.ID, joinPaths(file.Path, source), source, recursive, overwrite, limiter, split, targetFileDatums, targetFileBytes, headerRecords, filesPut)
					})
				}
			}
			return eg.Wait()
		}),
	}
	putFile.Flags().StringSliceVarP(&filePaths, "file", "f", []string{"-"}, "The file to be put, it can be a local file or a URL.")
	putFile.Flags().StringVarP(&inputFile, "input-file", "i", "", "Read filepaths or URLs from a file.  If - is used, paths are read from the standard input.")
	putFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively put the files in a directory.")
	putFile.Flags().IntVarP(&parallelism, "parallelism", "p", DefaultParallelism, "The maximum number of files that can be uploaded in parallel.")
	putFile.Flags().StringVar(&split, "split", "", "Split the input file into smaller files, subject to the constraints of --target-file-datums and --target-file-bytes. Permissible values are `line`, `json`, `sql` and `csv`.")
	putFile.Flags().UintVar(&targetFileDatums, "target-file-datums", 0, "The upper bound of the number of datums that each file contains, the last file will contain fewer if the datums don't divide evenly; needs to be used with --split.")
	putFile.Flags().UintVar(&targetFileBytes, "target-file-bytes", 0, "The target upper bound of the number of bytes that each file contains; needs to be used with --split.")
	putFile.Flags().UintVar(&headerRecords, "header-records", 0, "the number of records that will be converted to a PFS 'header', and prepended to future retrievals of any subset of data from PFS; needs to be used with --split=(json|line|csv)")
	putFile.Flags().BoolVarP(&putFileCommit, "commit", "c", false, "DEPRECATED: Put file(s) in a new commit.")
	putFile.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.")
	commands = append(commands, cmdutil.CreateAlias(putFile, "put file"))

	copyFile := &cobra.Command{
		Use:   "{{alias}} <src-repo>@<src-branch-or-commit>:<src-path> <dst-repo>@<dst-branch-or-commit>:<dst-path>",
		Short: "Copy files between pfs paths.",
		Long:  "Copy files between pfs paths.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) (retErr error) {
			srcFile, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			destFile, err := cmdutil.ParseFile(args[1])
			if err != nil {
				return err
			}
			c, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user", client.WithMaxConcurrentStreams(parallelism))
			if err != nil {
				return err
			}
			defer c.Close()
			return c.CopyFile(
				srcFile.Commit.Repo.Name, srcFile.Commit.ID, srcFile.Path,
				destFile.Commit.Repo.Name, destFile.Commit.ID, destFile.Path,
				overwrite,
			)
		}),
	}
	copyFile.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite the existing content of the file, either from previous commits or previous calls to 'put file' within this commit.")
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
$ {{alias}} foo@master^2:XXX`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			if recursive {
				if outputPath == "" {
					return fmt.Errorf("an output path needs to be specified when using the --recursive flag")
				}
				puller := sync.NewPuller()
				return puller.Pull(client, outputPath, file.Commit.Repo.Name, file.Commit.ID, file.Path, false, false, parallelism, nil, "")
			}
			var w io.Writer
			// If an output path is given, print the output to stdout
			if outputPath == "" {
				w = os.Stdout
			} else {
				f, err := os.Create(outputPath)
				if err != nil {
					return err
				}
				defer f.Close()
				w = f
			}
			return client.GetFile(file.Commit.Repo.Name, file.Commit.ID, file.Path, 0, 0, w)
		}),
	}
	getFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively download a directory.")
	getFile.Flags().StringVarP(&outputPath, "output", "o", "", "The path where data will be downloaded.")
	getFile.Flags().IntVarP(&parallelism, "parallelism", "p", DefaultParallelism, "The maximum number of files that can be downloaded in parallel")
	commands = append(commands, cmdutil.CreateAlias(getFile, "get file"))

	inspectFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			fileInfo, err := client.InspectFile(file.Commit.Repo.Name, file.Commit.ID, file.Path)
			if err != nil {
				return err
			}
			if fileInfo == nil {
				return fmt.Errorf("file %s not found", file.Path)
			}
			if raw {
				return marshaller.Marshal(os.Stdout, fileInfo)
			}
			return pretty.PrintDetailedFileInfo(fileInfo)
		}),
	}
	inspectFile.Flags().AddFlagSet(rawFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectFile, "inspect file"))

	var history int64
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
$ {{alias}} foo@master --history -1`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			if raw {
				return client.ListFileF(file.Commit.Repo.Name, file.Commit.ID, file.Path, history, func(fi *pfsclient.FileInfo) error {
					return marshaller.Marshal(os.Stdout, fi)
				})
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
			if err := client.ListFileF(file.Commit.Repo.Name, file.Commit.ID, file.Path, history, func(fi *pfsclient.FileInfo) error {
				pretty.PrintFileInfo(writer, fi, fullTimestamps)
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	listFile.Flags().AddFlagSet(rawFlags)
	listFile.Flags().AddFlagSet(fullTimestampsFlags)
	listFile.Flags().Int64Var(&history, "history", 0, "Return revision history for files.")
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
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			fileInfos, err := client.GlobFile(file.Commit.Repo.Name, file.Commit.ID, file.Path)
			if err != nil {
				return err
			}
			if raw {
				for _, fileInfo := range fileInfos {
					if err := marshaller.Marshal(os.Stdout, fileInfo); err != nil {
						return err
					}
				}
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo, fullTimestamps)
			}
			return writer.Flush()
		}),
	}
	globFile.Flags().AddFlagSet(rawFlags)
	globFile.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(globFile, "glob file"))

	var shallow bool
	diffFile := &cobra.Command{
		Use:   "{{alias}} <new-repo>@<new-branch-or-commit>:<new-path> [<old-repo>@<old-branch-or-commit>:<old-path>]",
		Short: "Return a diff of two file trees.",
		Long:  "Return a diff of two file trees.",
		Example: `
# Return the diff of the file "path" of the repo "foo" between the head of the
# "master" branch and its parent.
$ {{alias}} foo@master:path

# Return the diff between the master branches of repos foo and bar at paths
# path1 and path2, respectively.
$ {{alias}} foo@master:path1 bar@master:path2`,
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			newFile, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			oldFile := client.NewFile("", "", "")
			if len(args) == 2 {
				oldFile, err = cmdutil.ParseFile(args[1])
				if err != nil {
					return err
				}
			}

			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()

			newFiles, oldFiles, err := client.DiffFile(
				newFile.Commit.Repo.Name, newFile.Commit.ID, newFile.Path,
				oldFile.Commit.Repo.Name, oldFile.Commit.ID, oldFile.Path,
				shallow,
			)
			if err != nil {
				return err
			}

			if len(newFiles) > 0 {
				fmt.Println("New Files:")
				writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
				for _, fileInfo := range newFiles {
					pretty.PrintFileInfo(writer, fileInfo, fullTimestamps)
				}
				if err := writer.Flush(); err != nil {
					return err
				}
			}
			if len(oldFiles) > 0 {
				fmt.Println("Old Files:")
				writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
				for _, fileInfo := range oldFiles {
					pretty.PrintFileInfo(writer, fileInfo, fullTimestamps)
				}
				if err := writer.Flush(); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	diffFile.Flags().BoolVarP(&shallow, "shallow", "s", false, "Specifies whether or not to diff subdirectories")
	diffFile.Flags().AddFlagSet(fullTimestampsFlags)
	commands = append(commands, cmdutil.CreateAlias(diffFile, "diff file"))

	deleteFile := &cobra.Command{
		Use:   "{{alias}} <repo>@<branch-or-commit>:<path/in/pfs>",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			file, err := cmdutil.ParseFile(args[0])
			if err != nil {
				return err
			}
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.DeleteFile(file.Commit.Repo.Name, file.Commit.ID, file.Path)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(deleteFile, "delete file"))

	objectDocs := &cobra.Command{
		Short: "Docs for objects.",
		Long: `Objects are content-addressed blobs of data that are directly stored in the backend object store.

Objects are a low-level resource and should not be accessed directly by most users.`,
	}
	cmdutil.SetDocsUsage(objectDocs)
	commands = append(commands, cmdutil.CreateAlias(objectDocs, "object"))

	getObject := &cobra.Command{
		Use:   "{{alias}} <hash>",
		Short: "Print the contents of an object.",
		Long:  "Print the contents of an object.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.GetObject(args[0], os.Stdout)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getObject, "get object"))

	tagDocs := &cobra.Command{
		Short: "Docs for tags.",
		Long: `Tags are aliases for objects. Many tags can refer to the same object.

Tags are a low-level resource and should not be accessed directly by most users.`,
	}
	cmdutil.SetDocsUsage(tagDocs)
	commands = append(commands, cmdutil.CreateAlias(tagDocs, "tag"))

	getTag := &cobra.Command{
		Use:   "{{alias}} <tag>",
		Short: "Print the contents of a tag.",
		Long:  "Print the contents of a tag.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.GetTag(args[0], os.Stdout)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getTag, "get tag"))

	var debug bool
	var commits cmdutil.RepeatedStringArg
	mount := &cobra.Command{
		Use:   "{{alias}} <path/to/mount/point>",
		Short: "Mount pfs locally. This command blocks.",
		Long:  "Mount pfs locally. This command blocks.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "fuse")
			if err != nil {
				return err
			}
			defer client.Close()
			mountPoint := args[0]
			commits, err := parseCommits(commits)
			if err != nil {
				return err
			}
			opts := &fuse.Options{
				Fuse: &nodefs.Options{
					Debug: debug,
				},
				Commits: commits,
			}
			return fuse.Mount(client, mountPoint, opts)
		}),
	}
	mount.Flags().BoolVarP(&debug, "debug", "d", false, "Turn on debug messages.")
	mount.Flags().VarP(&commits, "commits", "c", "Commits to mount for repos, arguments should be of the form \"repo@commit\"")
	mount.MarkFlagCustom("commits", "__pachctl_get_repo_branch")
	commands = append(commands, cmdutil.CreateAlias(mount, "mount"))

	unmount := &cobra.Command{
		Use:   "{{alias}} <path/to/mount/point>",
		Short: "Unmount pfs.",
		Long:  "Unmount pfs.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			if len(args) == 1 {
				return syscall.Unmount(args[0], 0)
			}

			if all {
				stdin := strings.NewReader(`
	mount | grep pfs:// | cut -f 3 -d " "
	`)
				var stdout bytes.Buffer
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stdout: &stdout,
					Stderr: os.Stderr,
				}, "sh"); err != nil {
					return err
				}
				scanner := bufio.NewScanner(&stdout)
				var mounts []string
				for scanner.Scan() {
					mounts = append(mounts, scanner.Text())
				}
				if len(mounts) == 0 {
					fmt.Println("No mounts found.")
					return nil
				}
				fmt.Printf("Unmount the following filesystems? yN\n")
				for _, mount := range mounts {
					fmt.Printf("%s\n", mount)
				}
				r := bufio.NewReader(os.Stdin)
				bytes, err := r.ReadBytes('\n')
				if err != nil {
					return err
				}
				if bytes[0] == 'y' || bytes[0] == 'Y' {
					for _, mount := range mounts {
						if err := syscall.Unmount(mount, 0); err != nil {
							return err
						}
					}
				}
			}
			return nil
		}),
	}
	unmount.Flags().BoolVarP(&all, "all", "a", false, "unmount all pfs mounts")
	commands = append(commands, cmdutil.CreateAlias(unmount, "unmount"))

	return commands
}

func parseCommits(args []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, arg := range args {
		split := strings.Split(arg, "@")
		if len(split) != 2 {
			return nil, fmt.Errorf("malformed input %s, must be of the form repo@commit", args)
		}
		result[split[0]] = split[1]
	}
	return result, nil
}

func putFileHelper(c *client.APIClient, pfc client.PutFileClient,
	repo, commit, path, source string, recursive, overwrite bool, // destination
	limiter limit.ConcurrencyLimiter,
	split string, targetFileDatums, targetFileBytes, headerRecords uint, // split
	filesPut *gosync.Map) (retErr error) {
	if _, ok := filesPut.LoadOrStore(path, nil); ok {
		return fmt.Errorf("multiple files put with the path %s, aborting, "+
			"some files may already have been put and should be cleaned up with "+
			"'delete file' or 'delete commit'", path)
	}
	putFile := func(reader io.ReadSeeker) error {
		if split == "" {
			if overwrite {
				return sync.PushFile(c, pfc, client.NewFile(repo, commit, path), reader)
			}
			_, err := pfc.PutFile(repo, commit, path, reader)
			return err
		}

		var delimiter pfsclient.Delimiter
		switch split {
		case "line":
			delimiter = pfsclient.Delimiter_LINE
		case "json":
			delimiter = pfsclient.Delimiter_JSON
		case "sql":
			delimiter = pfsclient.Delimiter_SQL
		case "csv":
			delimiter = pfsclient.Delimiter_CSV
		default:
			return fmt.Errorf("unrecognized delimiter '%s'; only accepts one of "+
				"{json,line,sql,csv}", split)
		}
		_, err := pfc.PutFileSplit(repo, commit, path, delimiter, int64(targetFileDatums), int64(targetFileBytes), int64(headerRecords), overwrite, reader)
		return err
	}

	if source == "-" {
		if recursive {
			return errors.New("cannot set -r and read from stdin (must also set -f or -i)")
		}
		limiter.Acquire()
		defer limiter.Release()
		fmt.Fprintln(os.Stderr, "Reading from stdin.")
		return putFile(os.Stdin)
	}
	// try parsing the filename as a url, if it is one do a PutFileURL
	if url, err := url.Parse(source); err == nil && url.Scheme != "" {
		limiter.Acquire()
		defer limiter.Release()
		return pfc.PutFileURL(repo, commit, path, url.String(), recursive, overwrite)
	}
	if recursive {
		var eg errgroup.Group
		if err := filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			// file doesn't exist
			if info == nil {
				return fmt.Errorf("%s doesn't exist", filePath)
			}
			if info.IsDir() {
				return nil
			}
			childDest := filepath.Join(path, strings.TrimPrefix(filePath, source))
			eg.Go(func() error {
				// don't do a second recursive 'put file', just put the one file at
				// filePath into childDest, and then this walk loop will go on to the
				// next one
				return putFileHelper(c, pfc, repo, commit, childDest, filePath, false,
					overwrite, limiter, split, targetFileDatums, targetFileBytes,
					headerRecords, filesPut)
			})
			return nil
		}); err != nil {
			return err
		}
		return eg.Wait()
	}
	limiter.Acquire()
	defer limiter.Release()
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return putFile(f)
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

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
)

const (
	codestart = "```sh\n\n"
	codeend   = "\n```"

	// DefaultParallelism is the default parallelism used by get-file
	// and put-file.
	DefaultParallelism = 10
)

// Cmds returns a slice containing pfs commands.
func Cmds(noMetrics *bool) []*cobra.Command {
	metrics := !*noMetrics
	raw := false
	rawFlag := func(cmd *cobra.Command) {
		cmd.Flags().BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")
	}
	marshaller := &jsonpb.Marshaler{Indent: "  "}

	repo := &cobra.Command{
		Use:   "repo",
		Short: "Docs for repos.",
		Long: `Repos, short for repository, are the top level data object in Pachyderm.

	Repos are created with create-repo.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	var description string
	createRepo := &cobra.Command{
		Use:   "create-repo repo-name",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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

	updateRepo := &cobra.Command{
		Use:   "update-repo repo-name",
		Short: "Update a repo.",
		Long:  "Update a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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

	inspectRepo := &cobra.Command{
		Use:   "inspect-repo repo-name",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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
			return pretty.PrintDetailedRepoInfo(repoInfo)
		}),
	}
	rawFlag(inspectRepo)

	listRepo := &cobra.Command{
		Use:   "list-repo",
		Short: "Return all repos.",
		Long:  "Return all repos.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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
				pretty.PrintRepoInfo(writer, repoInfo)
			}
			return writer.Flush()
		}),
	}
	rawFlag(listRepo)

	var force bool
	var all bool
	deleteRepo := &cobra.Command{
		Use:   "delete-repo repo-name",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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

	commit := &cobra.Command{
		Use:   "commit",
		Short: "Docs for commits.",
		Long: `Commits are atomic transactions on the content of a repo.

Creating a commit is a multistep process:
- start a new commit with start-commit
- write files to it through fuse or with put-file
- finish the new commit with finish-commit

Commits that have been started but not finished are NOT durable storage.
Commits become reliable (and immutable) when they are finished.

Commits can be created with another commit as a parent.
This layers the data in the commit over the data in the parent.
`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	var parent string
	startCommit := &cobra.Command{
		Use:   "start-commit repo-name [branch]",
		Short: "Start a new commit.",
		Long: `Start a new commit with parent-commit as the parent, or start a commit on the given branch; if the branch does not exist, it will be created.

Examples:

` + codestart + `# Start a new commit in repo "test" that's not on any branch
$ pachctl start-commit test

# Start a commit in repo "test" on branch "master"
$ pachctl start-commit test master

# Start a commit with "master" as the parent in repo "test", on a new branch "patch"; essentially a fork.
$ pachctl start-commit test patch -p master

# Start a commit with XXX as the parent in repo "test", not on any branch
$ pachctl start-commit test -p XXX
` + codeend,
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) error {
			cli, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			var branch string
			if len(args) == 2 {
				branch = args[1]
			}
			commit, err := cli.PfsAPIClient.StartCommit(cli.Ctx(),
				&pfsclient.StartCommitRequest{
					Branch:      branch,
					Parent:      client.NewCommit(args[0], parent),
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
	startCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents")
	startCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")

	finishCommit := &cobra.Command{
		Use:   "finish-commit repo-name commit-id",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			cli, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if description != "" {
				_, err := cli.PfsAPIClient.FinishCommit(cli.Ctx(),
					&pfsclient.FinishCommitRequest{
						Commit:      client.NewCommit(args[0], args[1]),
						Description: description,
					})
				return grpcutil.ScrubGRPC(err)
			}
			return cli.FinishCommit(args[0], args[1])
		}),
	}
	finishCommit.Flags().StringVarP(&description, "message", "m", "", "A description of this commit's contents (overwrites any existing commit description)")
	finishCommit.Flags().StringVar(&description, "description", "", "A description of this commit's contents (synonym for --message)")

	inspectCommit := &cobra.Command{
		Use:   "inspect-commit repo-name commit-id",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			commitInfo, err := client.InspectCommit(args[0], args[1])
			if err != nil {
				return err
			}
			if commitInfo == nil {
				return fmt.Errorf("commit %s not found", args[1])
			}
			if raw {
				return marshaller.Marshal(os.Stdout, commitInfo)
			}
			return pretty.PrintDetailedCommitInfo(commitInfo)
		}),
	}
	rawFlag(inspectCommit)

	var from string
	var number int
	listCommit := &cobra.Command{
		Use:   "list-commit repo-name",
		Short: "Return all commits on a set of repos.",
		Long: `Return all commits on a set of repos.

Examples:

` + codestart + `# return commits in repo "foo"
$ pachctl list-commit foo

# return commits in repo "foo" on branch "master"
$ pachctl list-commit foo master

# return the last 20 commits in repo "foo" on branch "master"
$ pachctl list-commit foo master -n 20

# return commits that are the ancestors of XXX
$ pachctl list-commit foo XXX

# return commits in repo "foo" since commit XXX
$ pachctl list-commit foo master --from XXX
` + codeend,
		Run: cmdutil.RunBoundedArgs(1, 2, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}

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
				pretty.PrintCommitInfo(writer, ci)
				return nil
			}); err != nil {
				return err
			}
			return writer.Flush()
		}),
	}
	listCommit.Flags().StringVarP(&from, "from", "f", "", "list all commits since this commit")
	listCommit.Flags().IntVarP(&number, "number", "n", 0, "list only this many commits; if set to zero, list all commits")
	rawFlag(listCommit)

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
			pretty.PrintCommitInfo(writer, commitInfo)
		}
		return writer.Flush()
	}

	var repos cmdutil.RepeatedStringArg
	flushCommit := &cobra.Command{
		Use:   "flush-commit commit [commit ...]",
		Short: "Wait for all commits caused by the specified commits to finish and return them.",
		Long: `Wait for all commits caused by the specified commits to finish and return them.

Examples:

` + codestart + `# return commits caused by foo/XXX and bar/YYY
$ pachctl flush-commit foo/XXX bar/YYY

# return commits caused by foo/XXX leading to repos bar and baz
$ pachctl flush-commit foo/XXX -r bar -r baz
` + codeend,
		Run: cmdutil.Run(func(args []string) error {
			commits, err := cmdutil.ParseCommits(args)
			if err != nil {
				return err
			}

			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}

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
	rawFlag(flushCommit)

	var new bool
	subscribeCommit := &cobra.Command{
		Use:   "subscribe-commit repo branch",
		Short: "Print commits as they are created (finished).",
		Long: `Print commits as they are created in the specified repo and
branch.  By default, all existing commits on the specified branch are
returned first.  A commit is only considered "created" when it's been
finished.

Examples:

` + codestart + `# subscribe to commits in repo "test" on branch "master"
$ pachctl subscribe-commit test master

# subscribe to commits in repo "test" on branch "master", but only since commit XXX.
$ pachctl subscribe-commit test master --from XXX

# subscribe to commits in repo "test" on branch "master", but only for new
# commits created from now on.
$ pachctl subscribe-commit test master --new
` + codeend,
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			repo, branch := args[0], args[1]
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}

			if new && from != "" {
				return fmt.Errorf("--new and --from cannot both be provided")
			}

			if new {
				from = branch
			}

			commitIter, err := c.SubscribeCommit(repo, branch, from, pfsclient.CommitState_STARTED)
			if err != nil {
				return err
			}

			return printCommitIter(commitIter)
		}),
	}
	subscribeCommit.Flags().StringVar(&from, "from", "", "subscribe to all commits since this commit")
	subscribeCommit.Flags().BoolVar(&new, "new", false, "subscribe to only new commits created from now on")
	rawFlag(subscribeCommit)

	deleteCommit := &cobra.Command{
		Use:   "delete-commit repo-name commit-id",
		Short: "Delete an input commit.",
		Long:  "Delete an input commit. An input is a commit which is not the output of a pipeline.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return client.DeleteCommit(args[0], args[1])
		}),
	}

	var branchProvenance cmdutil.RepeatedStringArg
	var head string
	createBranch := &cobra.Command{
		Use:   "create-branch <repo-name> <branch-name> [flags]",
		Short: "Create a new branch, or update an existing branch, on a repo.",
		Long:  "Create a new branch, or update an existing branch, on a repo, starting a commit on the branch will also create it, so there's often no need to call this.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			provenance, err := cmdutil.ParseBranches(branchProvenance)
			if err != nil {
				return err
			}
			return client.CreateBranch(args[0], args[1], head, provenance)
		}),
	}
	createBranch.Flags().VarP(&branchProvenance, "provenance", "p", "The provenance for the branch.")
	createBranch.Flags().StringVarP(&head, "head", "", "", "The head of the newly created branch.")

	listBranch := &cobra.Command{
		Use:   "list-branch repo-name",
		Short: "Return all branches on a repo.",
		Long:  "Return all branches on a repo.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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
	rawFlag(listBranch)

	setBranch := &cobra.Command{
		Use:   "set-branch repo-name commit-id/branch-name new-branch-name",
		Short: "DEPRECATED Set a commit and its ancestors to a branch",
		Long: `DEPRECATED Set a commit and its ancestors to a branch.

Examples:

` + codestart + `# Set commit XXX and its ancestors as branch master in repo foo.
$ pachctl set-branch foo XXX master

# Set the head of branch test as branch master in repo foo.
# After running this command, "test" and "master" both point to the
# same commit.
$ pachctl set-branch foo test master` + codeend,
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			fmt.Fprintf(os.Stderr, "set-branch is DEPRECATED, use create-branch instead.\n")
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return client.SetBranch(args[0], args[1], args[2])
		}),
	}

	deleteBranch := &cobra.Command{
		Use:   "delete-branch repo-name branch-name",
		Short: "Delete a branch",
		Long:  "Delete a branch, while leaving the commits intact",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return client.DeleteBranch(args[0], args[1], force)
		}),
	}
	deleteBranch.Flags().BoolVarP(&force, "force", "f", false, "remove the branch regardless of errors; use with care")

	file := &cobra.Command{
		Use:   "file",
		Short: "Docs for files.",
		Long: `Files are the lowest level data object in Pachyderm.

	Files can be written to started (but not finished) commits with put-file.
	Files can be read from finished commits with get-file.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

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
		Use:   "put-file repo-name branch [path/to/file/in/pfs]",
		Short: "Put a file into the filesystem.",
		Long: `Put-file supports a number of ways to insert data into pfs:
` + codestart + `# Put data from stdin as repo/branch/path:
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
$ pachctl put-file repo branch -i http://host/path
` + codeend + `
NOTE there's a small performance overhead for using a branch name as opposed
to a commit ID in put-file.  In most cases the performance overhead is
negligible, but if you are putting a large number of small files, you might
want to consider using commit IDs directly.
`,
		Run: cmdutil.RunBoundedArgs(2, 3, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine(metrics, "user", client.WithMaxConcurrentStreams(parallelism))
			if err != nil {
				return err
			}
			pfc, err := c.NewPutFileClient()
			if err != nil {
				return err
			}
			defer func() {
				if err := pfc.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			repoName := args[0]
			branch := args[1]
			var path string
			if len(args) == 3 {
				path = args[2]
				if url, err := url.Parse(path); err == nil && url.Scheme != "" {
					fmt.Fprintf(os.Stderr, "warning: PFS destination \"%s\" looks like a URL; did you mean -f %s?\n", path, path)
				}
			}
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
				if len(args) == 2 {
					// The user has not specified a path so we use source as path.
					if source == "-" {
						return fmt.Errorf("must specify filename when reading data from stdin")
					}
					eg.Go(func() error {
						return putFileHelper(c, pfc, repoName, branch, joinPaths("", source), source, recursive, overwrite, limiter, split, targetFileDatums, targetFileBytes, headerRecords, filesPut)
					})
				} else if len(sources) == 1 && len(args) == 3 {
					// We have a single source and the user has specified a path,
					// we use the path and ignore source (in terms of naming the file).
					eg.Go(func() error {
						return putFileHelper(c, pfc, repoName, branch, path, source, recursive, overwrite, limiter, split, targetFileDatums, targetFileBytes, headerRecords, filesPut)
					})
				} else if len(sources) > 1 && len(args) == 3 {
					// We have multiple sources and the user has specified a path,
					// we use that path as a prefix for the filepaths.
					eg.Go(func() error {
						return putFileHelper(c, pfc, repoName, branch, joinPaths(path, source), source, recursive, overwrite, limiter, split, targetFileDatums, targetFileBytes, headerRecords, filesPut)
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
	putFile.Flags().StringVar(&split, "split", "", "Split the input file into smaller files, subject to the constraints of --target-file-datums and --target-file-bytes. Permissible values are `json` and `line`.")
	putFile.Flags().UintVar(&targetFileDatums, "target-file-datums", 0, "The upper bound of the number of datums that each file contains, the last file will contain fewer if the datums don't divide evenly; needs to be used with --split.")
	putFile.Flags().UintVar(&targetFileBytes, "target-file-bytes", 0, "The target upper bound of the number of bytes that each file contains; needs to be used with --split.")
	putFile.Flags().UintVar(&headerRecords, "header-records", 0, "the number of records that will be converted to a PFS 'header', and prepended to future retrievals of any subset of data from PFS; needs to be used with --split=(json|line|csv)")
	putFile.Flags().BoolVarP(&putFileCommit, "commit", "c", false, "DEPRECATED: Put file(s) in a new commit.")
	putFile.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite the existing content of the file, either from previous commits or previous calls to put-file within this commit.")

	copyFile := &cobra.Command{
		Use:   "copy-file src-repo src-commit src-path dst-repo dst-commit dst-path",
		Short: "Copy files between pfs paths.",
		Long:  "Copy files between pfs paths.",
		Run: cmdutil.RunFixedArgs(6, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine(metrics, "user", client.WithMaxConcurrentStreams(parallelism))
			if err != nil {
				return err
			}
			return c.CopyFile(args[0], args[1], args[2], args[3], args[4], args[5], overwrite)
		}),
	}
	copyFile.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite the existing content of the file, either from previous commits or previous calls to put-file within this commit.")

	var outputPath string
	getFile := &cobra.Command{
		Use:   "get-file repo-name commit-id path/to/file",
		Short: "Return the contents of a file.",
		Long: `Return the contents of a file.
` + codestart + `# get file "XXX" on branch "master" in repo "foo"
$ pachctl get-file foo master XXX

# get file "XXX" in the parent of the current head of branch "master"
# in repo "foo"
$ pachctl get-file foo master^ XXX

# get file "XXX" in the grandparent of the current head of branch "master"
# in repo "foo"
$ pachctl get-file foo master^2 XXX
` + codeend,
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if recursive {
				if outputPath == "" {
					return fmt.Errorf("an output path needs to be specified when using the --recursive flag")
				}
				puller := sync.NewPuller()
				return puller.Pull(client, outputPath, args[0], args[1], args[2], false, false, parallelism, nil, "")
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
			return client.GetFile(args[0], args[1], args[2], 0, 0, w)
		}),
	}
	getFile.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively download a directory.")
	getFile.Flags().StringVarP(&outputPath, "output", "o", "", "The path where data will be downloaded.")
	getFile.Flags().IntVarP(&parallelism, "parallelism", "p", DefaultParallelism, "The maximum number of files that can be downloaded in parallel")

	inspectFile := &cobra.Command{
		Use:   "inspect-file repo-name commit-id path/to/file",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			fileInfo, err := client.InspectFile(args[0], args[1], args[2])
			if err != nil {
				return err
			}
			if fileInfo == nil {
				return fmt.Errorf("file %s not found", args[2])
			}
			if raw {
				return marshaller.Marshal(os.Stdout, fileInfo)
			}
			return pretty.PrintDetailedFileInfo(fileInfo)
		}),
	}
	rawFlag(inspectFile)

	listFile := &cobra.Command{
		Use:   "list-file repo-name commit-id path/to/dir",
		Short: "Return the files in a directory.",
		Long: `Return the files in a directory.

Examples:

` + codestart + `# list top-level files on branch "master" in repo "foo"
$ pachctl list-file foo master

# list files under path XXX on branch "master" in repo "foo"
$ pachctl list-file foo master XXX

# list top-level files in the parent commit of the current head of "master"
# in repo "foo"
$ pachctl list-file foo master^

# list top-level files in the grandparent of the current head of "master"
# in repo "foo"
$ pachctl list-file foo master^2
` + codeend,
		Run: cmdutil.RunBoundedArgs(2, 3, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			var path string
			if len(args) == 3 {
				path = args[2]
			}
			if raw {
				return client.ListFileF(args[0], args[1], path, func(fi *pfsclient.FileInfo) error {
					return marshaller.Marshal(os.Stdout, fi)
				})
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
			if err := client.ListFileF(args[0], args[1], path, func(fi *pfsclient.FileInfo) error {
				pretty.PrintFileInfo(writer, fi)
				return nil
			}); err != nil {
				return nil
			}
			return writer.Flush()
		}),
	}
	rawFlag(listFile)

	globFile := &cobra.Command{
		Use:   "glob-file repo-name commit-id pattern",
		Short: "Return files that match a glob pattern in a commit.",
		Long: `Return files that match a glob pattern in a commit (that is, match a glob pattern
in a repo at the state represented by a commit). Glob patterns are
documented [here](https://golang.org/pkg/path/filepath/#Match).

Examples:

` + codestart + `# Return files in repo "foo" on branch "master" that start
# with the character "A".  Note how the double quotation marks around "A*" are
# necessary because otherwise your shell might interpret the "*".
$ pachctl glob-file foo master "A*"

# Return files in repo "foo" on branch "master" under directory "data".
$ pachctl glob-file foo master "data/*"
` + codeend,
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			fileInfos, err := client.GlobFile(args[0], args[1], args[2])
			if err != nil {
				return err
			}
			if raw {
				for _, fileInfo := range fileInfos {
					if err := marshaller.Marshal(os.Stdout, fileInfo); err != nil {
						return err
					}
				}
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo)
			}
			return writer.Flush()
		}),
	}
	rawFlag(globFile)

	var shallow bool
	diffFile := &cobra.Command{
		Use:   "diff-file new-repo-name new-commit-id new-path [old-repo-name old-commit-id old-path]",
		Short: "Return a diff of two file trees.",
		Long: `Return a diff of two file trees.

Examples:

` + codestart + `# Return the diff between foo master path and its parent.
$ pachctl diff-file foo master path

# Return the diff between foo master path1 and bar master path2.
$ pachctl diff-file foo master path1 bar master path2
` + codeend,
		Run: cmdutil.RunBoundedArgs(3, 6, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			var newFiles []*pfsclient.FileInfo
			var oldFiles []*pfsclient.FileInfo
			switch {
			case len(args) == 3:
				newFiles, oldFiles, err = client.DiffFile(args[0], args[1], args[2], "", "", "", shallow)
			case len(args) == 6:
				newFiles, oldFiles, err = client.DiffFile(args[0], args[1], args[2], args[3], args[4], args[5], shallow)
			default:
				return fmt.Errorf("diff-file expects either 3 or 6 args, got %d", len(args))
			}
			if err != nil {
				return err
			}
			if len(newFiles) > 0 {
				fmt.Println("New Files:")
				writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
				for _, fileInfo := range newFiles {
					pretty.PrintFileInfo(writer, fileInfo)
				}
				if err := writer.Flush(); err != nil {
					return err
				}
			}
			if len(oldFiles) > 0 {
				fmt.Println("Old Files:")
				writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeader)
				for _, fileInfo := range oldFiles {
					pretty.PrintFileInfo(writer, fileInfo)
				}
				if err := writer.Flush(); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	diffFile.Flags().BoolVarP(&shallow, "shallow", "s", false, "Specifies whether or not to diff subdirectories")

	deleteFile := &cobra.Command{
		Use:   "delete-file repo-name commit-id path/to/file",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		Run: cmdutil.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return client.DeleteFile(args[0], args[1], args[2])
		}),
	}

	getObject := &cobra.Command{
		Use:   "get-object hash",
		Short: "Return the contents of an object",
		Long:  "Return the contents of an object",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return client.GetObject(args[0], os.Stdout)
		}),
	}

	getTag := &cobra.Command{
		Use:   "get-tag tag",
		Short: "Return the contents of a tag",
		Long:  "Return the contents of a tag",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			return client.GetTag(args[0], os.Stdout)
		}),
	}

	var debug bool
	var commits cmdutil.RepeatedStringArg
	mount := &cobra.Command{
		Use:   "mount path/to/mount/point",
		Short: "Mount pfs locally. This command blocks.",
		Long:  "Mount pfs locally. This command blocks.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(metrics, "fuse")
			if err != nil {
				return err
			}
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
	mount.Flags().VarP(&commits, "commits", "c", "Commits to mount for repos, arguments should be of the form \"repo:commit\"")

	unmount := &cobra.Command{
		Use:   "unmount path/to/mount/point",
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

	var result []*cobra.Command
	result = append(result, repo)
	result = append(result, createRepo)
	result = append(result, updateRepo)
	result = append(result, inspectRepo)
	result = append(result, listRepo)
	result = append(result, deleteRepo)
	result = append(result, commit)
	result = append(result, startCommit)
	result = append(result, finishCommit)
	result = append(result, inspectCommit)
	result = append(result, listCommit)
	result = append(result, flushCommit)
	result = append(result, subscribeCommit)
	result = append(result, deleteCommit)
	result = append(result, createBranch)
	result = append(result, listBranch)
	result = append(result, setBranch)
	result = append(result, deleteBranch)
	result = append(result, file)
	result = append(result, putFile)
	result = append(result, copyFile)
	result = append(result, getFile)
	result = append(result, inspectFile)
	result = append(result, listFile)
	result = append(result, globFile)
	result = append(result, diffFile)
	result = append(result, deleteFile)
	result = append(result, getObject)
	result = append(result, getTag)
	result = append(result, mount)
	result = append(result, unmount)
	return result
}

func parseCommits(args []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, arg := range args {
		split := strings.Split(arg, ":")
		if len(split) != 2 {
			return nil, fmt.Errorf("malformed input %s, must be of the form repo:commit", args)
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
			"delete-file or delete-commit", path)
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
				// don't do a second recursive put-file, just put the one file at
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

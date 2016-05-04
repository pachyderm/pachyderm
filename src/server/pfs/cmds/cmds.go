package cmds

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmd"

	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func Cmds(address string) []*cobra.Command {
	var fileNumber int
	var fileModulus int
	var blockNumber int
	var blockModulus int
	shard := func() *pfsclient.Shard {
		return &pfsclient.Shard{
			FileNumber:   uint64(fileNumber),
			FileModulus:  uint64(fileModulus),
			BlockNumber:  uint64(blockNumber),
			BlockModulus: uint64(blockModulus),
		}
	}

	addShardFlags := func(cmd *cobra.Command) {
		cmd.Flags().IntVarP(&fileNumber, "file-shard", "s", 0, "file shard to read")
		cmd.Flags().IntVarP(&fileModulus, "file-modulus", "m", 1, "modulus of file shard")
		cmd.Flags().IntVarP(&blockNumber, "block-shard", "b", 0, "block shard to read")
		cmd.Flags().IntVarP(&blockModulus, "block-modulus", "n", 1, "modulus of block shard")
	}

	repo := &cobra.Command{
		Use:   "repo",
		Short: "Docs for repos.",
		Long: `Repos, short for repository, are the top level data object in Pachyderm.

Repos are created with create-repo.`,
		Run: cmd.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	createRepo := &cobra.Command{
		Use:   "create-repo repo-name",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		Run: cmd.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			return client.CreateRepo(args[0])
		}),
	}

	inspectRepo := &cobra.Command{
		Use:   "inspect-repo repo-name",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		Run: cmd.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			repoInfo, err := client.InspectRepo(args[0])
			if err != nil {
				return err
			}
			if repoInfo == nil {
				return fmt.Errorf("repo %s not found", args[0])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintRepoHeader(writer)
			pretty.PrintRepoInfo(writer, repoInfo)
			return writer.Flush()
		}),
	}

	listRepo := &cobra.Command{
		Use:   "list-repo",
		Short: "Return all repos.",
		Long:  "Reutrn all repos.",
		Run: cmd.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			repoInfos, err := client.ListRepo()
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintRepoHeader(writer)
			for _, repoInfo := range repoInfos {
				pretty.PrintRepoInfo(writer, repoInfo)
			}
			return writer.Flush()
		}),
	}

	deleteRepo := &cobra.Command{
		Use:   "delete-repo repo-name",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		Run: cmd.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			return client.DeleteRepo(args[0])
		}),
	}

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
This layers the data in the commit over the data in the parent.`,
		Run: cmd.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	var parentCommitID string
	startCommit := &cobra.Command{
		Use:   "start-commit repo-name [branch]",
		Short: "Start a new commit.",
		Long:  "Start a new commit with parent-commit-id as the parent.",
		Run: cmd.RunBoundedArgs(1, 2, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			branch := ""
			if len(args) == 2 {
				branch = args[1]
			}
			commit, err := client.StartCommit(args[0],
				parentCommitID, branch)
			if err != nil {
				return err
			}
			fmt.Println(commit.ID)
			return nil
		}),
	}
	startCommit.Flags().StringVarP(&parentCommitID, "parent", "p", "", "parent id")

	var cancel bool
	finishCommit := &cobra.Command{
		Use:   "finish-commit repo-name commit-id",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		Run: cmd.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			if cancel {
				return client.CancelCommit(args[0], args[1])
			}
			return client.FinishCommit(args[0], args[1])
		}),
	}
	finishCommit.Flags().BoolVarP(&cancel, "cancel", "c", false, "cancel the commit")

	inspectCommit := &cobra.Command{
		Use:   "inspect-commit repo-name commit-id",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		Run: cmd.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewFromAddress(address)
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
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintCommitInfoHeader(writer)
			pretty.PrintCommitInfo(writer, commitInfo)
			return writer.Flush()
		}),
	}

	var all bool
	var block bool
	listCommit := &cobra.Command{
		Use:   "list-commit repo-name",
		Short: "Return all commits on a set of repos",
		Long: `Return all commits on a set of repos.

Examples:

	# return commits in repo "foo" and repo "bar"
	$ pachctl list-commit foo bar

	# return commits in repo "foo" since commit abc123 and those in repo "bar" since commit def456
	$ pachctl list-commit foo/abc123 bar/def456

`,
		Run: pkgcobra.Run(func(args []string) error {
			commits, err := cmd.ParseCommits(args)
			if err != nil {
				return err
			}

			var repos []string
			var fromCommits []string
			for _, commit := range commits {
				repos = append(repos, commit.Repo.Name)
				fromCommits = append(fromCommits, commit.ID)
			}

			_client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}

			commitInfos, err := _client.ListCommit(repos, fromCommits, client.CommitTypeNone, block, all)
			if err != nil {
				return err
			}

			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintCommitInfoHeader(writer)
			for _, commitInfo := range commitInfos {
				pretty.PrintCommitInfo(writer, commitInfo)
			}
			return writer.Flush()
		}),
	}
	listCommit.Flags().BoolVarP(&all, "all", "a", false, "list all commits including cancelled commits")
	listCommit.Flags().BoolVarP(&block, "block", "b", false, "block until there are new commits since the `from` commits")

	listBranch := &cobra.Command{
		Use:   "list-branch repo-name",
		Short: "Return all branches on a repo.",
		Long:  "Return all branches on a repo.",
		Run: cmd.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			commitInfos, err := client.ListBranch(args[0])
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintCommitInfoHeader(writer)
			for _, commitInfo := range commitInfos {
				pretty.PrintCommitInfo(writer, commitInfo)
			}
			return writer.Flush()
		}),
	}

	file := &cobra.Command{
		Use:   "file",
		Short: "Docs for files.",
		Long: `Files are the lowest level data object in Pachyderm.

Files can be written to started (but not finished) commits with put-file.
Files can be read from finished commits with get-file.`,
		Run: cmd.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	putFile := &cobra.Command{
		Use:   "put-file repo-name commit-id path/to/file",
		Short: "Put a file from stdin",
		Long:  "Put a file from stdin. commit-id must be a writeable commit.",
		Run: cmd.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			_, err = client.PutFile(args[0], args[1], args[2], os.Stdin)
			return err
		}),
	}

	var fromCommitID string
	var unsafe bool
	getFile := &cobra.Command{
		Use:   "get-file repo-name commit-id path/to/file",
		Short: "Return the contents of a file.",
		Long:  "Return the contents of a file.",
		Run: cmd.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			return client.GetFile(args[0], args[1], args[2], 0, 0, fromCommitID, shard(), os.Stdout)
		}),
	}
	addShardFlags(getFile)
	getFile.Flags().StringVarP(&fromCommitID, "from", "f", "", "from commit")
	getFile.Flags().BoolVar(&unsafe, "unsafe", false, "use this flag if you need to read data written in the current commit; this operation will race with concurrent writes")

	inspectFile := &cobra.Command{
		Use:   "inspect-file repo-name commit-id path/to/file",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		Run: cmd.RunFixedArgs(3, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			fileInfo, err := client.InspectFile(args[0], args[1], args[2], "", shard())
			if err != nil {
				return err
			}
			if fileInfo == nil {
				return fmt.Errorf("file %s not found", args[2])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintFileInfoHeader(writer)
			pretty.PrintFileInfo(writer, fileInfo)
			return writer.Flush()
		}),
	}
	addShardFlags(inspectFile)
	inspectFile.Flags().BoolVar(&unsafe, "unsafe", false, "use this flag if you need to inspect files written in the current commit; this operation will race with concurrent writes")

	listFile := &cobra.Command{
		Use:   "list-file repo-name commit-id path/to/dir",
		Short: "Return the files in a directory.",
		Long:  "Return the files in a directory.",
		Run: cmd.RunBoundedArgs(2, 3, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			var path string
			if len(args) == 3 {
				path = args[2]
			}
			fileInfos, err := client.ListFile(args[0], args[1], path, "", shard(), true)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintFileInfoHeader(writer)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo)
			}
			return writer.Flush()
		}),
	}
	addShardFlags(listFile)
	listFile.Flags().BoolVar(&unsafe, "unsafe", false, "use this flag if you need to list files written in the current commit; this operation will race with concurrent writes")

	deleteFile := &cobra.Command{
		Use:   "delete-file repo-name commit-id path/to/file",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		Run: cmd.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			return client.DeleteFile(args[0], args[1], args[2])
		}),
	}

	mount := &cobra.Command{
		Use:   "mount path/to/mount/point",
		Short: "Mount pfs locally.",
		Long:  "Mount pfs locally.",
		Run: func(c *cobra.Command, args []string) {
			//lion.SetLevel(lion.LevelDebug)
			client, err := client.NewFromAddress(address)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			mounter := fuse.NewMounter(address, client.PfsAPIClient)
			if len(args) != 1 {
				c.Usage()
				os.Exit(1)
			}
			mountPoint := args[0]
			err = mounter.Mount(mountPoint, shard(), nil, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}
	addShardFlags(mount)

	var result []*cobra.Command
	result = append(result, repo)
	result = append(result, createRepo)
	result = append(result, inspectRepo)
	result = append(result, listRepo)
	result = append(result, deleteRepo)
	result = append(result, commit)
	result = append(result, startCommit)
	result = append(result, finishCommit)
	result = append(result, inspectCommit)
	result = append(result, listCommit)
	result = append(result, listBranch)
	result = append(result, file)
	result = append(result, putFile)
	result = append(result, getFile)
	result = append(result, inspectFile)
	result = append(result, listFile)
	result = append(result, deleteFile)
	result = append(result, mount)
	return result
}

func parseCommitMounts(args []string) []*fuse.CommitMount {
	var result []*fuse.CommitMount
	for _, arg := range args {
		commitMount := &fuse.CommitMount{Commit: client.NewCommit("", "")}
		repo, commitAlias := path.Split(arg)
		commitMount.Commit.Repo.Name = path.Clean(repo)
		split := strings.Split(commitAlias, ":")
		if len(split) > 0 {
			commitMount.Commit.ID = split[0]
		}
		if len(split) > 1 {
			commitMount.Alias = split[1]
		}
		result = append(result, commitMount)
	}
	return result
}

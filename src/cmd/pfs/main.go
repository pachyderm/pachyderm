package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"go.pedge.io/env"
	"go.pedge.io/proto/client"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/pretty"
	"github.com/pachyderm/pachyderm/src/pkg/cobramainutil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	defaultEnv = map[string]string{
		"PFS_ADDRESS": "0.0.0.0:650",
	}
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	Address            string `env:"PFS_ADDRESS"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	address := appEnv.PachydermPfsd1Port
	if address == "" {
		address = appEnv.Address
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	apiClient := pfs.NewApiClient(clientConn)

	var shard int
	var modulus int

	initCmd := cobramainutil.Command{
		Use:     "init repo-name",
		Long:    "Initalize a repo.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.RepoCreate(apiClient, args[0])
		},
	}.ToCobraCommand()

	listReposCmd := cobramainutil.Command{
		Use:     "list-repos",
		Long:    "List repositories.",
		NumArgs: 0,
		Run: func(cmd *cobra.Command, args []string) error {
			repoInfos, err := pfsutil.RepoList(apiClient)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintRepoHeader(writer)
			for _, repoInfo := range repoInfos {
				pretty.PrintRepoInfo(writer, repoInfo)
			}
			return writer.Flush()
		},
	}.ToCobraCommand()

	mkdirCmd := cobramainutil.Command{
		Use:     "mkdir repo-name commit-id path/to/dir",
		Long:    "Make a directory. Sub directories must already exist.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.MakeDirectory(apiClient, args[0], args[1], args[2])
		},
	}.ToCobraCommand()

	putCmd := cobramainutil.Command{
		Use:     "put repo-name branch-id path/to/file",
		Long:    "Put a file from stdin. Directories must exist. branch-id must be a writeable commit.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			_, err := pfsutil.FilePut(apiClient, args[0], args[1], args[2], 0, os.Stdin)
			return err
		},
	}.ToCobraCommand()

	getCmd := cobramainutil.Command{
		Use:     "get repo-name commit-id path/to/file",
		Long:    "Get a file from stdout. commit-id must be a readable commit.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.FileGet(apiClient, args[0], args[1], args[2], 0, pfsutil.GetAll, os.Stdout)
		},
	}.ToCobraCommand()

	lsCmd := cobramainutil.Command{
		Use:     "ls repo-name branch-id path/to/dir",
		Long:    "List a directory. Directory must exist.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			fileInfos, err := pfsutil.FileList(apiClient, args[0], args[1], args[2], uint64(shard), uint64(modulus))
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintFileInfoHeader(writer)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo)
			}
			return writer.Flush()
		},
	}.ToCobraCommand()
	lsCmd.Flags().IntVarP(&shard, "shard", "s", 0, "shard to read from")
	lsCmd.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	branchCmd := cobramainutil.Command{
		Use:     "branch repo-name commit-id",
		Long:    "Branch a commit. commit-id must be a readable commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			commit, err := pfsutil.CommitStart(apiClient, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Println(commit.Id)
			return nil
		},
	}.ToCobraCommand()

	commitCmd := cobramainutil.Command{
		Use:     "commit repo-name branch-id",
		Long:    "Commit a branch. branch-id must be a writeable commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.CommitFinish(apiClient, args[0], args[1])
		},
	}.ToCobraCommand()

	commitInfoCmd := cobramainutil.Command{
		Use:     "commit-info repo-name commit-id",
		Long:    "Get info for a commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			commitInfo, err := pfsutil.CommitInspect(apiClient, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", commitInfo)
			return nil
		},
	}.ToCobraCommand()

	listCommitsCmd := cobramainutil.Command{
		Use:     "list-commits repo-name",
		Long:    "List commits on the repo.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			commitInfos, err := pfsutil.CommitList(apiClient, args[0])
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintCommitInfoHeader(writer)
			for _, commitInfo := range commitInfos {
				pretty.PrintCommitInfo(writer, commitInfo)
			}
			return writer.Flush()
		},
	}.ToCobraCommand()

	mountCmd := cobramainutil.Command{
		Use:        "mount mountpoint repo-name [commit-id]",
		Long:       "Mount a repo as a local file system.",
		MinNumArgs: 2,
		MaxNumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			mountPoint := args[0]
			repo := args[1]
			commitID := ""
			if len(args) == 3 {
				commitID = args[2]
			}
			mounter := fuse.NewMounter(apiClient)
			if err := mounter.Mount(repo, commitID, mountPoint, uint64(shard), uint64(modulus)); err != nil {
				return err
			}
			return mounter.Wait(mountPoint)
		},
	}.ToCobraCommand()
	mountCmd.Flags().IntVarP(&shard, "shard", "s", 0, "shard to read from")
	mountCmd.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	rootCmd := &cobra.Command{
		Use: "pfs",
		Long: `Access the PFS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PFS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:650.`,
	}

	rootCmd.AddCommand(protoclient.NewVersionCommand(clientConn, pachyderm.Version, nil))
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listReposCmd)
	rootCmd.AddCommand(mkdirCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(commitInfoCmd)
	rootCmd.AddCommand(listCommitsCmd)
	rootCmd.AddCommand(mountCmd)
	return rootCmd.Execute()
}

package main

import (
	"fmt"
	"os"
	"strings"

	"go.pedge.io/proto/client"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/cobramainutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
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
	mainutil.Main(do, &appEnv{}, defaultEnv)
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
		Use:     "init repository-name",
		Long:    "Initalize a repository.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.InitRepository(apiClient, args[0])
		},
	}.ToCobraCommand()

	mkdirCmd := cobramainutil.Command{
		Use:     "mkdir repository-name commit-id path/to/dir",
		Long:    "Make a directory. Sub directories must already exist.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.MakeDirectory(apiClient, args[0], args[1], args[2])
		},
	}.ToCobraCommand()

	putCmd := cobramainutil.Command{
		Use:     "put repository-name branch-id path/to/file",
		Long:    "Put a file from stdin. Directories must exist. branch-id must be a writeable commit.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			_, err := pfsutil.PutFile(apiClient, args[0], args[1], args[2], 0, os.Stdin)
			return err
		},
	}.ToCobraCommand()

	getCmd := cobramainutil.Command{
		Use:     "get repository-name commit-id path/to/file",
		Long:    "Get a file from stdout. commit-id must be a readable commit.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.GetFile(apiClient, args[0], args[1], args[2], 0, pfsutil.GetAll, os.Stdout)
		},
	}.ToCobraCommand()

	lsCmd := cobramainutil.Command{
		Use:     "ls repository-name branch-id path/to/dir",
		Long:    "List a directory. Directory must exist.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			listFilesResponse, err := pfsutil.ListFiles(apiClient, args[0], args[1], args[2], uint64(shard), uint64(modulus))
			if err != nil {
				return err
			}
			for _, fileInfo := range listFilesResponse.FileInfo {
				fmt.Printf("%+v\n", fileInfo)
			}
			return nil
		},
	}.ToCobraCommand()
	lsCmd.Flags().IntVarP(&shard, "shard", "s", 0, "shard to read from")
	lsCmd.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	branchCmd := cobramainutil.Command{
		Use:     "branch repository-name commit-id",
		Long:    "Branch a commit. commit-id must be a readable commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			branchResponse, err := pfsutil.Branch(apiClient, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Println(branchResponse.Commit.Id)
			return nil
		},
	}.ToCobraCommand()

	commitCmd := cobramainutil.Command{
		Use:     "commit repository-name branch-id",
		Long:    "Commit a branch. branch-id must be a writeable commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.Commit(apiClient, args[0], args[1])
		},
	}.ToCobraCommand()

	commitInfoCmd := cobramainutil.Command{
		Use:     "commit-info repository-name commit-id",
		Long:    "Get info for a commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			commitInfoResponse, err := pfsutil.GetCommitInfo(apiClient, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", commitInfoResponse.CommitInfo)
			return nil
		},
	}.ToCobraCommand()

	listCommitsCmd := cobramainutil.Command{
		Use:     "list-commits repository-name",
		Long:    "List commits on the repository.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			listCommitsResponse, err := pfsutil.ListCommits(apiClient, args[0])
			if err != nil {
				return err
			}
			for _, commitInfo := range listCommitsResponse.CommitInfo {
				fmt.Printf("%+v\n", commitInfo)
			}
			return nil
		},
	}.ToCobraCommand()

	mountCmd := cobramainutil.Command{
		Use:        "mount mountpoint repository-name [commit-id]",
		Long:       "Mount a repository as a local file system.",
		MinNumArgs: 2,
		MaxNumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			mountPoint := args[0]
			repository := args[1]
			commitID := ""
			if len(args) == 3 {
				commitID = args[2]
			}
			mounter := fuse.NewMounter(apiClient)
			if err := mounter.Mount(repository, commitID, mountPoint, uint64(shard), uint64(modulus)); err != nil {
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

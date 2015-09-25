package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"go.pedge.io/env"
	"go.pedge.io/proto/client"
	"go.pedge.io/protolog/logrus"

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
	logrus.Register()
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

	createRepo := cobramainutil.Command{
		Use:     "create-repo repo-name",
		Short:   "Create a new repo.",
		Long:    "Create a new repo.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.CreateRepo(apiClient, args[0])
		},
	}.ToCobraCommand()

	inspectRepo := cobramainutil.Command{
		Use:     "inspect-repo repo-name",
		Short:   "Return info about a repo.",
		Long:    "Return info about a repo.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			repoInfo, err := pfsutil.InspectRepo(apiClient, args[0])
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
		},
	}.ToCobraCommand()

	listRepo := cobramainutil.Command{
		Use:     "list-repo",
		Short:   "Return all repos.",
		Long:    "Reutrn all repos.",
		NumArgs: 0,
		Run: func(cmd *cobra.Command, args []string) error {
			repoInfos, err := pfsutil.ListRepo(apiClient)
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

	deleteRepo := cobramainutil.Command{
		Use:     "delete-repo repo-name",
		Short:   "Delete a repo.",
		Long:    "Delete a repo.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.DeleteRepo(apiClient, args[0])
		},
	}.ToCobraCommand()

	startCommit := cobramainutil.Command{
		Use:     "start-commit repo-name parent-commit-id",
		Short:   "Start a new commit.",
		Long:    "Start a new commit with parent-commit-id as the parent.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			commit, err := pfsutil.StartCommit(apiClient, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Println(commit.Id)
			return nil
		},
	}.ToCobraCommand()

	finishCommit := cobramainutil.Command{
		Use:     "finish-commit repo-name commit-id",
		Short:   "Finish a started commit.",
		Long:    "Finish a started commit. Commit-id must be a writeable commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.FinishCommit(apiClient, args[0], args[1])
		},
	}.ToCobraCommand()

	inspectCommit := cobramainutil.Command{
		Use:     "inspect-commit repo-name commit-id",
		Short:   "Return info about a commit.",
		Long:    "Return info about a commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			commitInfo, err := pfsutil.InspectCommit(apiClient, args[0], args[1])
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
		},
	}.ToCobraCommand()

	listCommit := cobramainutil.Command{
		Use:     "list-commit repo-name",
		Short:   "Return all commits on a repo.",
		Long:    "Return all commits on a repo.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			commitInfos, err := pfsutil.ListCommit(apiClient, args[0])
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

	deleteCommit := cobramainutil.Command{
		Use:     "delete-commit repo-name commit-id",
		Short:   "Delete a commit.",
		Long:    "Delete a commit.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.DeleteCommit(apiClient, args[0], args[1])
		},
	}.ToCobraCommand()

	mkdir := cobramainutil.Command{
		Use:     "mkdir repo-name commit-id path/to/dir",
		Short:   "Make a directory.",
		Long:    "Make a directory. Parent directories need not exist.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.MakeDirectory(apiClient, args[0], args[1], args[2])
		},
	}.ToCobraCommand()

	putFile := cobramainutil.Command{
		Use:     "put-file repo-name commit-id path/to/file",
		Short:   "Put a file from stdin",
		Long:    "Put a file from stdin. Directories must exist. Commit -id must be a writeable commit.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			_, err := pfsutil.PutFile(apiClient, args[0], args[1], args[2], 0, os.Stdin)
			return err
		},
	}.ToCobraCommand()

	getFile := cobramainutil.Command{
		Use:     "get-file repo-name commit-id path/to/file",
		Short:   "Return the contents of a file.",
		Long:    "Return the contents of a file.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.GetFile(apiClient, args[0], args[1], args[2], 0, pfsutil.GetAll, os.Stdout)
		},
	}.ToCobraCommand()

	inspectFile := cobramainutil.Command{
		Use:     "inspect-file repo-name branch-id path/to/file",
		Short:   "Return info about a file.",
		Long:    "Return info about a file.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			fileInfo, err := pfsutil.InspectFile(apiClient, args[0], args[1], args[2])
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
		},
	}.ToCobraCommand()

	listFile := cobramainutil.Command{
		Use:     "list-file repo-name commit-id path/to/dir",
		Short:   "Return the files in a directory.",
		Long:    "Return the files in a directory.",
		NumArgs: 3,
		Run: func(cmd *cobra.Command, args []string) error {
			fileInfos, err := pfsutil.ListFile(apiClient, args[0], args[1], args[2], uint64(shard), uint64(modulus))
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
	listFile.Flags().IntVarP(&shard, "shard", "s", 0, "shard to read from")
	listFile.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	deleteFile := cobramainutil.Command{
		Use:     "delete-file repo-name commit-id path/to/file",
		Short:   "Delete a file.",
		Long:    "Delete a file.",
		NumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			return pfsutil.DeleteFile(apiClient, args[0], args[1], args[2])
		},
	}.ToCobraCommand()

	inspectServer := cobramainutil.Command{
		Use:     "inspect-server server-id",
		Short:   "Inspect a server.",
		Long:    "Inspect a server.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			serverInfo, err := pfsutil.InspectServer(apiClient, args[0])
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintServerInfoHeader(writer)
			pretty.PrintServerInfo(writer, serverInfo)
			return writer.Flush()
		},
	}.ToCobraCommand()

	listServer := cobramainutil.Command{
		Use:     "list-server",
		Short:   "Return all servers in the cluster.",
		Long:    "Return all servers in the cluster.",
		NumArgs: 0,
		Run: func(cmd *cobra.Command, args []string) error {
			serverInfos, err := pfsutil.ListServer(apiClient)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintServerInfoHeader(writer)
			for _, serverInfo := range serverInfos {
				pretty.PrintServerInfo(writer, serverInfo)
			}
			return writer.Flush()
		},
	}.ToCobraCommand()

	mount := cobramainutil.Command{
		Use:        "mount mountpoint repo-name [commit-id]",
		Short:      "Mount a repo as a local file system.",
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
	mount.Flags().IntVarP(&shard, "shard", "s", 0, "shard to read from")
	mount.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	rootCmd := &cobra.Command{
		Use: "pfs",
		Long: `Access the PFS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PFS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:650.`,
	}

	rootCmd.AddCommand(protoclient.NewVersionCommand(clientConn, pachyderm.Version, nil))
	rootCmd.AddCommand(createRepo)
	rootCmd.AddCommand(inspectRepo)
	rootCmd.AddCommand(listRepo)
	rootCmd.AddCommand(deleteRepo)
	rootCmd.AddCommand(startCommit)
	rootCmd.AddCommand(finishCommit)
	rootCmd.AddCommand(inspectCommit)
	rootCmd.AddCommand(listCommit)
	rootCmd.AddCommand(deleteCommit)
	rootCmd.AddCommand(mkdir)
	rootCmd.AddCommand(putFile)
	rootCmd.AddCommand(getFile)
	rootCmd.AddCommand(inspectFile)
	rootCmd.AddCommand(listFile)
	rootCmd.AddCommand(deleteFile)
	rootCmd.AddCommand(inspectServer)
	rootCmd.AddCommand(listServer)
	rootCmd.AddCommand(mount)
	return rootCmd.Execute()
}

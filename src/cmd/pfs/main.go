package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/peter-edge/go-env"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	defaultAddress = "0.0.0.0:650"
)

type appEnv struct {
	Address string `env:"PFS_ADDRESS"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	appEnv := &appEnv{}
	check(env.Populate(appEnv, env.PopulateOptions{}))
	if appEnv.Address == "" {
		appEnv.Address = defaultAddress
	}

	clientConn, err := grpc.Dial(appEnv.Address)
	check(err)
	apiClient := pfs.NewApiClient(clientConn)

	versionCmd := &cobra.Command{
		Use:  "version",
		Long: "Print the version.",
		Run: func(cmd *cobra.Command, args []string) {
			getVersionResponse, err := pfsutil.GetVersion(apiClient)
			check(err)
			fmt.Printf("Client: %s\nServer: %s\n", common.VersionString(), pfs.VersionString(getVersionResponse.Version))
		},
	}

	initCmd := &cobra.Command{
		Use:  "init repository-name",
		Long: "Initalize a repository.",
		Run: func(cmd *cobra.Command, args []string) {
			check(pfsutil.InitRepository(apiClient, args[0]))
		},
	}

	mkdirCmd := &cobra.Command{
		Use:  "mkdir repository-name commit-id path/to/dir",
		Long: "Make a directory. Sub directories must already exist.",
		Run: func(cmd *cobra.Command, args []string) {
			check(pfsutil.MakeDirectory(apiClient, args[0], args[1], args[2]))
		},
	}

	putCmd := &cobra.Command{
		Use:  "put repository-name branch-id path/to/file",
		Long: "Put a file from stdin. Directories must exist. branch-id must be a writeable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := pfsutil.PutFile(apiClient, args[0], args[1], args[2], os.Stdin)
			check(err)
		},
	}

	getCmd := &cobra.Command{
		Use:  "get repository-name commit-id path/to/file",
		Long: "Get a file from stdout. commit-id must be a readable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			reader, err := pfsutil.GetFile(apiClient, args[0], args[1], args[2])
			check(err)
			_, err = bufio.NewReader(reader).WriteTo(os.Stdout)
			check(err)
		},
	}

	lsCmd := &cobra.Command{
		Use:  "ls repository-name branch-id path/to/dir",
		Long: "List a directory. Directory must exist.",
		Run: func(cmd *cobra.Command, args []string) {
			listFilesResponse, err := pfsutil.ListFiles(apiClient, args[0], args[1], args[2], 0, 1)
			check(err)
			for _, fileInfo := range listFilesResponse.FileInfo {
				fmt.Printf("%+v\n", fileInfo)
			}
		},
	}

	branchCmd := &cobra.Command{
		Use:  "branch repository-name commit-id",
		Long: "Branch a commit. commit-id must be a readable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			branchResponse, err := pfsutil.Branch(apiClient, args[0], args[1])
			check(err)
			fmt.Println(branchResponse.Commit.Id)
		},
	}

	commitCmd := &cobra.Command{
		Use:  "commit repository-name branch-id",
		Long: "Commit a branch. branch-id must be a writeable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			check(pfsutil.Commit(apiClient, args[0], args[1]))
		},
	}

	commitInfoCmd := &cobra.Command{
		Use:  "commit-info repository-name commit-id",
		Long: "Get info for a commit.",
		Run: func(cmd *cobra.Command, args []string) {
			commitInfoResponse, err := pfsutil.GetCommitInfo(apiClient, args[0], args[1])
			check(err)
			fmt.Printf("%+v\n", commitInfoResponse.CommitInfo)
		},
	}

	mountCmd := &cobra.Command{
		Use:  "mount repository-name commit-id mount-point",
		Long: "Mount a repository as a local file system.",
		Run: func(cmd *cobra.Command, args []string) {
			check(fuse.Mount(apiClient, args[0], args[1], args[2]))
		},
	}

	rootCmd := &cobra.Command{
		Use: "pfs",
		Long: `Access the PFS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PFS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:650.`,
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mkdirCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(commitInfoCmd)
	rootCmd.AddCommand(mountCmd)
	check(rootCmd.Execute())

	os.Exit(0)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

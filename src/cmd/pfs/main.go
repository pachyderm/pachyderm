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

	initUsage := "init repository-name"
	initCmd := &cobra.Command{
		Use:  initUsage,
		Long: "Initalize a repository.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 1, initUsage)
			check(pfsutil.InitRepository(apiClient, args[0]))
		},
	}
	mkdirUsage := "mkdir repository-name commit-id path/to/dir"
	mkdirCmd := &cobra.Command{
		Use:  mkdirUsage,
		Long: "Make a directory. Sub directories must already exist.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 3, mkdirUsage)
			check(pfsutil.MakeDirectory(apiClient, args[0], args[1], args[2]))
		},
	}

	putUsage := "put repository-name branch-id path/to/file"
	putCmd := &cobra.Command{
		Use:  putUsage,
		Long: "Put a file from stdin. Directories must exist. branch-id must be a writeable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 3, putUsage)
			_, err := pfsutil.PutFile(apiClient, args[0], args[1], args[2], os.Stdin)
			check(err)
		},
	}

	getUsage := "get repository-name commit-id path/to/file"
	getCmd := &cobra.Command{
		Use:  getUsage,
		Long: "Get a file from stdout. commit-id must be a readable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 3, getUsage)
			reader, err := pfsutil.GetFile(apiClient, args[0], args[1], args[2])
			check(err)
			_, err = bufio.NewReader(reader).WriteTo(os.Stdout)
			check(err)
		},
	}

	lsUsage := "ls repository-name branch-id path/to/dir"
	lsCmd := &cobra.Command{
		Use:  lsUsage,
		Long: "List a directory. Directory must exist.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 3, lsUsage)
			listFilesResponse, err := pfsutil.ListFiles(apiClient, args[0], args[1], args[2], 0, 1)
			check(err)
			for _, fileInfo := range listFilesResponse.FileInfo {
				fmt.Printf("%+v\n", fileInfo)
			}
		},
	}

	branchUsage := "branch repository-name commit-id"
	branchCmd := &cobra.Command{
		Use:  branchUsage,
		Long: "Branch a commit. commit-id must be a readable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 2, branchUsage)
			branchResponse, err := pfsutil.Branch(apiClient, args[0], args[1])
			check(err)
			fmt.Println(branchResponse.Commit.Id)
		},
	}

	commitUsage := "commit repository-name branch-id"
	commitCmd := &cobra.Command{
		Use:  commitUsage,
		Long: "Commit a branch. branch-id must be a writeable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 2, commitUsage)
			check(pfsutil.Commit(apiClient, args[0], args[1]))
		},
	}

	commitInfoUsage := "commit-info repository-name commit-id"
	commitInfoCmd := &cobra.Command{
		Use:  commitInfoUsage,
		Long: "Get info for a commit.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 2, commitInfoUsage)
			commitInfoResponse, err := pfsutil.GetCommitInfo(apiClient, args[0], args[1])
			check(err)
			fmt.Printf("%+v\n", commitInfoResponse.CommitInfo)
		},
	}

	mountUsage := "mount repository-name commit-id mount-point"
	mountCmd := &cobra.Command{
		Use:  mountUsage,
		Long: "Mount a repository as a local file system.",
		Run: func(cmd *cobra.Command, args []string) {
			checkArgs(args, 3, mountUsage)
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

func checkArgs(args []string, expected int, usage string) {
	if len(args) != expected {
		fmt.Fprintf(os.Stderr,
			`Error, wrong number of arguments. Got %d, need %d.
%s
`,
			len(args), expected, usage)
		os.Exit(1)
	}
}

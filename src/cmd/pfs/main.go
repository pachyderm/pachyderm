package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/peter-edge/go-env"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type appEnv struct {
	Host string `env:"PFS_HOST,required"`
	Port int    `env:"PFS_PORT,required"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	appEnv := &appEnv{}
	check(env.Populate(appEnv, env.PopulateOptions{}))

	clientConn, err := grpc.Dial(fmt.Sprintf("%s:%d", appEnv.Host, appEnv.Port))
	check(err)
	apiClient := pfs.NewApiClient(clientConn)

	initCmd := &cobra.Command{
		Use:  "init repository-name",
		Long: "Initalize a repository.",
		Run: func(cmd *cobra.Command, args []string) {
			check(initRepository(apiClient, args[0]))
		},
	}

	mkdirCmd := &cobra.Command{
		Use:  "mkdir repository-name commit-id path/to/dir",
		Long: "Make a directory. Sub directories must already exist.",
		Run: func(cmd *cobra.Command, args []string) {
			check(makeDirectory(apiClient, args[0], args[1], args[2]))
		},
	}

	putCmd := &cobra.Command{
		Use:  "put repository-name branch-id path/to/file",
		Long: "Put a file from stdin. Directories must exist. branch-id must be a writeable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			check(putFile(apiClient, args[0], args[1], args[2], os.Stdin))
		},
	}

	getCmd := &cobra.Command{
		Use:  "get repository-name commit-id path/to/file",
		Long: "Get a file from stdout. commit-id must be a readable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			reader, err := getFile(apiClient, args[0], args[1], args[2])
			check(err)
			_, err = bufio.NewReader(reader).WriteTo(os.Stdout)
			check(err)
		},
	}

	lsCmd := &cobra.Command{
		Use:  "ls repository-name branch-id path/to/dir",
		Long: "List a directory. Directory must exist.",
		Run: func(cmd *cobra.Command, args []string) {
			listFilesResponse, err := listFiles(apiClient, args[0], args[1], args[2], 0, 1)
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
			branchResponse, err := branch(apiClient, args[0], args[1])
			check(err)
			fmt.Println(branchResponse.Commit.Id)
		},
	}

	commitCmd := &cobra.Command{
		Use:  "commit repository-name branch-id",
		Long: "Commit a branch. branch-id must be a writeable commit.",
		Run: func(cmd *cobra.Command, args []string) {
			check(commit(apiClient, args[0], args[1]))
		},
	}

	commitInfoCmd := &cobra.Command{
		Use:  "commit-info repository-name commit-id",
		Long: "Get info for a commit.",
		Run: func(cmd *cobra.Command, args []string) {
			commitInfoResponse, err := getCommitInfo(apiClient, args[0], args[1])
			check(err)
			fmt.Printf("%+v\n", commitInfoResponse.CommitInfo)
		},
	}

	rootCmd := &cobra.Command{
		Use: "pfs",
	}
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mkdirCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(commitInfoCmd)
	check(rootCmd.Execute())

	os.Exit(0)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func initRepository(apiClient pfs.ApiClient, repositoryName string) error {
	_, err := apiClient.InitRepository(
		context.Background(),
		&pfs.InitRepositoryRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
	return err
}

func branch(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.BranchResponse, error) {
	return apiClient.Branch(
		context.Background(),
		&pfs.BranchRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
}

func makeDirectory(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) error {
	_, err := apiClient.MakeDirectory(
		context.Background(),
		&pfs.MakeDirectoryRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
	return err
}

func putFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, reader io.Reader) error {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = apiClient.PutFile(
		context.Background(),
		&pfs.PutFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			Value: value,
		},
	)
	return err
}

func getFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) (io.Reader, error) {
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	if err := protoutil.WriteFromStreamingBytesClient(apiGetFileClient, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

func listFiles(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, shardNum int, shardModulo int) (*pfs.ListFilesResponse, error) {
	return apiClient.ListFiles(
		context.Background(),
		&pfs.ListFilesRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			Shard: &pfs.Shard{
				Number: uint64(shardNum),
				Modulo: uint64(shardModulo),
			},
		},
	)
}

func commit(apiClient pfs.ApiClient, repositoryName string, commitID string) error {
	_, err := apiClient.Commit(
		context.Background(),
		&pfs.CommitRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
	return err
}

func getCommitInfo(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.GetCommitInfoResponse, error) {
	return apiClient.GetCommitInfo(
		context.Background(),
		&pfs.GetCommitInfoRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
}

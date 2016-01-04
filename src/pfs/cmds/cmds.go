package cmds

import (
	"fmt"
	"io"
	"math"
	"os"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/pretty"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"google.golang.org/grpc"
)

func Cmds(address string) ([]*cobra.Command, error) {
	var number int
	var modulus int
	shard := func() *pfs.Shard {
		return &pfs.Shard{Number: uint64(number), Modulus: uint64(modulus)}
	}

	createRepo := &cobra.Command{
		Use:   "create-repo repo-name",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.CreateRepo(apiClient, args[0])
		}),
	}

	inspectRepo := &cobra.Command{
		Use:   "inspect-repo repo-name",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
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
		}),
	}

	listRepo := &cobra.Command{
		Use:   "list-repo",
		Short: "Return all repos.",
		Long:  "Reutrn all repos.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
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
		}),
	}

	deleteRepo := &cobra.Command{
		Use:   "delete-repo repo-name",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.DeleteRepo(apiClient, args[0])
		}),
	}

	startCommit := &cobra.Command{
		Use:   "start-commit repo-name [parent-commit-id]",
		Short: "Start a new commit.",
		Long:  "Start a new commit with parent-commit-id as the parent.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 1, Max: 2}, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			parentCommitID := ""
			if len(args) == 2 {
				parentCommitID = args[1]
			}
			commit, err := pfsutil.StartCommit(apiClient, args[0], parentCommitID)
			if err != nil {
				return err
			}
			fmt.Println(commit.Id)
			return nil
		}),
	}

	finishCommit := &cobra.Command{
		Use:   "finish-commit repo-name commit-id",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.FinishCommit(apiClient, args[0], args[1])
		}),
	}

	inspectCommit := &cobra.Command{
		Use:   "inspect-commit repo-name commit-id",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
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
		}),
	}

	listCommit := &cobra.Command{
		Use:   "list-commit repo-name",
		Short: "Return all commits on a repo.",
		Long:  "Return all commits on a repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			commitInfos, err := pfsutil.ListCommit(apiClient, args)
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

	deleteCommit := &cobra.Command{
		Use:   "delete-commit repo-name commit-id",
		Short: "Delete a commit.",
		Long:  "Delete a commit.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.DeleteCommit(apiClient, args[0], args[1])
		}),
	}

	putBlock := &cobra.Command{
		Use:   "put-block",
		Short: "Put a block from stdin",
		Long:  "Put a block from stdin. Directories must exist. commit-id must be a writeable commit.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getDriveAPIClient(address)
			if err != nil {
				return err
			}
			block, err := pfsutil.PutBlock(apiClient, os.Stdin)
			if err != nil {
				return err
			}
			fmt.Println(block.Hash)
			return nil
		}),
	}

	getBlock := &cobra.Command{
		Use:   "get-block hash",
		Short: "Return the contents of a block.",
		Long:  "Return the contents of a block.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getDriveAPIClient(address)
			if err != nil {
				return err
			}
			reader, err := pfsutil.GetBlock(apiClient, args[0], 0)
			if err != nil {
				return err
			}
			_, err = io.Copy(os.Stdout, reader)
			return err
		}),
	}

	inspectBlock := &cobra.Command{
		Use:   "inspect-block hash",
		Short: "Return info about a block.",
		Long:  "Return info about a block.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getDriveAPIClient(address)
			if err != nil {
				return err
			}
			blockInfo, err := pfsutil.InspectBlock(apiClient, args[0])
			if err != nil {
				return err
			}
			if blockInfo == nil {
				return fmt.Errorf("block %s not found", args[2])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintBlockInfoHeader(writer)
			pretty.PrintBlockInfo(writer, blockInfo)
			return writer.Flush()
		}),
	}

	listBlock := &cobra.Command{
		Use:   "list-block",
		Short: "Return the blocks in a directory.",
		Long:  "Return the blocks in a directory.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			apiClient, err := getDriveAPIClient(address)
			if err != nil {
				return err
			}
			blockInfos, err := pfsutil.ListBlock(apiClient)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintBlockInfoHeader(writer)
			for _, blockInfo := range blockInfos {
				pretty.PrintBlockInfo(writer, blockInfo)
			}
			return writer.Flush()
		}),
	}

	mkdir := &cobra.Command{
		Use:   "mkdir repo-name commit-id path/to/dir",
		Short: "Make a directory.",
		Long:  "Make a directory. Parent directories need not exist.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.MakeDirectory(apiClient, args[0], args[1], args[2])
		}),
	}

	putFile := &cobra.Command{
		Use:   "put-file repo-name commit-id path/to/file",
		Short: "Put a file from stdin",
		Long:  "Put a file from stdin. Directories must exist. commit-id must be a writeable commit.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			_, err = pfsutil.PutFile(apiClient, args[0], args[1], args[2], 0, os.Stdin)
			return err
		}),
	}

	getFile := &cobra.Command{
		Use:   "get-file repo-name commit-id path/to/file",
		Short: "Return the contents of a file.",
		Long:  "Return the contents of a file.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.GetFile(apiClient, args[0], args[1], args[2], 0, math.MaxInt64, shard(), os.Stdout)
		}),
	}
	getFile.Flags().IntVarP(&number, "shard", "s", 0, "shard to read from")
	getFile.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	inspectFile := &cobra.Command{
		Use:   "inspect-file repo-name commit-id path/to/file",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			fileInfo, err := pfsutil.InspectFile(apiClient, args[0], args[1], args[2], shard())
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
	inspectFile.Flags().IntVarP(&number, "shard", "s", 0, "shard to read from")
	inspectFile.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	listFile := &cobra.Command{
		Use:   "list-file repo-name commit-id path/to/dir",
		Short: "Return the files in a directory.",
		Long:  "Return the files in a directory.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 2, Max: 3}, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			var path string
			if len(args) == 3 {
				path = args[2]
			}
			fileInfos, err := pfsutil.ListFile(apiClient, args[0], args[1], path, shard())
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
	listFile.Flags().IntVarP(&number, "shard", "s", 0, "shard to read from")
	listFile.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	deleteFile := &cobra.Command{
		Use:   "delete-file repo-name commit-id path/to/file",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.DeleteFile(apiClient, args[0], args[1], args[2])
		}),
	}

	listChange := &cobra.Command{
		Use:   "list-change repo-name commit-id path/to/dir",
		Short: "Return the changes in a directory.",
		Long:  "Return the changes in a directory.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			changeInfos, err := pfsutil.ListChange(apiClient, args[0], args[1], args[2], shard())
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintChangeHeader(writer)
			for _, changeInfo := range changeInfos {
				pretty.PrintChange(writer, changeInfo)
			}
			return writer.Flush()
		}),
	}
	listChange.Flags().IntVarP(&number, "shard", "s", 0, "shard to read from")
	listChange.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	inspectServer := &cobra.Command{
		Use:   "inspect-server server-id",
		Short: "Inspect a server.",
		Long:  "Inspect a server.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			clusterAPIClient, err := getClusterAPIClient(address)
			if err != nil {
				return err
			}
			serverInfo, err := pfsutil.InspectServer(clusterAPIClient, args[0])
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintServerInfoHeader(writer)
			pretty.PrintServerInfo(writer, serverInfo)
			return writer.Flush()
		}),
	}

	listServer := &cobra.Command{
		Use:   "list-server",
		Short: "Return all servers in the cluster.",
		Long:  "Return all servers in the cluster.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			clusterAPIClient, err := getClusterAPIClient(address)
			if err != nil {
				return err
			}
			serverInfos, err := pfsutil.ListServer(clusterAPIClient)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintServerInfoHeader(writer)
			for _, serverInfo := range serverInfos {
				pretty.PrintServerInfo(writer, serverInfo)
			}
			return writer.Flush()
		}),
	}

	mount := &cobra.Command{
		Use:   "mount [mountpoint]",
		Short: "Mount pfs locally.",
		Long:  "Mount pfs locally.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 1}, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			mountPoint := "/pfs"
			if len(args) > 0 {
				mountPoint = args[0]
			}
			mounter := fuse.NewMounter(address, apiClient)
			return mounter.Mount(mountPoint, shard(), nil, nil)
		}),
	}
	mount.Flags().IntVarP(&number, "shard", "s", 0, "shard to read from")
	mount.Flags().IntVarP(&modulus, "modulus", "m", 1, "modulus of the shards")

	var result []*cobra.Command
	result = append(result, createRepo)
	result = append(result, inspectRepo)
	result = append(result, listRepo)
	result = append(result, deleteRepo)
	result = append(result, startCommit)
	result = append(result, finishCommit)
	result = append(result, inspectCommit)
	result = append(result, listCommit)
	result = append(result, deleteCommit)
	result = append(result, mkdir)
	result = append(result, putBlock)
	result = append(result, getBlock)
	result = append(result, inspectBlock)
	result = append(result, listBlock)
	result = append(result, putFile)
	result = append(result, getFile)
	result = append(result, inspectFile)
	result = append(result, listFile)
	result = append(result, deleteFile)
	result = append(result, listChange)
	result = append(result, inspectServer)
	result = append(result, listServer)
	result = append(result, mount)
	return result, nil
}

func getAPIClient(address string) (pfs.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfs.NewAPIClient(clientConn), nil
}

func getDriveAPIClient(address string) (drive.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return drive.NewAPIClient(clientConn), nil
}

func getClusterAPIClient(address string) (pfs.ClusterAPIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfs.NewClusterAPIClient(clientConn), nil
}

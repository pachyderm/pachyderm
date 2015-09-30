package pretty

import (
	"fmt"
	"io"
	"sort"
	"time"

	"go.pedge.io/proto/time"

	"github.com/docker/docker/pkg/units"
	"github.com/pachyderm/pachyderm/src/pfs"
)

func PrintRepoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tCREATED\tSIZE\t\n")
}

func PrintRepoInfo(w io.Writer, repoInfo *pfs.RepoInfo) {
	fmt.Fprintf(w, "%s\t", repoInfo.Repo.Name)
	fmt.Fprintf(
		w,
		"%s ago\t", units.HumanDuration(
			time.Since(
				prototime.TimestampToTime(
					repoInfo.Created,
				),
			),
		),
	)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(repoInfo.SizeBytes)))
}

func PrintCommitInfoHeader(w io.Writer) {
	fmt.Fprint(w, "ID\tPARENT\tSTATUS\tSTARTED\tFINISHED\tTOTAL_SIZE\tDIFF_SIZE\t\n")
}

func PrintCommitInfo(w io.Writer, commitInfo *pfs.CommitInfo) {
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.Id)
	if commitInfo.ParentCommit != nil {
		fmt.Fprintf(w, "%s\t", commitInfo.ParentCommit.Id)
	} else {
		fmt.Fprint(w, "<none>\t")
	}
	if commitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_WRITE {
		fmt.Fprint(w, "writeable\t")
	} else {
		fmt.Fprint(w, "read-only\t")
	}
	fmt.Fprintf(
		w,
		"%s ago\t", units.HumanDuration(
			time.Since(
				prototime.TimestampToTime(
					commitInfo.Started,
				),
			),
		),
	)
	fmt.Fprintf(
		w,
		"%s ago\t", units.HumanDuration(
			time.Since(
				prototime.TimestampToTime(
					commitInfo.Finished,
				),
			),
		),
	)
	fmt.Fprintf(w, "%s\t", units.BytesSize(float64(commitInfo.TotalBytes)))
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(commitInfo.CommitBytes)))
}

func PrintFileInfoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tTYPE\tMODIFIED\tLAST_COMMIT_MODIFIED\tSIZE\tPERMISSIONS\t\n")
}

func PrintFileInfo(w io.Writer, fileInfo *pfs.FileInfo) {
	fmt.Fprintf(w, "%s\t", fileInfo.File.Path)
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		fmt.Fprint(w, "file\t")
	} else {
		fmt.Fprint(w, "dir\t")
	}
	fmt.Fprintf(
		w,
		"%s ago\t", units.HumanDuration(
			time.Since(
				prototime.TimestampToTime(
					fileInfo.Modified,
				),
			),
		),
	)
	fmt.Fprint(w, "-\t")
	fmt.Fprintf(w, "%s\t", units.BytesSize(float64(fileInfo.SizeBytes)))
	fmt.Fprintf(w, "%4d\t\n", fileInfo.Perm)
}

func PrintServerInfoHeader(w io.Writer) {
	fmt.Fprint(w, "ID\tADDRESS\tVERSION\tMASTER\tREPLICA\t\n")
}

func PrintServerInfo(w io.Writer, serverInfo *pfs.ServerInfo) {
	fmt.Fprintf(w, "%s\t", serverInfo.ServerState.Id)
	fmt.Fprintf(w, "%s\t", serverInfo.ServerState.Address)
	fmt.Fprintf(w, "%d\t", serverInfo.ServerState.Version)
	var masters uint64Slice
	for shard := range serverInfo.ServerRole[serverInfo.ServerState.Version].Masters {
		masters = append(masters, shard)
	}
	sort.Sort(masters)
	for i, shard := range masters {
		fmt.Fprintf(w, "%d", shard)
		if i != len(masters)-1 {
			fmt.Fprint(w, ", ")
		}
	}
	fmt.Fprint(w, "\t")
	var replicas uint64Slice
	for shard := range serverInfo.ServerRole[serverInfo.ServerState.Version].Replicas {
		replicas = append(replicas, shard)
	}
	sort.Sort(replicas)
	for i, shard := range replicas {
		fmt.Fprintf(w, "%d", shard)
		if i != len(replicas)-1 {
			fmt.Fprint(w, ", ")
		}
	}
	fmt.Fprint(w, "\t\n")
}

func PrintChangeHeader(w io.Writer) {
	fmt.Fprintf(w, "NAME\t\n")
}

func PrintChange(w io.Writer, change *pfs.Change) {
	fmt.Fprintf(w, "%s\t\n", change.File.Path)
}

func PrintBlockInfoHeader(w io.Writer) {
	fmt.Fprintf(w, "HASH\tCREATED\tSIZE\t\n")
}

func PrintBlockInfo(w io.Writer, blockInfo *pfs.BlockInfo) {
	fmt.Fprintf(w, "%s\t", blockInfo.Block.Hash)
	fmt.Fprintf(
		w,
		"%s ago\t", units.HumanDuration(
			time.Since(
				prototime.TimestampToTime(
					blockInfo.Created,
				),
			),
		),
	)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(blockInfo.SizeBytes)))
}

type uint64Slice []uint64

func (s uint64Slice) Len() int           { return len(s) }
func (s uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s uint64Slice) Less(i, j int) bool { return s[i] < s[j] }

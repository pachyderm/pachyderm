package pretty

import (
	"fmt"
	"io"
	"time"

	"go.pedge.io/proto/time"

	"github.com/docker/go-units"
	. "github.com/pachyderm/pachyderm/src/client/pfs"
)

func PrintRepoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tCREATED\tSIZE\t\n")
}

func PrintRepoInfo(w io.Writer, repoInfo *RepoInfo) {
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
	fmt.Fprint(w, "BRANCH\tID\tPARENT\tSTARTED\tFINISHED\tSIZE\t\n")
}

func PrintCommitInfo(w io.Writer, commitInfo *CommitInfo) {
	fmt.Fprintf(w, "%s\t", commitInfo.Branch)
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.ID)
	if commitInfo.ParentCommit != nil {
		fmt.Fprintf(w, "%s\t", commitInfo.ParentCommit.ID)
	} else {
		fmt.Fprint(w, "<none>\t")
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
	finished := "\t"
	if commitInfo.Finished != nil {
		finished = fmt.Sprintf("%s ago\t", units.HumanDuration(
			time.Since(
				prototime.TimestampToTime(
					commitInfo.Finished,
				),
			),
		))
	}
	fmt.Fprintf(w, finished)
	fmt.Fprintf(w, "%s\t\n", units.BytesSize(float64(commitInfo.SizeBytes)))
}

func PrintFileInfoHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tTYPE\tMODIFIED\tLAST_COMMIT_MODIFIED\tSIZE\tPERMISSIONS\t\n")
}

func PrintFileInfo(w io.Writer, fileInfo *FileInfo) {
	fmt.Fprintf(w, "%s\t", fileInfo.File.Path)
	if fileInfo.FileType == FileType_FILE_TYPE_REGULAR {
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

func PrintBlockInfoHeader(w io.Writer) {
	fmt.Fprintf(w, "HASH\tCREATED\tSIZE\t\n")
}

func PrintBlockInfo(w io.Writer, blockInfo *BlockInfo) {
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

package pfs

import (
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/pkg/units"
)

func PrintRepositoryHeader(w io.Writer) {
	fmt.Fprintln(w, "NAME\t")
}

func PrintRepository(w io.Writer, repository *Repository) {
	fmt.Fprintf(w, "%s\t", repository.Name)
	fmt.Fprintln(w, "")
}

func PrintCommitInfoHeader(w io.Writer) {
	fmt.Fprintln(w, "ID\tPARENT\tSTATUS\tTIME_OPENED\tTIME_CLOSED\tTOTAL_SIZE\tDIFF_SIZE\t")
}

func PrintCommitInfo(w io.Writer, commitInfo *CommitInfo) {
	fmt.Fprintf(w, "%s\t", commitInfo.Commit.Id)
	if commitInfo.ParentCommit != nil {
		fmt.Fprintf(w, "%s\t", commitInfo.ParentCommit.Id)
	} else {
		fmt.Fprint(w, "<none>\t")
	}
	if commitInfo.CommitType == CommitType_COMMIT_TYPE_WRITE {
		fmt.Fprint(w, "writeable\t")
	} else {
		fmt.Fprint(w, "read-only\t")
	}
	fmt.Fprint(w, "-\t")
	fmt.Fprint(w, "-\t")
	fmt.Fprint(w, "-\t")
	fmt.Fprint(w, "-\t")
	fmt.Fprintln(w, "")
}

func PrintFileInfoHeader(w io.Writer) {
	fmt.Fprintln(w, "NAME\tTYPE\tMODIFIED\tLAST_COMMIT_MODIFIED\tSIZE\tPERMISSIONS\t")
}

func PrintFileInfo(w io.Writer, fileInfo *FileInfo) {
	fmt.Fprintf(w, "%s\t", fileInfo.Path.Path)
	if fileInfo.FileType == FileType_FILE_TYPE_REGULAR {
		fmt.Fprint(w, "file\t")
	} else {
		fmt.Fprint(w, "dir\t")
	}
	fmt.Fprintf(w, "%s ago\t", units.HumanDuration(
		time.Since(
			time.Unix(
				fileInfo.LastModified.Seconds,
				int64(fileInfo.LastModified.Nanos),
			),
		),
	))
	fmt.Fprint(w, "-\t")
	fmt.Fprintf(w, "%s\t", units.BytesSize(float64(fileInfo.SizeBytes)))
	fmt.Fprintf(w, "%d\t", fileInfo.Perm)
	fmt.Fprintln(w, "")
}

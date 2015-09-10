package pfs

import (
	"fmt"
	"io"
)

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

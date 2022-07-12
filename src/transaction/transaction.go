package transaction

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

func (r *TransactionRequest) String() string {
	switch {
	case r.CreatePipeline != nil:
		name := "CreatePipeline"
		if r.CreatePipeline.Update {
			name = "UpdatePipeline"
		}
		return fmt.Sprintf("%s{%s}", name, r.CreatePipeline.Pipeline)
	case r.CreateRepo != nil:
		return fmt.Sprintf("CreateRepo{%s}", r.CreateRepo.Repo)
	case r.DeleteRepo != nil:
		return fmt.Sprintf("DeleteRepo{%s}", r.DeleteRepo.Repo)
	case r.StartCommit != nil:
		return fmt.Sprintf("StartCommit{%s}", r.StartCommit.Branch)
	case r.FinishCommit != nil:
		return fmt.Sprintf("FinishCommit{%s}", r.FinishCommit.Commit)
	case r.SquashCommitSet != nil:
		return fmt.Sprintf("SquashCommit{%s}", r.SquashCommitSet.GetCommitSet().GetID())
	case r.CreateBranch != nil:
		return fmt.Sprintf("CreateBranch{%s}", r.CreateBranch.Branch)
	case r.DeleteBranch != nil:
		return fmt.Sprintf("DeleteBranch{%s}", r.DeleteBranch.Branch)
	case r.UpdateJobState != nil:
		return fmt.Sprintf("UpdateJobState{%s, %s}", r.UpdateJobState.Job, r.UpdateJobState.State)
	case r.StopJob != nil:
		return fmt.Sprintf("StopJob{%s}", r.StopJob.Job)
	}
	return fmt.Sprintf("UnknownRequest{%s}", proto.CompactTextString(r))
}

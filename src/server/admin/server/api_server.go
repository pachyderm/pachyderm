package server

import (
	"fmt"
	"io"
	"sync"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
)

type apiServer struct {
	log.Logger
	address        string
	pachClient     *client.APIClient
	pachClientOnce sync.Once
}

func (a *apiServer) Extract(request *admin.ExtractRequest, extractServer admin.API_ExtractServer) error {
	pachClient, err := a.getPachClient()
	if err != nil {
		return err
	}
	ris, err := pachClient.ListRepo(nil)
	if err != nil {
		return err
	}
	for _, ri := range ris {
		if len(ri.Provenance) > 0 {
			continue
		}
		if err := extractServer.Send(&admin.Op{
			Repo: &pfs.CreateRepoRequest{
				Repo:        ri.Repo,
				Provenance:  ri.Provenance,
				Description: ri.Description,
			},
		}); err != nil {
			return err
		}
		cis, err := pachClient.ListCommit(ri.Repo.Name, "", "", 0)
		if err != nil {
			return err
		}
		for _, ci := range sortCommitInfos(cis) {
			// Even without a parent, ParentCommit is used to indicate which
			// repo to make the commit in.
			if ci.ParentCommit == nil {
				ci.ParentCommit = client.NewCommit(ri.Repo.Name, "")
			}
			if err := extractServer.Send(&admin.Op{
				Commit: &pfs.BuildCommitRequest{
					Parent: ci.ParentCommit,
					Tree:   ci.Tree,
					ID:     ci.Commit.ID,
				},
			}); err != nil {
				return err
			}
		}
		bis, err := pachClient.ListBranch(ri.Repo.Name)
		if err != nil {
			return err
		}
		for _, bi := range bis {
			if err := extractServer.Send(&admin.Op{
				Branch: &pfs.CreateBranchRequest{
					Head:   bi.Head,
					Branch: bi.Branch,
				},
			},
			); err != nil {
				return err
			}
		}
	}
	pis, err := pachClient.ListPipeline()
	if err != nil {
		return err
	}
	for _, pi := range pis {
		if err := extractServer.Send(&admin.Op{
			Pipeline: &pps.CreatePipelineRequest{
				Pipeline:           pi.Pipeline,
				Transform:          pi.Transform,
				ParallelismSpec:    pi.ParallelismSpec,
				Egress:             pi.Egress,
				OutputBranch:       pi.OutputBranch,
				ScaleDownThreshold: pi.ScaleDownThreshold,
				ResourceRequests:   pi.ResourceRequests,
				ResourceLimits:     pi.ResourceLimits,
				Input:              pi.Input,
				Description:        pi.Description,
				Incremental:        pi.Incremental,
				CacheSize:          pi.CacheSize,
				EnableStats:        pi.EnableStats,
				Batch:              pi.Batch,
				MaxQueueSize:       pi.MaxQueueSize,
				Service:            pi.Service,
				ChunkSpec:          pi.ChunkSpec,
				DatumTimeout:       pi.DatumTimeout,
				JobTimeout:         pi.JobTimeout,
				Salt:               pi.Salt,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func sortCommitInfos(cis []*pfs.CommitInfo) []*pfs.CommitInfo {
	commitMap := make(map[string]*pfs.CommitInfo)
	for _, ci := range cis {
		commitMap[ci.Commit.ID] = ci
	}
	var result []*pfs.CommitInfo
	for _, ci := range cis {
		if commitMap[ci.Commit.ID] == nil {
			continue
		}
		var localResult []*pfs.CommitInfo
		for ci != nil {
			localResult = append(localResult, ci)
			delete(commitMap, ci.Commit.ID)
			if ci.ParentCommit != nil {
				ci = commitMap[ci.ParentCommit.ID]
			} else {
				ci = nil
			}
		}
		for i := range localResult {
			result = append(result, localResult[len(localResult)-i-1])
		}
	}
	return result
}

func (a *apiServer) Restore(restoreServer admin.API_RestoreServer) (retErr error) {
	ctx := restoreServer.Context()
	pachClient, err := a.getPachClient()
	if err != nil {
		return err
	}
	defer func() {
		for {
			_, err := restoreServer.Recv()
			if err != nil {
				break
			}
		}
		if err := restoreServer.SendAndClose(&types.Empty{}); err != nil && retErr == nil {
			retErr = err
		}
	}()
	for {
		req, err := restoreServer.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		op := req.Op
		switch {
		case op.Object != nil:
		case op.Repo != nil:
			if _, err := pachClient.PfsAPIClient.CreateRepo(ctx, op.Repo); err != nil {
				return fmt.Errorf("error creating repo: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Commit != nil:
			if _, err := pachClient.PfsAPIClient.BuildCommit(ctx, op.Commit); err != nil {
				return fmt.Errorf("error creating commit: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Branch != nil:
			if _, err := pachClient.PfsAPIClient.CreateBranch(ctx, op.Branch); err != nil {
				return fmt.Errorf("error creating branch: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Pipeline != nil:
			if _, err := pachClient.PpsAPIClient.CreatePipeline(ctx, op.Pipeline); err != nil {
				return fmt.Errorf("error creating pipeline: %v", grpcutil.ScrubGRPC(err))
			}
		}
	}
	return nil
}

func (a *apiServer) getPachClient() (*client.APIClient, error) {
	if a.pachClient == nil {
		var onceErr error
		a.pachClientOnce.Do(func() {
			a.pachClient, onceErr = client.NewFromAddress(a.address)
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return a.pachClient, nil
}

package pjs

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiServer struct {
	pjs.UnimplementedAPIServer
	env    Env
	txnEnv *txnenv.TransactionEnv
}

func newAPIServer(env Env) *apiServer {
	return &apiServer{
		env:    env,
		txnEnv: env.TxnEnv,
	}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pjs.CreateJobRequest) (response *pjs.CreateJobResponse, retErr error) {
	var ret pjs.CreateJobResponse
	if request.Input == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing data input")
	}
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		program, err := fileset.ParseID(request.Program)
		if err != nil {
			return errors.EnsureStack(err)
		}
		var inputs []fileset.PinnedFileset
		for _, input := range request.Input {
			inputID, err := fileset.ParseID(input)
			if err != nil {
				return errors.EnsureStack(err)
			}
			inputs = append(inputs, fileset.PinnedFileset(*inputID))
		}
		req := pjsdb.CreateJobRequest{
			Program: fileset.PinnedFileset(*program),
			Inputs:  inputs,
		}
		jobID, err := pjsdb.CreateJob(ctx, sqlTx, req)
		if err != nil {
			return errors.EnsureStack(err)
		}
		ret.Id = &pjs.Job{Id: int64(jobID)}
		return nil
	}); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (a *apiServer) InspectJob(ctx context.Context, req *pjs.InspectJobRequest) (*pjs.InspectJobResponse, error) {
	var jobInfo pjs.JobInfo
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		job, err := pjsdb.GetJob(ctx, sqlTx, pjsdb.JobID(req.Job.Id))
		if err != nil {
			return errors.EnsureStack(err)
		}
		jobInfo = pjs.JobInfo{
			Job: req.Job,
			ParentJob: &pjs.Job{
				Id: int64(job.Parent),
			},
			Program: job.Program.HexString(),
			State:   pjs.JobState_QUEUED,
		}
		for _, filesetID := range job.Inputs {
			jobInfo.Input = append(jobInfo.Input, filesetID.HexString())
		}
		if !job.Processing.IsZero() {
			jobInfo.State = pjs.JobState_PROCESSING
		}
		if !job.Done.IsZero() {
			jobInfo.State = pjs.JobState_DONE
		}
		if jobInfo.State == pjs.JobState_DONE {
			if len(job.Outputs) != 0 {
				jobInfoSuccess := pjs.JobInfo_Success{}
				for _, filesetID := range job.Outputs {
					jobInfo.Input = append(jobInfoSuccess.Output, filesetID.HexString())
				}
				jobInfo.Result = &pjs.JobInfo_Success_{Success: &jobInfoSuccess}
			} else {
				jobInfo.Result = &pjs.JobInfo_Error{
					Error: pjs.JobErrorCode(pjs.JobErrorCode_value[job.Error]),
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &pjs.InspectJobResponse{
		Details: &pjs.JobInfoDetails{
			JobInfo: &jobInfo,
		},
	}, nil
}

package pjs

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiServer struct {
	pjs.UnimplementedAPIServer
	env Env
}

func newAPIServer(env Env) *apiServer {
	return &apiServer{
		env: env,
	}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pjs.CreateJobRequest) (response *pjs.CreateJobResponse, retErr error) {
	var ret pjs.CreateJobResponse
	if request.Input == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing data input")
	}
	program, err := fileset.ParseID(request.Program)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var inputs []fileset.PinnedFileset
	for _, input := range request.Input {
		inputID, err := fileset.ParseID(input)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		inputs = append(inputs, fileset.PinnedFileset(*inputID))
	}
	req := pjsdb.CreateJobRequest{
		Program: fileset.PinnedFileset(*program),
		Inputs:  inputs,
	}
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
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
			if errors.As(err, &pjsdb.JobNotFoundError{}) {
				return status.Errorf(codes.NotFound, "job %d not found", req.Job.Id)
			}
			return errors.EnsureStack(err)
		}
		jobInfo = pjs.JobInfo{
			Job: req.Job,
			ParentJob: &pjs.Job{
				Id: int64(job.Parent),
			},
			Program: job.Program.HexString(),
		}
		for _, filesetID := range job.Inputs {
			jobInfo.Input = append(jobInfo.Input, filesetID.HexString())
		}
		switch {
		case !job.Queued.IsZero() && job.Processing.IsZero() && job.Done.IsZero():
			jobInfo.State = pjs.JobState_QUEUED
		case !job.Queued.IsZero() && !job.Processing.IsZero() && job.Done.IsZero():
			jobInfo.State = pjs.JobState_PROCESSING
		case !job.Queued.IsZero() && !job.Processing.IsZero() && !job.Done.IsZero():
			jobInfo.State = pjs.JobState_DONE
		default:
			return errors.New("the job state is invalid")
		}
		if jobInfo.State == pjs.JobState_DONE {
			if len(job.Outputs) != 0 {
				jobInfoSuccess := pjs.JobInfo_Success{}
				for _, filesetID := range job.Outputs {
					jobInfoSuccess.Output = append(jobInfoSuccess.Output, filesetID.HexString())
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

func (a *apiServer) ProcessQueue(srv pjs.API_ProcessQueueServer) (retErr error) {
	ctx := srv.Context()
	req, err := srv.Recv()
	if err != nil {
		return errors.Wrap(err, "process queue")
	}
	if req.Queue == nil {
		return status.Errorf(codes.InvalidArgument, "first message must pick Queue")
	}

	for {
		var jobID pjsdb.JobID
		if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
			jid, err := pjsdb.DequeueAndProcess(ctx, sqlTx, req.Queue.Id)
			if err != nil {
				return errors.Wrap(err, "dequeue and process")
			}
			jobID = jid
			return nil
		}); err != nil {
			// await here
			if errors.As(err, &pjsdb.DequeueFromEmptyQueueError{}) {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}
		var inputsID []fileset.ID
		if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
			job, err := pjsdb.GetJob(ctx, sqlTx, jobID)
			if err != nil {
				return errors.Wrap(err, "get job")
			}
			inputsID = job.Inputs
			return nil
		}); err != nil {
			return err
		}
		var inputs []string
		for _, filesetID := range inputsID {
			inputs = append(inputs, filesetID.HexString())
		}
		if err := srv.Send(&pjs.ProcessQueueResponse{
			Context: "context is not implemented",
			Input:   inputs,
		}); err != nil {
			return errors.Wrap(err, "send")
		}
		req, err := srv.Recv()
		if err != nil {
			return errors.Wrap(err, "receive")
		}
		if req.Result == nil {
			return status.Errorf(codes.InvalidArgument, "expected Result. HAVE: %v", req)
		}
		if req.GetFailed() {
			if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
				if err := pjsdb.ErrorJob(ctx, sqlTx, jobID, pjs.JobErrorCode_FAILED); err != nil {
					return errors.Wrap(err, "error job")
				}
				return nil
			}); err != nil {
				return err
			}
		} else if out := req.GetSuccess(); out != nil {
			if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
				if err := pjsdb.CompleteJob(ctx, sqlTx, jobID, out.Output); err != nil {
					return errors.Wrap(err, "complete job")
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}
}

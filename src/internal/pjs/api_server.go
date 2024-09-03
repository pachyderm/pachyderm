package pjs

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"

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
	var parent pjsdb.JobID
	if request.Context != "" {
		parent, err = a.resolveJobCtx(ctx, request.Context)
		if err != nil {
			return nil, errors.Wrap(err, "walk job")
		}
	}
	req := pjsdb.CreateJobRequest{
		Parent:  parent,
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
	var jobInfo *pjs.JobInfo
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		job, err := pjsdb.GetJob(ctx, sqlTx, pjsdb.JobID(req.Job.Id))
		if err != nil {
			if errors.As(err, &pjsdb.JobNotFoundError{}) {
				return status.Errorf(codes.NotFound, "job %d not found", req.Job.Id)
			}
			return errors.EnsureStack(err)
		}
		jobInfo, err = toJobInfo(job)
		if err != nil {
			return errors.Wrap(err, "inspect job")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &pjs.InspectJobResponse{
		Details: &pjs.JobInfoDetails{
			JobInfo: jobInfo,
		},
	}, nil
}

func (a *apiServer) ProcessQueue(srv pjs.API_ProcessQueueServer) (retErr error) {
	ctx := srv.Context()
	req, err := srv.Recv()
	if err != nil {
		return errors.Wrap(err, "recieve")
	}
	if req.Queue == nil {
		return status.Errorf(codes.InvalidArgument, "first message must pick Queue")
	}

	for {
		var (
			jobID  pjsdb.JobID
			jobCtx pjsdb.JobContext
		)
		if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
			resp, err := pjsdb.DequeueAndProcess(ctx, sqlTx, req.Queue.Id)
			if err != nil {
				return errors.Wrap(err, "dequeue and process")
			}
			jobID = resp.ID
			jobCtx = resp.JobContext
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
		}, dbutil.WithReadOnly()); err != nil {
			return errors.Wrap(err, "with tx")
		}
		var inputs []string
		for _, filesetID := range inputsID {
			inputs = append(inputs, filesetID.HexString())
		}
		if err := srv.Send(&pjs.ProcessQueueResponse{
			Context: hex.EncodeToString(jobCtx.Token),
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
				return errors.Wrap(err, "with tx")
			}
		} else if out := req.GetSuccess(); out != nil {
			if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
				if err := pjsdb.CompleteJob(ctx, sqlTx, jobID, out.Output); err != nil {
					return errors.Wrap(err, "complete job")
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "with tx")
			}
		}
	}
}

func (a *apiServer) CancelJob(ctx context.Context, req *pjs.CancelJobRequest) (*pjs.CancelJobResponse, error) {
	// handle job context and request validation.
	var id pjsdb.JobID
	if req.Context != "" {
		jid, err := a.resolveJobCtx(ctx, req.Context)
		if err != nil {
			return nil, errors.Wrap(err, "resolve job ctx")
		}
		id = jid
	} else {
		// TODO(PJS): do auth.
	} //nolint:SA9003
	// TODO(PJS): remove this once auth is implemented.
	reqID := pjsdb.JobID(req.Job.Id)
	if id != reqID {
		log.Error(ctx, "job context token does not match requested job.", zap.Int64("request.Job.ID", req.Job.Id))
		id = reqID // defaulting to this for testing until auth is implemented.
	}

	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		if _, err := pjsdb.CancelJob(ctx, sqlTx, id); err != nil {
			return errors.Wrap(err, "cancel job")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "with tx")
	}
	return &pjs.CancelJobResponse{}, nil
}

func (a *apiServer) WalkJob(req *pjs.WalkJobRequest, srv pjs.API_WalkJobServer) (err error) {
	ctx, done := log.SpanContext(srv.Context(), "walkJob")
	defer done(log.Errorp(&err))

	// handle job context and request validation.
	var id pjsdb.JobID
	if req.Context != "" {
		id, err = a.resolveJobCtx(ctx, req.Context)
		if err != nil {
			return errors.Wrap(err, "resolve job ctx")
		}
	} else {
		// TODO(PJS): do auth.
	} //nolint:SA9003
	// TODO(PJS): remove this once auth is implemented.
	reqID := pjsdb.JobID(req.Job.Id)
	if id != reqID {
		log.Error(ctx, "job context token does not match requested job.", zap.Int64("request.Job.ID", req.Job.Id))
		id = reqID // defaulting to this for testing until auth is implemented.
	}

	// walk and stream back results.
	var jobs []pjsdb.Job
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		jobs, err = pjsdb.WalkJob(ctx, sqlTx, id, pjsdb.WalkAlgorithm(req.Algorithm.Number()), req.MaxDepth)
		return errors.Wrap(err, "walk job in pjsdb")
	}, dbutil.WithReadOnly()); err != nil {
		return errors.Wrap(err, "with tx")
	}
	for i, job := range jobs {
		jobInfo, err := toJobInfo(job)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("to job info, iteration=%d/%d", i, len(jobs)))
		}
		resp := &pjs.ListJobResponse{
			Id:   jobInfo.Job,
			Info: jobInfo,
			Details: &pjs.JobInfoDetails{
				JobInfo: jobInfo,
			},
		}
		if err := srv.Send(resp); err != nil {
			return errors.Wrap(err, fmt.Sprintf("send, iteration=%d/%d", i, len(jobs)))
		}
	}
	return nil
}

func (a *apiServer) resolveJobCtx(ctx context.Context, jobCtx string) (id pjsdb.JobID, err error) {
	token, err := pjsdb.JobContextTokenFromHex(jobCtx)
	if err != nil {
		return 0, errors.Wrap(err, "job context token from string")
	}
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		id, err = pjsdb.ResolveJobContext(ctx, sqlTx, token)
		if err != nil {
			// TODO(PJS): handle NotFound with a special error type that returns the right proto error code.
			return errors.Wrap(err, "resolve job ctx in pjsdb")
		}
		return nil
	}, dbutil.WithReadOnly()); err != nil {
		return 0, errors.Wrap(err, "with tx")
	}
	return id, nil
}

func toJobInfo(job pjsdb.Job) (*pjs.JobInfo, error) {
	jobInfo := &pjs.JobInfo{
		Job: &pjs.Job{
			Id: int64(job.ID),
		},
		ParentJob: &pjs.Job{
			Id: int64(job.Parent),
		},
		Program: job.Program.HexString(),
	}
	for _, filesetID := range job.Inputs {
		jobInfo.Input = append(jobInfo.Input, filesetID.HexString())
	}
	switch {
	case job.Done != time.Time{}:
		jobInfo.State = pjs.JobState_DONE
		jobInfo.Result = &pjs.JobInfo_Error{
			Error: pjs.JobErrorCode(pjs.JobErrorCode_value[job.Error]),
		}
		if len(job.Outputs) != 0 {
			jobInfoSuccess := pjs.JobInfo_Success{}
			for _, filesetID := range job.Outputs {
				jobInfoSuccess.Output = append(jobInfoSuccess.Output, filesetID.HexString())
			}
			jobInfo.Result = &pjs.JobInfo_Success_{Success: &jobInfoSuccess}
		}
	case job.Processing != time.Time{}:
		jobInfo.State = pjs.JobState_PROCESSING
	default:
		jobInfo.State = pjs.JobState_QUEUED
	}
	return jobInfo, nil
}

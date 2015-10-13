package persist

import (
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type rethinkAPIServer struct{}

func newRethinkAPIServer() *rethinkAPIServer {
	return &rethinkAPIServer{}
}

// id cannot be set
// a JobStatus of type created will also be created
func (a *rethinkAPIServer) CreateJob(ctx context.Context, request *Job) (*Job, error) {
	return nil, nil
}

func (a *rethinkAPIServer) GetJobByID(ctx context.Context, request *google_protobuf.StringValue) (*Job, error) {
	return nil, nil
}

func (a *rethinkAPIServer) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue) (*Jobs, error) {
	return nil, nil
}

// id cannot be set
func (a *rethinkAPIServer) CreateJobStatus(ctx context.Context, request *JobStatus) (*JobStatus, error) {
	return nil, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetJobStatusesByJobID(ctx context.Context, request *google_protobuf.StringValue) (*JobStatuses, error) {
	return nil, nil
}

// id cannot be set
func (a *rethinkAPIServer) CreateJobLog(ctx context.Context, request *JobLog) (*JobLog, error) {
	return nil, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetJobLogsByJobID(ctx context.Context, request *google_protobuf.StringValue) (*JobLogs, error) {
	return nil, nil
}

// id and previous_id cannot be set
// name must not already exist
func (a *rethinkAPIServer) CreatePipeline(ctx context.Context, request *Pipeline) (*Pipeline, error) {
	return nil, nil
}

// id and previous_id cannot be set
// update by name, name must already exist
func (a *rethinkAPIServer) UpdatePipeline(ctx context.Context, request *Pipeline) (*Pipeline, error) {
	return nil, nil
}

func (a *rethinkAPIServer) GetPipelineByID(ctx context.Context, request *google_protobuf.StringValue) (*Pipeline, error) {
	return nil, nil
}

// ordered by time, latest to earliest
func (a *rethinkAPIServer) GetPipelinesByName(ctx context.Context, request *google_protobuf.StringValue) (*Pipeline, error) {
	return nil, nil
}

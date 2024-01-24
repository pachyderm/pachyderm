package pjs

import (
	"context"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	pjs "github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Env struct {
	DB      *pachsql.DB
	Storage *storage.Server
}

type JobID uint64

type Sum = pachhash.Output

var _ pjs.APIServer = &Server{}

type Server struct {
	env Env

	mu        sync.Mutex
	lastJobID JobID
	jobs      map[JobID]*Job
	queues    map[Sum]*Queue

	pjs.UnimplementedAPIServer
}

func NewServer(env Env) *Server {
	return &Server{
		env: env,

		jobs:   make(map[JobID]*Job),
		queues: make(map[[32]byte]*Queue),
	}
}

func (s *Server) CreateJob(ctx context.Context, req *pjs.CreateJobRequest) (*pjs.CreateJobResponse, error) {
	if req.Input == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing data input")
	}
	jid, _ := s.createJob(req.GetSpec(), req.GetInput())
	return &pjs.CreateJobResponse{
		Id: &pjs.Job{Id: int64(jid)},
	}, nil
}

func (s *Server) InspectJob(ctx context.Context, req *pjs.InspectJobRequest) (*pjs.InspectJobResponse, error) {
	jid := JobID(req.Job.Id)
	s.mu.Lock()
	job, exists := s.jobs[jid]
	s.mu.Unlock()
	if !exists {
		return nil, status.Errorf(codes.NotFound, "job %v not found", req.Job.Id)
	}
	return &pjs.InspectJobResponse{
		// TODO: This is cumbersome, info should be at the top level
		// - id tier0 (always set)
		// - info tier1 (optionally set)
		// - details tier2
		// - ...
		// - detailed details tierN (optionally set)
		// each contains mutually exclusive information, as the service wants to break it down.
		Details: &pjs.JobInfoDetails{
			JobInfo: &pjs.JobInfo{
				Job:   &pjs.Job{Id: int64(jid)},
				Spec:  clone(job.codeSpec),
				Input: &pjs.QueueElement{Data: job.inputData, Filesets: job.inputFilesets},
				State: job.state,
			},
		},
	}, nil
}

func (s *Server) ProcessQueue(srv pjs.API_ProcessQueueServer) (retErr error) {
	ctx := srv.Context()
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	if req.Queue == nil {
		return status.Errorf(codes.InvalidArgument, "first message must pick Queue")
	}

	var queueID Sum
	copy(queueID[:], req.Queue.Id)
	s.mu.Lock()
	if _, exists := s.queues[queueID]; !exists {
		s.queues[queueID] = newQueue()
	}
	queue := s.queues[queueID]
	s.mu.Unlock()

	for {
		jid, err := queue.Await(ctx)
		if err != nil {
			return err
		}
		s.mu.Lock()
		job := s.jobs[jid]
		job.state = pjs.JobState_PROCESSING
		s.mu.Unlock()
		if err := func() error {
			if err := srv.Send(&pjs.ProcessQueueResponse{
				Context: "context is not implemented",
				Input: &pjs.QueueElement{
					Data:     job.inputData,
					Filesets: job.inputFilesets,
				},
			}); err != nil {
				return err
			}
			req, err := srv.Recv()
			if err != nil {
				return err
			}
			if req.Result == nil {
				return status.Errorf(codes.InvalidArgument, "expected Result. HAVE: %v", req)
			}
			if req.GetFailed() {
				s.completeError(jid, pjs.JobErrorCode_FAILED)
			} else if out := req.GetOutput(); out != nil {
				s.completeOk(jid, out)
			}
			return nil
		}(); err != nil {
			s.completeError(jid, pjs.JobErrorCode_DISCONNECTED)
			return err
		}
	}
}

// createJob creates a new job and adds it to the right queue atomically.
// createJob acquires mu
func (s *Server) createJob(codeSpec *anypb.Any, input *pjs.QueueElement) (JobID, *Job) {
	j := &Job{
		codeSpec:      codeSpec,
		inputData:     input.Data,
		inputFilesets: input.Filesets,
		state:         pjs.JobState_QUEUED,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// allocate job id
	s.lastJobID++
	jid := s.lastJobID
	// add job to jobs "table"
	s.jobs[jid] = j
	// create queue if necessary
	specID := j.SpecID()
	if _, exists := s.queues[specID]; !exists {
		s.queues[specID] = newQueue()
	}
	s.queues[specID].Enqueue(jid)
	return jid, j
}

// completeError completes a job with an error
func (s *Server) completeError(jid JobID, ec pjs.JobErrorCode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[jid]
	j.errCode = ec
	j.state = pjs.JobState_DONE
}

// completeOK completes a job with a result
func (s *Server) completeOk(jid JobID, output *pjs.QueueElement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[jid]
	j.output = output
	j.state = pjs.JobState_DONE
}

type Job struct {
	codeSpec      *anypb.Any
	inputData     []byte
	inputFilesets []string

	state   pjs.JobState
	errCode pjs.JobErrorCode
	output  *pjs.QueueElement
}

func (j *Job) Type() string {
	return j.codeSpec.TypeUrl
}

func (j *Job) SpecID() pachhash.Output {
	return CodeSpecID(j.codeSpec)
}

func (j *Job) InputID() pachhash.Output {
	// TODO: need to figure out fileset hashing.
	// Can't hash the handles, have to hash their fingerprints
	return pachhash.Sum(j.inputData)
}

type Queue struct {
	mu     sync.Mutex
	wakeUp chan struct{}
	jobs   []JobID
}

func newQueue() *Queue {
	q := &Queue{
		wakeUp: make(chan struct{}, 1),
	}
	return q
}

func (q *Queue) Enqueue(jid JobID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, jid)
	select {
	case q.wakeUp <- struct{}{}:
	default:
	}
}

func (q *Queue) Await(ctx context.Context) (JobID, error) {
	for {
		if jid := q.dequeue(); jid != 0 {
			return jid, nil
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-q.wakeUp:
		}
	}
}

// dequeue immediately returns JobID(0) if there are no jobs available
func (q *Queue) dequeue() (ret JobID) {
	// TODO: FIFO not LIFO
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) > 0 {
		q.jobs, ret = q.jobs[:len(q.jobs)-1], q.jobs[len(q.jobs)]
	}
	return ret
}

func clone[T proto.Message](x T) T {
	return proto.Clone(x).(T)
}

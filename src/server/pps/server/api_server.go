package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	workerpkg "github.com/pachyderm/pachyderm/src/server/worker"
	"github.com/robfig/cron"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"golang.org/x/sync/errgroup"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	kube_labels "k8s.io/kubernetes/pkg/labels"
)

const (
	// MaxPodsPerChunk is the maximum number of pods we can schedule for each
	// chunk in case of failures.
	MaxPodsPerChunk = 3
	// DefaultUserImage is the image used for jobs when the user does not specify
	// an image.
	DefaultUserImage = "ubuntu:16.04"
)

var (
	trueVal = true
	zeroVal = int64(0)
	suite   = "pachyderm"
)

func newErrJobNotFound(job string) error {
	return fmt.Errorf("job %v not found", job)
}

func newErrPipelineNotFound(pipeline string) error {
	return fmt.Errorf("pipeline %v not found", pipeline)
}

func newErrPipelineExists(pipeline string) error {
	return fmt.Errorf("pipeline %v already exists", pipeline)
}

type errEmptyInput struct {
	error
}

func newErrEmptyInput(commitID string) *errEmptyInput {
	return &errEmptyInput{
		error: fmt.Errorf("job was not started due to empty input at commit %v", commitID),
	}
}

func newErrParentInputsMismatch(parent string) error {
	return fmt.Errorf("job does not have the same set of inputs as its parent %v", parent)
}

type ctxAndCancel struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type apiServer struct {
	log.Logger
	etcdPrefix            string
	hasher                *ppsserver.Hasher
	address               string
	etcdClient            *etcd.Client
	kubeClient            *kube.Client
	pachClient            *client.APIClient
	pachClientOnce        sync.Once
	namespace             string
	workerImage           string
	workerSidecarImage    string
	workerImagePullPolicy string
	storageRoot           string
	storageBackend        string
	storageHostPath       string
	reporter              *metrics.Reporter
	// collections
	pipelines col.Collection
	jobs      col.Collection
}

func merge(from, to map[string]bool) {
	for s := range from {
		to[s] = true
	}
}

func validateNames(names map[string]bool, input *pps.Input) error {
	switch {
	case input.Atom != nil:
		if names[input.Atom.Name] {
			return fmt.Errorf("name %s was used more than once", input.Atom.Name)
		}
		names[input.Atom.Name] = true
	case input.Cron != nil:
		if names[input.Cron.Name] {
			return fmt.Errorf("name %s was used more than once", input.Cron.Name)
		}
		names[input.Cron.Name] = true
	case input.Union != nil:
		for _, input := range input.Union {
			namesCopy := make(map[string]bool)
			merge(names, namesCopy)
			if err := validateNames(namesCopy, input); err != nil {
				return err
			}
			// we defer this because subinputs of a union input are allowed to
			// have conflicting names but other inputs that are, for example,
			// crossed with this union cannot conflict with any of the names it
			// might present
			defer merge(namesCopy, names)
		}
	case input.Cross != nil:
		for _, input := range input.Cross {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *apiServer) validateInput(ctx context.Context, pipelineName string, input *pps.Input, job bool) error {
	pachClient, err := a.getPachClient()
	if err != nil {
		return err
	}
	if err := validateNames(make(map[string]bool), input); err != nil {
		return err
	}
	pachClient = pachClient.WithCtx(ctx) // pachClient will propagate auth info
	repoBranch := make(map[string]string)
	var result error
	pps.VisitInput(input, func(input *pps.Input) {
		if err := func() error {
			set := false
			if input.Atom != nil {
				set = true
				switch {
				case len(input.Atom.Name) == 0:
					return fmt.Errorf("input must specify a name")
				case input.Atom.Name == "out":
					return fmt.Errorf("input cannot be named \"out\", as pachyderm " +
						"already creates /pfs/out to collect job output")
				case input.Atom.Repo == "":
					return fmt.Errorf("input must specify a repo")
				case input.Atom.Branch == "" && !job:
					return fmt.Errorf("input must specify a branch")
				case input.Atom.Commit == "" && job:
					return fmt.Errorf("input must specify a commit")
				case len(input.Atom.Glob) == 0:
					return fmt.Errorf("input must specify a glob")
				}
				if repoBranch[input.Atom.Repo] != "" && repoBranch[input.Atom.Repo] != input.Atom.Branch {
					return fmt.Errorf("cannot use the same repo in multiple inputs with different branches")
				}
				repoBranch[input.Atom.Repo] = input.Atom.Branch
				if job {
					// for jobs we check that the input commit exists
					if _, err := pachClient.InspectCommit(input.Atom.Repo, input.Atom.Commit); err != nil {
						return err
					}
				} else {
					// for pipelines we only check that the repo exists
					if _, err = pachClient.InspectRepo(input.Atom.Repo); err != nil {
						return err
					}
				}
			}
			if input.Cross != nil {
				if set {
					return fmt.Errorf("multiple input types set")
				}
				set = true
			}
			if input.Union != nil {
				if set {
					return fmt.Errorf("multiple input types set")
				}
				set = true
			}
			if input.Cron != nil {
				if set {
					return fmt.Errorf("multiple input types set")
				}
				set = true
				if _, err := cron.Parse(input.Cron.Spec); err != nil {
					return err
				}
				if _, err := pachClient.InspectRepo(input.Cron.Repo); err != nil {
					return err
				}
			}
			if !set {
				return fmt.Errorf("no input set")
			}
			return nil
		}(); err != nil && result == nil {
			result = err
		}
	})
	return result
}

func validateTransform(transform *pps.Transform) error {
	if len(transform.Cmd) == 0 {
		return fmt.Errorf("no cmd set")
	}
	return nil
}

func (a *apiServer) validateJob(ctx context.Context, jobInfo *pps.JobInfo) error {
	if err := validateTransform(jobInfo.Transform); err != nil {
		return err
	}
	return a.validateInput(ctx, jobInfo.Pipeline.Name, jobInfo.Input, true)
}

func translateJobInputs(inputs []*pps.JobInput) *pps.Input {
	result := &pps.Input{}
	for _, input := range inputs {
		result.Cross = append(result.Cross,
			&pps.Input{
				Atom: &pps.AtomInput{
					Name:   input.Name,
					Repo:   input.Commit.Repo.Name,
					Commit: input.Commit.ID,
					Glob:   input.Glob,
					Lazy:   input.Lazy,
				},
			})
	}
	return result
}

func untranslateJobInputs(input *pps.Input) []*pps.JobInput {
	var result []*pps.JobInput
	if input.Cross != nil {
		for _, input := range input.Cross {
			if input.Atom == nil {
				return nil
			}
			result = append(result, &pps.JobInput{
				Name:   input.Atom.Name,
				Commit: client.NewCommit(input.Atom.Repo, input.Atom.Commit),
				Glob:   input.Atom.Glob,
				Lazy:   input.Atom.Lazy,
			})
		}
	}
	return result
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	// First translate Inputs field to Input field.
	if len(request.Inputs) > 0 {
		if request.Input != nil {
			return nil, fmt.Errorf("cannot set both Inputs and Input field")
		}
		request.Input = translateJobInputs(request.Inputs)
	}

	job := &pps.Job{uuid.NewWithoutUnderscores()}
	pps.SortInput(request.Input)
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobInfo := &pps.JobInfo{
			Job:             job,
			Transform:       request.Transform,
			Pipeline:        request.Pipeline,
			ParallelismSpec: request.ParallelismSpec,
			Input:           request.Input,
			OutputRepo:      request.OutputRepo,
			OutputBranch:    request.OutputBranch,
			Started:         now(),
			Finished:        nil,
			OutputCommit:    nil,
			Service:         request.Service,
			ParentJob:       request.ParentJob,
			ResourceSpec:    request.ResourceSpec,
			NewBranch:       request.NewBranch,
			Incremental:     request.Incremental,
			Stats:           &pps.ProcessStats{},
			EnableStats:     request.EnableStats,
			Salt:            request.Salt,
			PipelineVersion: request.PipelineVersion,
			Batch:           request.Batch,
		}
		if request.Pipeline != nil {
			pipelineInfo := new(pps.PipelineInfo)
			if err := a.pipelines.ReadWrite(stm).Get(request.Pipeline.Name, pipelineInfo); err != nil {
				return err
			}
			if jobInfo.Salt != pipelineInfo.Salt || jobInfo.PipelineVersion != pipelineInfo.Version {
				return fmt.Errorf("job is made from an outdated version of the pipeline")
			}
			jobInfo.Transform = pipelineInfo.Transform
			jobInfo.ParallelismSpec = pipelineInfo.ParallelismSpec
			jobInfo.OutputRepo = &pfs.Repo{pipelineInfo.Pipeline.Name}
			jobInfo.OutputBranch = pipelineInfo.OutputBranch
			jobInfo.Egress = pipelineInfo.Egress
			jobInfo.ResourceSpec = pipelineInfo.ResourceSpec
			jobInfo.Incremental = pipelineInfo.Incremental
			jobInfo.EnableStats = pipelineInfo.EnableStats
		} else {
			if jobInfo.OutputRepo == nil {
				jobInfo.OutputRepo = &pfs.Repo{job.ID}
			}
		}
		if err := a.validateJob(ctx, jobInfo); err != nil {
			return err
		}
		return a.updateJobState(stm, jobInfo, pps.JobState_JOB_STARTING)
	})
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	jobs := a.jobs.ReadOnly(ctx)

	if request.BlockState {
		watcher, err := jobs.WatchOne(request.Job.ID)
		if err != nil {
			return nil, err
		}
		defer watcher.Close()

		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return nil, fmt.Errorf("the stream for job updates closed unexpectedly")
			}
			switch ev.Type {
			case watch.EventError:
				return nil, ev.Err
			case watch.EventDelete:
				return nil, fmt.Errorf("job %s was deleted", request.Job.ID)
			case watch.EventPut:
				var jobID string
				var jobInfo pps.JobInfo
				if err := ev.Unmarshal(&jobID, &jobInfo); err != nil {
					return nil, err
				}
				if jobStateToStopped(jobInfo.State) {
					return &jobInfo, nil
				}
			}
		}
	}

	jobInfo := new(pps.JobInfo)
	if err := jobs.Get(request.Job.ID, jobInfo); err != nil {
		return nil, err
	}
	if jobInfo.Input == nil {
		jobInfo.Input = translateJobInputs(jobInfo.Inputs)
	}
	// If the job is running we fill in WorkerStatus field, otherwise we just
	// return the jobInfo.
	if jobInfo.State != pps.JobState_JOB_RUNNING {
		return jobInfo, nil
	}
	workerPoolID := ppsserver.PipelineRcName(jobInfo.Pipeline.Name, jobInfo.PipelineVersion)
	workerStatus, err := status(ctx, workerPoolID, a.etcdClient, a.etcdPrefix)
	if err != nil {
		logrus.Errorf("failed to get worker status with err: %s", err.Error())
	} else {
		// It's possible that the workers might be working on datums for other
		// jobs, we omit those since they're not part of the status for this
		// job.
		for _, status := range workerStatus {
			if status.JobID == jobInfo.Job.ID {
				jobInfo.WorkerStatus = append(jobInfo.WorkerStatus, status)
			}
		}
	}
	return jobInfo, nil
}

func (a *apiServer) ListJob(ctx context.Context, request *pps.ListJobRequest) (response *pps.JobInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.JobInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.JobInfo), client.MaxListItemsLog)
			a.Log(request, &pps.JobInfos{response.JobInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	jobs := a.jobs.ReadOnly(ctx)
	var iter col.Iterator
	var err error
	if request.Pipeline != nil {
		iter, err = jobs.GetByIndex(ppsdb.JobsPipelineIndex, request.Pipeline)
	} else {
		iter, err = jobs.List()
	}
	if err != nil {
		return nil, err
	}

	var jobInfos []*pps.JobInfo
	for {
		var jobID string
		var jobInfo pps.JobInfo
		ok, err := iter.Next(&jobID, &jobInfo)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if jobInfo.Input == nil {
			jobInfo.Input = translateJobInputs(jobInfo.Inputs)
		}
		jobInfos = append(jobInfos, &jobInfo)
	}

	return &pps.JobInfos{jobInfos}, nil
}

func (a *apiServer) DeleteJob(ctx context.Context, request *pps.DeleteJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.jobs.ReadWrite(stm).Delete(request.Job.ID)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) StopJob(ctx context.Context, request *pps.StopJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.jobs.ReadWrite(stm)
		jobInfo := new(pps.JobInfo)
		if err := jobs.Get(request.Job.ID, jobInfo); err != nil {
			return err
		}
		return a.updateJobState(stm, jobInfo, pps.JobState_JOB_KILLED)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) RestartDatum(ctx context.Context, request *pps.RestartDatumRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: request.Job,
	})
	if err != nil {
		return nil, err
	}
	workerPoolID := ppsserver.PipelineRcName(jobInfo.Pipeline.Name, jobInfo.PipelineVersion)
	if err := cancel(ctx, workerPoolID, a.etcdClient, a.etcdPrefix, request.Job.ID, request.DataFilters); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) ListDatum(ctx context.Context, request *pps.ListDatumRequest) (response *pps.ListDatumResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.DatumInfos) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.DatumInfos), client.MaxListItemsLog)
			logResponse := &pps.ListDatumResponse{
				TotalPages: response.TotalPages,
				Page:       response.Page,
				DatumInfos: response.DatumInfos[:client.MaxListItemsLog],
			}
			a.Log(request, logResponse, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())
	response = &pps.ListDatumResponse{}
	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: &pps.Job{
			ID: request.Job.ID,
		},
	})
	if err != nil {
		return nil, err
	}
	pachClient, err := a.getPachClient()
	if err != nil {
		return nil, err
	}
	pfsClient := pachClient.PfsAPIClient
	getTotalPages := func(totalSize int) int64 {
		return int64(math.Ceil(float64(totalSize) / float64(request.PageSize)))
	}
	getPageBounds := func(totalSize int) (int, int, error) {
		start := int(request.Page * request.PageSize)
		if start > totalSize-1 {
			return 0, 0, io.EOF
		}
		end := start + int(request.PageSize)
		if totalSize < end {
			end = totalSize
		}
		return start, end, nil
	}
	df, err := workerpkg.NewDatumFactory(ctx, pfsClient, jobInfo.Input)
	if err != nil {
		return nil, err
	}
	if jobInfo.StatsCommit == nil {
		start := 0
		end := df.Len()
		if request.PageSize > 0 {
			var err error
			start, end, err = getPageBounds(df.Len())
			if err != nil {
				return nil, err
			}
			response.Page = request.Page
			response.TotalPages = getTotalPages(df.Len())
		}
		var datumInfos []*pps.DatumInfo
		for i := start; i < end; i++ {
			datum := df.Datum(i)
			id := workerpkg.HashDatum(jobInfo.Pipeline.Name, jobInfo.Salt, datum)
			datumInfo := &pps.DatumInfo{
				Datum: &pps.Datum{
					ID:  id,
					Job: jobInfo.Job,
				},
				State: pps.DatumState_STARTING,
			}
			for _, input := range datum {
				datumInfo.Data = append(datumInfo.Data, input.FileInfo)
			}
			datumInfos = append(datumInfos, datumInfo)
		}
		response.DatumInfos = datumInfos
		return response, nil
	}

	// List the files under / to get all the datums
	file := &pfs.File{
		Commit: jobInfo.StatsCommit,
		Path:   "/",
	}
	allFileInfos, err := pfsClient.ListFile(ctx, &pfs.ListFileRequest{file, true})
	if err != nil {
		return nil, err
	}

	var datumFileInfos []*pfs.FileInfo
	// Omit files at the top level that correspond to aggregate job stats
	blacklist := map[string]bool{
		"stats": true,
		"logs":  true,
		"pfs":   true,
	}
	pathToDatumHash := func(path string) (string, error) {
		_, datumHash := filepath.Split(path)
		if _, ok := blacklist[datumHash]; ok {
			return "", fmt.Errorf("value %v is not a datum hash", datumHash)
		}
		return datumHash, nil
	}
	for _, fileInfo := range allFileInfos.FileInfo {
		if _, err := pathToDatumHash(fileInfo.File.Path); err != nil {
			// not a datum
			continue
		}
		datumFileInfos = append(datumFileInfos, fileInfo)
	}
	// Sort results (failed first)
	sort.Sort(byDatumState(datumFileInfos))
	if request.PageSize > 0 {
		response.Page = request.Page
		response.TotalPages = getTotalPages(len(datumFileInfos))
		start, end, err := getPageBounds(len(datumFileInfos))
		if err != nil {
			return nil, err
		}
		datumFileInfos = datumFileInfos[start:end]
	}

	var egGetDatums errgroup.Group
	limiter := limit.New(200)
	datumInfos := make([]*pps.DatumInfo, len(datumFileInfos))
	for index, fileInfo := range datumFileInfos {
		fileInfo := fileInfo
		index := index
		egGetDatums.Go(func() error {
			limiter.Acquire()
			defer limiter.Release()
			datumHash, err := pathToDatumHash(fileInfo.File.Path)
			if err != nil {
				// not a datum
				return nil
			}
			datum, err := a.getDatum(ctx, jobInfo.StatsCommit.Repo.Name, jobInfo.StatsCommit, request.Job.ID, datumHash, df)
			if err != nil {
				return err
			}
			datumInfos[index] = datum
			return nil
		})
	}
	err = egGetDatums.Wait()
	if err != nil {
		return nil, err
	}

	response.DatumInfos = datumInfos
	return response, nil
}

type byDatumState []*pfs.FileInfo

func datumFileToState(f *pfs.FileInfo) pps.DatumState {
	for _, childFileName := range f.Children {
		if childFileName == "skipped" {
			return pps.DatumState_SKIPPED
		}
		if childFileName == "failure" {
			return pps.DatumState_FAILED
		}
	}
	return pps.DatumState_SUCCESS
}

func (a byDatumState) Len() int      { return len(a) }
func (a byDatumState) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDatumState) Less(i, j int) bool {
	iState := datumFileToState(a[i])
	jState := datumFileToState(a[j])
	return iState < jState
}

func (a *apiServer) getDatum(ctx context.Context, repo string, commit *pfs.Commit, jobID string, datumID string, df workerpkg.DatumFactory) (datumInfo *pps.DatumInfo, retErr error) {
	datumInfo = &pps.DatumInfo{
		Datum: &pps.Datum{
			ID:  datumID,
			Job: &pps.Job{jobID},
		},
		State: pps.DatumState_SUCCESS,
	}

	pachClient, err := a.getPachClient()
	if err != nil {
		return nil, err
	}
	pfsClient := pachClient.PfsAPIClient

	// Check if skipped
	stateFile := &pfs.File{
		Commit: commit,
		Path:   fmt.Sprintf("/%v/skipped", datumID),
	}
	_, err = pfsClient.InspectFile(ctx, &pfs.InspectFileRequest{stateFile})
	if err == nil {
		datumInfo.State = pps.DatumState_SKIPPED
		return datumInfo, nil
	} else if !isNotFoundErr(err) {
		return nil, err
	}

	// Check if failed
	stateFile = &pfs.File{
		Commit: commit,
		Path:   fmt.Sprintf("/%v/failure", datumID),
	}
	_, err = pfsClient.InspectFile(ctx, &pfs.InspectFileRequest{stateFile})
	if err == nil {
		datumInfo.State = pps.DatumState_FAILED
	} else if !isNotFoundErr(err) {
		return nil, err
	}

	// Populate stats
	var buffer bytes.Buffer
	if err := pachClient.WithCtx(ctx).GetFile(commit.Repo.Name, commit.ID, fmt.Sprintf("/%v/stats", datumID), 0, 0, &buffer); err != nil {
		return nil, err
	}
	stats := &pps.ProcessStats{}
	err = jsonpb.Unmarshal(&buffer, stats)
	if err != nil {
		return nil, err
	}
	datumInfo.Stats = stats
	buffer.Reset()
	if err := pachClient.WithCtx(ctx).GetFile(commit.Repo.Name, commit.ID, fmt.Sprintf("/%v/index", datumID), 0, 0, &buffer); err != nil {
		return nil, err
	}
	i, err := strconv.Atoi(buffer.String())
	if err != nil {
		return nil, err
	}
	if i >= df.Len() {
		return nil, fmt.Errorf("index %d out of range", i)
	}
	inputs := df.Datum(i)
	for _, input := range inputs {
		datumInfo.Data = append(datumInfo.Data, input.FileInfo)
	}
	datumInfo.PfsState = &pfs.File{
		Commit: commit,
		Path:   fmt.Sprintf("/%v/pfs", datumID),
	}

	return datumInfo, nil
}

func (a *apiServer) InspectDatum(ctx context.Context, request *pps.InspectDatumRequest) (response *pps.DatumInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: &pps.Job{
			ID: request.Datum.Job.ID,
		},
	})
	if err != nil {
		return nil, err
	}

	if !jobInfo.EnableStats {
		return nil, fmt.Errorf("stats not enabled on %v", jobInfo.Pipeline.Name)
	}
	if jobInfo.StatsCommit == nil {
		return nil, fmt.Errorf("job not finished, no stats output yet")
	}
	pachClient, err := a.getPachClient()
	if err != nil {
		return nil, err
	}
	pfsClient := pachClient.PfsAPIClient
	df, err := workerpkg.NewDatumFactory(ctx, pfsClient, jobInfo.Input)
	if err != nil {
		return nil, err
	}

	// Populate datumInfo given a path
	datumInfo, err := a.getDatum(ctx, jobInfo.StatsCommit.Repo.Name, jobInfo.StatsCommit, request.Datum.Job.ID, request.Datum.ID, df)
	if err != nil {
		return nil, err
	}

	return datumInfo, nil
}

func (a *apiServer) lookupRcNameForPipeline(ctx context.Context, pipeline *pps.Pipeline) (string, error) {
	var pipelineInfo pps.PipelineInfo
	err := a.pipelines.ReadOnly(ctx).Get(pipeline.Name, &pipelineInfo)
	if err != nil {
		return "", fmt.Errorf("could not get pipeline information for %s: %s", pipeline.Name, err.Error())
	}
	return ppsserver.PipelineRcName(pipeline.Name, pipelineInfo.Version), nil
}

func (a *apiServer) GetLogs(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	// No deadline in request, but we create one here, since we do expect the call
	// to finish reasonably quickly
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	// Validate request
	if request.Pipeline == nil && request.Job == nil {
		return fmt.Errorf("must set either pipeline or job filter in call to GetLogs")
	}

	// Get list of pods containing logs we're interested in (based on pipeline and
	// job filters)
	var rcName string
	if request.Pipeline != nil {
		// If the user provides a pipeline, get logs from the pipeline RC directly
		var err error
		rcName, err = a.lookupRcNameForPipeline(ctx, request.Pipeline)
		if err != nil {
			return err
		}
	} else if request.Job != nil {
		// If user provides a job, lookup the pipeline from the job info, and then
		// get the pipeline RC
		var jobInfo pps.JobInfo
		err := a.jobs.ReadOnly(ctx).Get(request.Job.ID, &jobInfo)
		if err != nil {
			return fmt.Errorf("could not get job information for %s: %s", request.Job.ID, err.Error())
		}
		rcName, err = a.lookupRcNameForPipeline(ctx, jobInfo.Pipeline)
		if err != nil {
			return err
		}
	}

	pods, err := a.rcPods(rcName)
	if err != nil {
		return fmt.Errorf("could not get pods in rc \"%s\" containing logs: %s", rcName, err.Error())
	}
	if len(pods) == 0 {
		return fmt.Errorf("no pods belonging to the rc \"%s\" were found", rcName)
	}

	// Spawn one goroutine per pod. Each goro writes its pod's logs to a channel
	// and channels are read into the output server in a stable order.
	// (sort the pods to make sure that the order of log lines is stable)
	sort.Sort(podSlice(pods))
	logChs := make([]chan *pps.LogMessage, len(pods))
	errCh := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)
	for i := 0; i < len(pods); i++ {
		logChs[i] = make(chan *pps.LogMessage)
	}
	for i, pod := range pods {
		i, pod := i, pod
		go func() {
			defer close(logChs[i]) // Main thread reads from here, so must close
			// Get full set of logs from pod i
			err := backoff.Retry(func() error {
				result := a.kubeClient.Pods(a.namespace).GetLogs(
					pod.ObjectMeta.Name, &api.PodLogOptions{
						Container: client.PPSWorkerUserContainerName,
					}).Timeout(10 * time.Second).Do()
				fullLogs, err := result.Raw()
				if err != nil {
					if apiStatus, ok := err.(errors.APIStatus); ok &&
						strings.Contains(apiStatus.Status().Message, "PodInitializing") {
						return nil // No logs to collect from this node yet, just skip it
					}
					return err
				}
				// Occasionally, fullLogs is truncated and contains the string
				// 'unexpected stream type ""' at the end. I believe this is a recent
				// bug in k8s (https://github.com/kubernetes/kubernetes/issues/47800)
				// so we're adding a special handler for this corner case.
				// TODO(msteffen) remove this handling once the issue is fixed
				if bytes.HasSuffix(fullLogs, []byte("unexpected stream type \"\"")) {
					return fmt.Errorf("interrupted log stream due to kubernetes/kubernetes/issues/47800")
				}

				// Parse pods' log lines, and filter out irrelevant ones
				scanner := bufio.NewScanner(bytes.NewReader(fullLogs))
				for scanner.Scan() {
					logBytes := scanner.Bytes()
					msg := new(pps.LogMessage)
					if err := jsonpb.Unmarshal(bytes.NewReader(logBytes), msg); err != nil {
						continue
					}

					// Filter out log lines that don't match on pipeline or job
					if request.Pipeline != nil && request.Pipeline.Name != msg.PipelineName {
						continue
					}
					if request.Job != nil && request.Job.ID != msg.JobID {
						continue
					}
					if request.Datum != nil && request.Datum.ID != msg.DatumID {
						continue
					}
					if request.Master != msg.Master {
						continue
					}
					if !workerpkg.MatchDatum(request.DataFilters, msg.Data) {
						continue
					}

					// Log message passes all filters -- return it
					select {
					case logChs[i] <- msg:
					case <-done:
						return nil
					}
				}
				return nil
			}, backoff.New10sBackOff())

			// Used up all retries -- no logs from worker
			if err != nil {
				select {
				case errCh <- err:
				case <-done:
				default:
				}
			}
		}()
	}

nextLogCh:
	for _, logCh := range logChs {
		for {
			msg, ok := <-logCh
			if !ok {
				continue nextLogCh
			}
			if err := apiGetLogsServer.Send(msg); err != nil {
				return err
			}
		}
	}
	select {
	case err := <-errCh:
		return err
	default:
	}
	return nil
}

func (a *apiServer) validatePipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) error {
	if err := a.validateInput(ctx, pipelineInfo.Pipeline.Name, pipelineInfo.Input, false); err != nil {
		return err
	}
	if err := validateTransform(pipelineInfo.Transform); err != nil {
		return err
	}
	if pipelineInfo.ParallelismSpec != nil {
		if pipelineInfo.ParallelismSpec.Constant < 0 {
			return fmt.Errorf("ParallelismSpec.Constant must be > 0")
		}
		if pipelineInfo.ParallelismSpec.Coefficient < 0 {
			return fmt.Errorf("ParallelismSpec.Coefficient must be > 0")
		}
		if pipelineInfo.ParallelismSpec.Constant != 0 &&
			pipelineInfo.ParallelismSpec.Coefficient != 0 {
			return fmt.Errorf("contradictory parallelism strategies: must set at " +
				"most one of ParallelismSpec.Constant and ParallelismSpec.Coefficient")
		}
	}
	if pipelineInfo.OutputBranch == "" {
		return fmt.Errorf("pipeline needs to specify an output branch")
	}
	if _, err := resource.ParseQuantity(pipelineInfo.CacheSize); err != nil {
		return fmt.Errorf("could not parse cacheSize '%s': %v", pipelineInfo.CacheSize, err)
	}
	if pipelineInfo.Incremental {
		pachClient, err := a.getPachClient()
		if err != nil {
			return err
		}
		pfsClient := pachClient.PfsAPIClient
		// for incremental jobs we can't have shared provenance
		var provenance []*pfs.Repo
		for _, commit := range pps.InputCommits(pipelineInfo.Input) {
			provenance = append(provenance, commit.Repo)
		}
		provMap := make(map[string]bool)
		for _, provRepo := range provenance {
			if provMap[provRepo.Name] {
				return fmt.Errorf("can't create an incremental pipeline with inputs that share provenance")
			}
			provMap[provRepo.Name] = true
			resp, err := pfsClient.InspectRepo(ctx, &pfs.InspectRepoRequest{Repo: provRepo})
			if err != nil {
				return err
			}
			for _, provRepo := range resp.RepoInfo.Provenance {
				if provMap[provRepo.Name] {
					return fmt.Errorf("can't create an incremental pipeline with inputs that share provenance")
				}
				provMap[provRepo.Name] = true
			}
		}
	}
	return nil
}

func translatePipelineInputs(inputs []*pps.PipelineInput) *pps.Input {
	result := &pps.Input{}
	for _, input := range inputs {
		var fromCommitID string
		if input.From != nil {
			fromCommitID = input.From.ID
		}
		atomInput := &pps.AtomInput{
			Name:       input.Name,
			Repo:       input.Repo.Name,
			Branch:     input.Branch,
			Glob:       input.Glob,
			Lazy:       input.Lazy,
			FromCommit: fromCommitID,
		}
		result.Cross = append(result.Cross, &pps.Input{
			Atom: atomInput,
		})
	}
	return result
}

// authorizing a pipeline modification varies slightly depending on whether the
// pipeline is being created, updated, or deleted
type pipelineModification uint8

const (
	pipelineCreate pipelineModification = iota
	pipelineUpdate
	pipelineDelete
)

// authorizeModifyPipeline checks if the user indicated by 'ctx' is authorized
// to perform 'modification' on the pipeline in 'info'
func (a *apiServer) authorizeModifyPipeline(ctx context.Context, modification pipelineModification, info *pps.PipelineInfo) error {
	pachClient, err := a.getPachClient()
	if err != nil {
		return err
	}
	authClient := pachClient.AuthAPIClient
	if _, err = authClient.WhoAmI(auth.In2Out(ctx), &auth.WhoAmIRequest{}); err != nil {
		if auth.IsNotActivatedError(err) {
			// TODO(msteffen): Think about how auth will degrade if auth is activated
			// then deactivated. This is potentially insecure.
			return nil // Auth isn't activated, user may proceed
		}
		return err
	}

	// Check that the user is authorized to read all input repos, and write to the
	// output repo (which the pipeline needs to be able to do on the user's
	// behalf)
	// collect inputs by bfs info.Input (info.Inputs is no longer set)
	if len(info.Inputs) > 0 {
		return fmt.Errorf("cannot authorize using PipelineInfo with 'Inputs' field set")
	}
	inputRepos := make(map[string]struct{})
	queue := []*pps.Input{info.Input}
	for len(queue) > 0 {
		in := queue[0]
		queue = queue[1:]
		if in == nil {
			break
		} else if in.Atom != nil {
			inputRepos[in.Atom.Repo] = struct{}{}
		} else if len(in.Cross) > 0 {
			queue = append(queue, in.Cross...)
		} else if len(in.Union) > 0 {
			queue = append(queue, in.Union...)
		} else {
			return fmt.Errorf("cannot authorize pipeline input that is not an atom, a cross, or a union")
		}
	}

	var eg errgroup.Group
	for inputRepo := range inputRepos {
		inputRepo := inputRepo
		eg.Go(func() error {
			resp, err := authClient.Authorize(auth.In2Out(ctx), &auth.AuthorizeRequest{
				Repo:  inputRepo,
				Scope: auth.Scope_READER,
			})
			if err != nil {
				return err
			}
			if !resp.Authorized {
				return fmt.Errorf("not authorized to perform this operation: "+
					"insufficient access to the repo \"%s\"", inputRepo)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	var req *auth.AuthorizeRequest
	if modification == pipelineDelete {
		// To delete a pipeline, you must own the output repo
		req = &auth.AuthorizeRequest{
			Repo:  info.Pipeline.Name,
			Scope: auth.Scope_OWNER,
		}
	} else {
		// Otherwise, you just need to be a writer of the output repo
		req = &auth.AuthorizeRequest{
			Repo:  info.Pipeline.Name,
			Scope: auth.Scope_WRITER,
		}
	}

	// Check that the user is authorized to write to the output repo
	resp, err := authClient.Authorize(auth.In2Out(ctx), req)
	if err != nil {
		if modification == pipelineCreate && strings.HasSuffix(
			err.Error(),
			fmt.Sprintf("ACL not found for repo %s", info.Pipeline.Name)) {
			// Output repo doesn't exist yet. It will be created
			// TODO(msteffen): This is not a great way of handling this error, but
			// right now the way we create ACLs for new repos doesn't really make
			// sense.
			return nil
		}
		return err
	}
	if !resp.Authorized {
		return fmt.Errorf("not authorized to perform this operation: "+
			"insufficient access to the repo \"%s\"", info.Pipeline.Name)
	}
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreatePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	pachClient, err := a.getPachClient()
	if err != nil {
		return nil, err
	}
	authClient := pachClient.AuthAPIClient
	pfsClient := pachClient.PfsAPIClient

	// First translate Inputs field to Input field.
	if len(request.Inputs) > 0 {
		if request.Input != nil {
			return nil, fmt.Errorf("cannot set both Inputs and Input field")
		}
		request.Input = translatePipelineInputs(request.Inputs)
	}

	pipelineInfo := &pps.PipelineInfo{
		Pipeline:           request.Pipeline,
		Version:            1,
		Transform:          request.Transform,
		ParallelismSpec:    request.ParallelismSpec,
		Input:              request.Input,
		OutputBranch:       request.OutputBranch,
		Egress:             request.Egress,
		CreatedAt:          now(),
		ScaleDownThreshold: request.ScaleDownThreshold,
		ResourceSpec:       request.ResourceSpec,
		Description:        request.Description,
		Incremental:        request.Incremental,
		CacheSize:          request.CacheSize,
		EnableStats:        request.EnableStats,
		Salt:               uuid.NewWithoutDashes(),
		Batch:              request.Batch,
	}
	setPipelineDefaults(pipelineInfo)
	var visitErr error
	pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
		if input.Cron != nil {
			if err := pachClient.CreateRepo(input.Cron.Repo); err != nil && !strings.Contains(err.Error(), "already exists") {
				visitErr = err
			}
		}
	})
	if visitErr != nil {
		return nil, visitErr
	}
	if err := a.validatePipeline(ctx, pipelineInfo); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("error dialing auth client: %v", authClient)
	}
	modification := pipelineCreate
	if request.Update {
		modification = pipelineUpdate
	}
	if err := a.authorizeModifyPipeline(ctx, modification, pipelineInfo); err != nil {
		return nil, err
	}
	capabilityResp, err := authClient.GetCapability(auth.In2Out(ctx), &auth.GetCapabilityRequest{})
	if err != nil {
		return nil, fmt.Errorf("error getting capability for the user: %v", err)
	}
	pipelineInfo.Capability = capabilityResp.Capability // User is authorized -- grant capability token to pipeline

	pipelineName := pipelineInfo.Pipeline.Name

	var provenance []*pfs.Repo
	for _, commit := range pps.InputCommits(pipelineInfo.Input) {
		provenance = append(provenance, commit.Repo)
	}

	pps.SortInput(pipelineInfo.Input)
	if request.Update {
		if _, err := a.StopPipeline(ctx, &pps.StopPipelineRequest{request.Pipeline}); err != nil {
			return nil, err
		}
		var oldPipelineInfo pps.PipelineInfo
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelines := a.pipelines.ReadWrite(stm)
			if err := pipelines.Get(pipelineName, &oldPipelineInfo); err != nil {
				return err
			}
			pipelineInfo.Version = oldPipelineInfo.Version + 1
			if !request.Reprocess {
				pipelineInfo.Salt = oldPipelineInfo.Salt
			}
			pipelines.Put(pipelineName, pipelineInfo)
			return nil
		})
		if err != nil {
			return nil, err
		}

		// Revoke the old capability
		if oldPipelineInfo.Capability != "" {
			if _, err := authClient.RevokeAuthToken(auth.In2Out(ctx), &auth.RevokeAuthTokenRequest{
				Token: oldPipelineInfo.Capability,
			}); err != nil && !auth.IsNotActivatedError(err) {
				return nil, fmt.Errorf("error revoking old capability: %v", err)
			}
		}

		// Rename the original output branch to `outputBranch-vN`, where N
		// is the previous version number of the pipeline.
		// We ignore NotFound errors because this pipeline might not have
		// even output anything yet, in which case the output branch
		// may not actually exist.
		if _, err := pfsClient.SetBranch(ctx, &pfs.SetBranchRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{pipelineName},
				ID:   oldPipelineInfo.OutputBranch,
			},
			Branch: fmt.Sprintf("%s-v%d", oldPipelineInfo.OutputBranch, oldPipelineInfo.Version),
		}); err != nil && !isNotFoundErr(err) {
			return nil, err
		}

		if _, err := pfsClient.DeleteBranch(ctx, &pfs.DeleteBranchRequest{
			Repo:   &pfs.Repo{pipelineName},
			Branch: oldPipelineInfo.OutputBranch,
		}); err != nil && !isNotFoundErr(err) {
			return nil, err
		}

		if _, err := a.StartPipeline(ctx, &pps.StartPipelineRequest{request.Pipeline}); err != nil {
			return nil, err
		}

		// We only need to restart downstream pipelines if the provenance
		// of our output repo changed.
		outputRepo := &pfs.Repo{pipelineInfo.Pipeline.Name}
		inspectResp, err := pfsClient.InspectRepo(ctx, &pfs.InspectRepoRequest{
			Repo: outputRepo,
		})
		if err != nil {
			return nil, err
		}

		// Check if the new and old provenance are equal
		provSet := make(map[string]bool)
		for _, oldProv := range inspectResp.RepoInfo.Provenance {
			provSet[oldProv.Name] = true
		}
		for _, newProv := range provenance {
			delete(provSet, newProv.Name)
		}
		provenanceChanged := len(provSet) > 0 || len(inspectResp.RepoInfo.Provenance) != len(provenance)

		if _, err := pfsClient.CreateRepo(auth.In2Out(ctx), &pfs.CreateRepoRequest{
			Repo:       outputRepo,
			Provenance: provenance,
			Update:     true,
		}); err != nil && !isAlreadyExistsErr(err) {
			return nil, err
		}

		if provenanceChanged {

			// Restart all downstream pipelines so they relaunch with the
			// correct provenance.
			repoInfos, err := pfsClient.ListRepo(ctx, &pfs.ListRepoRequest{
				Provenance: []*pfs.Repo{{request.Pipeline.Name}},
			})
			if err != nil {
				return nil, err
			}
			for _, repoInfo := range repoInfos.RepoInfo {
				if _, err := a.StopPipeline(ctx, &pps.StopPipelineRequest{&pps.Pipeline{repoInfo.Repo.Name}}); err != nil {
					if isNotFoundErr(err) {
						continue
					}
					return nil, err
				}
				if _, err := a.StartPipeline(ctx, &pps.StartPipelineRequest{&pps.Pipeline{repoInfo.Repo.Name}}); err != nil {
					return nil, err
				}
			}
		}
	} else {
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelines := a.pipelines.ReadWrite(stm)
			err := pipelines.Create(pipelineName, pipelineInfo)
			if isAlreadyExistsErr(err) {
				return newErrPipelineExists(pipelineName)
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		// Create output repo
		// The pipeline manager also creates the output repo, but we want to
		// also create the repo here to make sure that the output repo is
		// guaranteed to be there after CreatePipeline returns.  This is
		// because it's a very common pattern to create many pipelines in a
		// row, some of which depend on the existence of the output repos
		// of upstream pipelines.
		if _, err := pfsClient.CreateRepo(auth.In2Out(ctx), &pfs.CreateRepoRequest{
			Repo:       &pfs.Repo{pipelineInfo.Pipeline.Name},
			Provenance: provenance,
		}); err != nil && !isAlreadyExistsErr(err) {
			return nil, err
		}
	}

	return &types.Empty{}, nil
}

// setPipelineDefaults sets the default values for a pipeline info
func setPipelineDefaults(pipelineInfo *pps.PipelineInfo) {
	now := time.Now()
	pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
		if input.Atom != nil {
			if input.Atom.Branch == "" {
				input.Atom.Branch = "master"
			}
			if input.Atom.Name == "" {
				input.Atom.Name = input.Atom.Repo
			}
		}
		if input.Cron != nil {
			if input.Cron.Start == nil {
				start, _ := types.TimestampProto(now)
				input.Cron.Start = start
			}
			if input.Cron.Repo == "" {
				input.Cron.Repo = fmt.Sprintf("%s_%s", pipelineInfo.Pipeline.Name, input.Cron.Name)
			}
		}
	})
	if pipelineInfo.OutputBranch == "" {
		// Output branches default to master
		pipelineInfo.OutputBranch = "master"
	}
	if pipelineInfo.CacheSize == "" {
		pipelineInfo.CacheSize = "64M"
	}
	if pipelineInfo.ResourceSpec == nil && pipelineInfo.CacheSize != "" {
		pipelineInfo.ResourceSpec = &pps.ResourceSpec{
			Memory: pipelineInfo.CacheSize,
		}
	}
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	pipelineInfo := new(pps.PipelineInfo)
	if err := a.pipelines.ReadOnly(ctx).Get(request.Pipeline.Name, pipelineInfo); err != nil {
		return nil, err
	}
	if pipelineInfo.Input == nil {
		pipelineInfo.Input = translatePipelineInputs(pipelineInfo.Inputs)
	}
	return pipelineInfo, nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *pps.ListPipelineRequest) (response *pps.PipelineInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.PipelineInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.PipelineInfo), client.MaxListItemsLog)
			a.Log(request, &pps.PipelineInfos{response.PipelineInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	pipelineIter, err := a.pipelines.ReadOnly(ctx).List()
	if err != nil {
		return nil, err
	}

	pipelineInfos := new(pps.PipelineInfos)

	for {
		var pipelineName string
		pipelineInfo := new(pps.PipelineInfo)
		ok, err := pipelineIter.Next(&pipelineName, pipelineInfo)
		if err != nil {
			return nil, err
		}
		if ok {
			if pipelineInfo.Input == nil {
				pipelineInfo.Input = translatePipelineInputs(pipelineInfo.Inputs)
			}
			pipelineInfos.PipelineInfo = append(pipelineInfos.PipelineInfo, pipelineInfo)
		} else {
			break
		}
	}
	return pipelineInfos, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if request.All {
		pipelineInfos, err := a.ListPipeline(ctx, &pps.ListPipelineRequest{})
		if err != nil {
			return nil, err
		}

		for _, pipelineInfo := range pipelineInfos.PipelineInfo {
			request.Pipeline = pipelineInfo.Pipeline
			if _, err := a.deletePipeline(ctx, request); err != nil {
				return nil, err
			}
		}
		return &types.Empty{}, nil
	}
	return a.deletePipeline(ctx, request)
}

func (a *apiServer) deletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *types.Empty, retErr error) {
	pachClient, err := a.getPachClient()
	if err != nil {
		return nil, err
	}
	pipelineInfo, err := a.InspectPipeline(ctx, &pps.InspectPipelineRequest{request.Pipeline})
	if err != nil {
		return nil, fmt.Errorf("pipeline %v was not found: %v", request.Pipeline.Name, err)
	}
	// Check if the caller is authorized to delete this pipeline
	if err := a.authorizeModifyPipeline(ctx, pipelineDelete, pipelineInfo); err != nil {
		return nil, err
	}
	// Revoke the pipeline's capability
	if pipelineInfo.Capability != "" {
		authClient := pachClient.AuthAPIClient
		if err != nil {
			return nil, fmt.Errorf("error dialing auth client: %v", authClient)
		}
		if _, err := authClient.RevokeAuthToken(auth.In2Out(ctx), &auth.RevokeAuthTokenRequest{
			Token: pipelineInfo.Capability,
		}); err != nil && !auth.IsNotActivatedError(err) {
			return nil, fmt.Errorf("error revoking old capability: %v", err)
		}
	}

	iter, err := a.jobs.ReadOnly(ctx).GetByIndex(ppsdb.JobsPipelineIndex, request.Pipeline)
	if err != nil {
		return nil, err
	}

	for {
		var jobID string
		var jobInfo pps.JobInfo
		ok, err := iter.Next(&jobID, &jobInfo)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if request.DeleteJobs {
			if _, err := a.DeleteJob(ctx, &pps.DeleteJobRequest{&pps.Job{jobID}}); err != nil {
				return nil, err
			}
		} else {
			if !jobStateToStopped(jobInfo.State) {
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					var jobInfo pps.JobInfo
					if err := jobs.Get(jobID, &jobInfo); err != nil {
						return err
					}
					// We need to check again here because the job's state
					// might've changed since we first retrieved it
					if !jobStateToStopped(jobInfo.State) {
						jobInfo.State = pps.JobState_JOB_KILLED
					}
					jobs.Put(jobID, &jobInfo)
					return nil
				}); err != nil {
					return nil, err
				}
			}
		}
	}

	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.pipelines.ReadWrite(stm).Delete(request.Pipeline.Name)
	}); err != nil {
		return nil, err
	}

	// Delete output repo
	if request.DeleteRepo {
		pfsClient := pachClient.PfsAPIClient
		if _, err := pfsClient.DeleteRepo(ctx, &pfs.DeleteRepoRequest{
			Repo:  &pfs.Repo{request.Pipeline.Name},
			Force: true,
		}); err != nil {
			return nil, err
		}
	}

	return &types.Empty{}, nil
}

func (a *apiServer) StartPipeline(ctx context.Context, request *pps.StartPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.updatePipelineState(ctx, request.Pipeline.Name, pps.PipelineState_PIPELINE_RUNNING); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) StopPipeline(ctx context.Context, request *pps.StopPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.updatePipelineState(ctx, request.Pipeline.Name, pps.PipelineState_PIPELINE_PAUSED); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) RerunPipeline(ctx context.Context, request *pps.RerunPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	pipelineInfos, err := a.ListPipeline(ctx, &pps.ListPipelineRequest{})
	if err != nil {
		return nil, err
	}

	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		if _, err := a.DeletePipeline(ctx, &pps.DeletePipelineRequest{
			Pipeline: pipelineInfo.Pipeline,
		}); err != nil {
			return nil, err
		}
	}

	jobInfos, err := a.ListJob(ctx, &pps.ListJobRequest{})
	if err != nil {
		return nil, err
	}

	for _, jobInfo := range jobInfos.JobInfo {
		if _, err := a.DeleteJob(ctx, &pps.DeleteJobRequest{jobInfo.Job}); err != nil {
			return nil, err
		}
	}

	return &types.Empty{}, err
}

func (a *apiServer) GarbageCollect(ctx context.Context, request *pps.GarbageCollectRequest) (response *pps.GarbageCollectResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	pachClient, err := a.getPachClient()
	if err != nil {
		return nil, err
	}
	pfsClient := pachClient.PfsAPIClient
	objClient := pachClient.ObjectAPIClient

	// The set of objects that are in use.
	activeObjects := make(map[string]bool)
	var activeObjectsMu sync.Mutex
	// A helper function for adding active objects in a thread-safe way
	addActiveObjects := func(objects ...*pfs.Object) {
		activeObjectsMu.Lock()
		defer activeObjectsMu.Unlock()
		for _, object := range objects {
			if object != nil {
				activeObjects[object.Hash] = true
			}
		}
	}
	// A helper function for adding objects that are actually hash trees,
	// which in turn contain active objects.
	addActiveTree := func(object *pfs.Object) error {
		if object == nil {
			return nil
		}
		addActiveObjects(object)
		getObjectClient, err := objClient.GetObject(ctx, object)
		if err != nil {
			return fmt.Errorf("error getting commit tree: %v", err)
		}

		var buf bytes.Buffer
		if err := grpcutil.WriteFromStreamingBytesClient(getObjectClient, &buf); err != nil {
			return fmt.Errorf("error reading commit tree: %v", err)
		}

		tree, err := hashtree.Deserialize(buf.Bytes())
		if err != nil {
			return err
		}

		return tree.Walk(func(path string, node *hashtree.NodeProto) error {
			if node.FileNode != nil {
				addActiveObjects(node.FileNode.Objects...)
			}
			return nil
		})
	}

	// Get all repos
	repoInfos, err := pfsClient.ListRepo(ctx, &pfs.ListRepoRequest{})
	if err != nil {
		return nil, err
	}

	// Get all commit trees
	limiter := limit.New(100)
	var eg errgroup.Group
	for _, repo := range repoInfos.RepoInfo {
		repo := repo
		commitInfos, err := pfsClient.ListCommit(ctx, &pfs.ListCommitRequest{
			Repo: repo.Repo,
		})
		if err != nil {
			return nil, err
		}
		for _, commit := range commitInfos.CommitInfo {
			commit := commit
			limiter.Acquire()
			eg.Go(func() error {
				defer limiter.Release()
				return addActiveTree(commit.Tree)
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Get all objects referenced by pipeline tags
	pipelineInfos, err := a.ListPipeline(ctx, &pps.ListPipelineRequest{})
	if err != nil {
		return nil, err
	}

	// The set of tags that are active
	activeTags := make(map[string]bool)
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		tags, err := objClient.ListTags(ctx, &pfs.ListTagsRequest{
			Prefix:        client.DatumTagPrefix(pipelineInfo.Salt),
			IncludeObject: true,
		})
		if err != nil {
			return nil, fmt.Errorf("error listing tagged objects: %v", err)
		}

		for resp, err := tags.Recv(); err != io.EOF; resp, err = tags.Recv() {
			resp := resp
			if err != nil {
				return nil, err
			}
			activeTags[resp.Tag] = true
			limiter.Acquire()
			eg.Go(func() error {
				defer limiter.Release()
				return addActiveTree(resp.Object)
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Iterate through all objects.  If they are not active, delete them.
	objects, err := objClient.ListObjects(ctx, &pfs.ListObjectsRequest{})
	if err != nil {
		return nil, err
	}

	var objectsToDelete []*pfs.Object
	deleteObjectsIfMoreThan := func(n int) error {
		if len(objectsToDelete) > n {
			if _, err := objClient.DeleteObjects(ctx, &pfs.DeleteObjectsRequest{
				Objects: objectsToDelete,
			}); err != nil {
				return fmt.Errorf("error deleting objects: %v", err)
			}
			objectsToDelete = []*pfs.Object{}
		}
		return nil
	}
	for object, err := objects.Recv(); err != io.EOF; object, err = objects.Recv() {
		if err != nil {
			return nil, fmt.Errorf("error receiving objects from ListObjects: %v", err)
		}
		if !activeObjects[object.Hash] {
			objectsToDelete = append(objectsToDelete, object)
		}
		// Delete objects in batches
		if err := deleteObjectsIfMoreThan(100); err != nil {
			return nil, err
		}
	}
	if err := deleteObjectsIfMoreThan(0); err != nil {
		return nil, err
	}

	// Iterate through all tags.  If they are not active, delete them
	tags, err := objClient.ListTags(ctx, &pfs.ListTagsRequest{})
	if err != nil {
		return nil, err
	}
	var tagsToDelete []string
	deleteTagsIfMoreThan := func(n int) error {
		if len(tagsToDelete) > n {
			if _, err := objClient.DeleteTags(ctx, &pfs.DeleteTagsRequest{
				Tags: tagsToDelete,
			}); err != nil {
				return fmt.Errorf("error deleting tags: %v", err)
			}
			tagsToDelete = []string{}
		}
		return nil
	}
	for resp, err := tags.Recv(); err != io.EOF; resp, err = tags.Recv() {
		if err != nil {
			return nil, fmt.Errorf("error receiving tags from ListTags: %v", err)
		}
		if !activeTags[resp.Tag] {
			tagsToDelete = append(tagsToDelete, resp.Tag)
		}
		if err := deleteTagsIfMoreThan(100); err != nil {
			return nil, err
		}
	}
	if err := deleteTagsIfMoreThan(0); err != nil {
		return nil, err
	}

	if err := a.incrementGCGeneration(ctx); err != nil {
		return nil, err
	}

	return &pps.GarbageCollectResponse{}, nil
}

// incrementGCGeneration increments the GC generation number in etcd
func (a *apiServer) incrementGCGeneration(ctx context.Context) error {
	resp, err := a.etcdClient.Get(ctx, client.GCGenerationKey)
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		// If the generation number does not exist, create it.
		// It's important that the new generation is 1, as the first
		// generation is assumed to be 0.
		if _, err := a.etcdClient.Put(ctx, client.GCGenerationKey, "1"); err != nil {
			return err
		}
	} else {
		oldGen, err := strconv.Atoi(string(resp.Kvs[0].Value))
		if err != nil {
			return err
		}
		newGen := oldGen + 1
		if _, err := a.etcdClient.Put(ctx, client.GCGenerationKey, strconv.Itoa(newGen)); err != nil {
			return err
		}
	}
	return nil
}

func isAlreadyExistsErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

// pipelineStateToStopped defines what pipeline states are "stopped"
// states, meaning that pipelines in this state should not be managed
// by pipelineManager
func pipelineStateToStopped(state pps.PipelineState) bool {
	switch state {
	case pps.PipelineState_PIPELINE_STARTING:
		return false
	case pps.PipelineState_PIPELINE_RUNNING:
		return false
	case pps.PipelineState_PIPELINE_RESTARTING:
		return false
	case pps.PipelineState_PIPELINE_PAUSED:
		return true
	case pps.PipelineState_PIPELINE_FAILURE:
		return true
	default:
		panic(fmt.Sprintf("unrecognized pipeline state: %s", state))
	}
}

func (a *apiServer) updatePipelineState(ctx context.Context, pipelineName string, state pps.PipelineState) error {
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(pipelineName, pipelineInfo); err != nil {
			return err
		}
		pipelineInfo.State = state
		pipelines.Put(pipelineName, pipelineInfo)
		return nil
	})
	if isNotFoundErr(err) {
		return newErrPipelineNotFound(pipelineName)
	}
	return err
}

func (a *apiServer) updateJobState(stm col.STM, jobInfo *pps.JobInfo, state pps.JobState) error {
	// Update job counts
	if jobInfo.Pipeline != nil {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(jobInfo.Pipeline.Name, pipelineInfo); err != nil {
			return err
		}
		if pipelineInfo.JobCounts == nil {
			pipelineInfo.JobCounts = make(map[int32]int32)
		}
		if pipelineInfo.JobCounts[int32(jobInfo.State)] != 0 {
			pipelineInfo.JobCounts[int32(jobInfo.State)]--
		}
		pipelineInfo.JobCounts[int32(state)]++
		pipelines.Put(pipelineInfo.Pipeline.Name, pipelineInfo)
	}
	jobInfo.State = state
	jobs := a.jobs.ReadWrite(stm)
	jobs.Put(jobInfo.Job.ID, jobInfo)
	return nil
}

func jobStateToStopped(state pps.JobState) bool {
	switch state {
	case pps.JobState_JOB_STARTING:
		return false
	case pps.JobState_JOB_RUNNING:
		return false
	case pps.JobState_JOB_SUCCESS:
		return true
	case pps.JobState_JOB_FAILURE:
		return true
	case pps.JobState_JOB_KILLED:
		return true
	default:
		panic(fmt.Sprintf("unrecognized job state: %s", state))
	}
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

// RepoNameToEnvString is a helper which uppercases a repo name for
// use in environment variable names.
func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

func (a *apiServer) rcPods(rcName string) ([]api.Pod, error) {
	podList, err := a.kubeClient.Pods(a.namespace).List(api.ListOptions{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: kube_labels.SelectorFromSet(labels(rcName)),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func labels(app string) map[string]string {
	return map[string]string{
		"app":   app,
		"suite": suite,
	}
}

type podSlice []api.Pod

func (s podSlice) Len() int {
	return len(s)
}
func (s podSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s podSlice) Less(i, j int) bool {
	return s[i].ObjectMeta.Name < s[j].ObjectMeta.Name
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}

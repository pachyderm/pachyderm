package worker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/montanaflynn/stats"
	"github.com/robfig/cron"
	"golang.org/x/sync/errgroup"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	pfs_sync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

const (
	// maximumRetriesPerDatum is the maximum number of times each datum
	// can failed to be processed before we declare that the job has failed.
	maximumRetriesPerDatum = 3

	masterLockPath = "_master_worker_lock"

	// The number of datums the master caches
	numCachedDatums = 1000000

	ttl = int64(30)
)

func (a *APIServer) getMasterLogger() *taggedLogger {
	result := &taggedLogger{
		template:  a.logMsgTemplate, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line
	result.template.Master = true
	return result
}

func (a *APIServer) getWorkerLogger() *taggedLogger {
	result := &taggedLogger{
		template:  a.logMsgTemplate, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line
	return result
}

func (logger *taggedLogger) jobLogger(jobID string) *taggedLogger {
	result := logger.clone()
	result.template.JobID = jobID
	return result
}

func (a *APIServer) master() {
	masterLock := dlock.NewDLock(a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt))
	logger := a.getMasterLogger()
	b := backoff.NewInfiniteBackOff()
	// Setting a high backoff so that when this master fails, the other
	// workers are more likely to become the master.
	// Also, we've observed race conditions where StopPipeline would cause
	// a master to restart before it's deleted.  PPS would then get confused
	// by the restart and create the workers again, because the restart would
	// bring the pipeline state from PAUSED to RUNNING.  By setting a high
	// retry interval, the master would be deleted before it gets a chance
	// to restart.
	b.InitialInterval = 10 * time.Second
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // make sure that everything this loop might spawn gets cleaned up
		ctx, err := masterLock.Lock(a.pachClient.AddMetadata(ctx))
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		logger.Logf("Launching worker master process")

		paused := false
		// Set pipeline state to running
		if _, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelineName := a.pipelineInfo.Pipeline.Name
			pipelines := a.pipelines.ReadWrite(stm)
			pipelinePtr := &pps.EtcdPipelineInfo{}
			if err := pipelines.Get(pipelineName, pipelinePtr); err != nil {
				return err
			}
			if pipelinePtr.State == pps.PipelineState_PIPELINE_PAUSED {
				paused = true
				return nil
			}
			pipelinePtr.State = pps.PipelineState_PIPELINE_RUNNING
			return pipelines.Put(pipelineName, pipelinePtr)
		}); err != nil {
			return err
		}
		if paused {
			return fmt.Errorf("can't run master for a paused pipeline")
		}
		return a.jobSpawner(ctx)
	}, b, func(err error, d time.Duration) error {
		logger.Logf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *APIServer) serviceMaster() {
	masterLock := dlock.NewDLock(a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt))
	logger := a.getMasterLogger()
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // make sure that everything this loop might spawn gets cleaned up
		ctx, err := masterLock.Lock(a.pachClient.AddMetadata(ctx))
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		logger.Logf("Launching master process")

		paused := false
		// Set pipeline state to running
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelineName := a.pipelineInfo.Pipeline.Name
			pipelines := a.pipelines.ReadWrite(stm)
			pipelinePtr := &pps.EtcdPipelineInfo{}
			if err := pipelines.Get(pipelineName, pipelinePtr); err != nil {
				return err
			}
			if pipelinePtr.State == pps.PipelineState_PIPELINE_PAUSED {
				paused = true
				return nil
			}
			pipelinePtr.State = pps.PipelineState_PIPELINE_RUNNING
			return pipelines.Put(pipelineName, pipelinePtr)
		}); err != nil {
			return err
		}
		if paused {
			return fmt.Errorf("can't run master for a paused pipeline")
		}
		return a.serviceSpawner(ctx)
	}, b, func(err error, d time.Duration) error {
		logger.Logf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *APIServer) jobInput(commitInfo *pfs.CommitInfo) *pps.Input {
	// branchToCommit maps strings of the form "<repo>/<branch>" to PFS commits
	branchToCommit := make(map[string]*pfs.Commit)
	key := path.Join
	for i, provCommit := range commitInfo.Provenance {
		branchToCommit[key(provCommit.Repo.Name, commitInfo.BranchProvenance[i].Name)] = provCommit
	}
	jobInput := proto.Clone(a.pipelineInfo.Input).(*pps.Input)
	pps.VisitInput(jobInput, func(input *pps.Input) {
		if input.Atom != nil {
			if commit, ok := branchToCommit[key(input.Atom.Repo, input.Atom.Branch)]; ok {
				input.Atom.Commit = commit.ID
			}
		}
		if input.Cron != nil {
			if commit, ok := branchToCommit[key(input.Cron.Repo, "master")]; ok {
				input.Cron.Commit = commit.ID
			}
		}
		if input.Git != nil {
			if commit, ok := branchToCommit[key(input.Git.Name, input.Git.Branch)]; ok {
				input.Git.Commit = commit.ID
			}
		}
	})
	return jobInput
}

// spawnScaleDownGoroutine is a helper function for jobSpawner. It spawns a
// goroutine that waits until this pipeline's ScaleDownThreshold is reached and
// then scales down the pipeline's RC. 'cancel' is a channel that, when closed
// cancels the spawned goroutine.
func (a *APIServer) spawnScaleDownGoroutine() (cancel chan struct{}) {
	logger := a.getMasterLogger()
	cancel = make(chan struct{})
	if a.pipelineInfo.ScaleDownThreshold != nil {
		go func() {
			scaleDownThreshold, err := types.DurationFromProto(a.pipelineInfo.ScaleDownThreshold)
			if err != nil {
				logger.Logf("invalid ScaleDownThreshold: \"%v\"", err)
				return
			}
			select {
			case <-time.After(scaleDownThreshold):
				if err := a.scaleDownWorkers(); err != nil {
					logger.Logf("could not scale down workers: \"%v\"", err)
				}
			case <-cancel:
			}
		}()
	}
	return cancel
}

// makeCronCommits makes commits to a single cron input's repo. It's
// a helper function called by makeCronCommits
func (a *APIServer) makeCronCommits(ctx context.Context, in *pps.Input) error {
	schedule, err := cron.Parse(in.Cron.Spec)
	if err != nil {
		return err // Shouldn't happen, as the input is validated in CreatePipeline
	}
	var tstamp *types.Timestamp
	var buffer bytes.Buffer
	if err := a.pachClient.GetFile(in.Cron.Repo, "master", "time", 0, 0, &buffer); err != nil && !isNilBranchErr(err) {
		return err
	} else if err != nil {
		// File not found, this happens the first time the pipeline is run
		tstamp = in.Cron.Start
	} else {
		if err := jsonpb.UnmarshalString(buffer.String(), tstamp); err != nil {
			return err
		}
	}
	t, err := types.TimestampFromProto(tstamp)
	if err != nil {
		return err
	}
	for {
		t = schedule.Next(t)
		time.Sleep(time.Until(t))
		if _, err := a.pachClient.StartCommit(in.Cron.Repo, "master"); err != nil {
			return err
		}
		timestamp, err := types.TimestampProto(t)
		if err != nil {
			return err
		}
		timeString, err := (&jsonpb.Marshaler{}).MarshalToString(timestamp)
		if err != nil {
			return err
		}
		if err := a.pachClient.DeleteFile(in.Cron.Repo, "master", "time"); err != nil {
			return err
		}
		if _, err := a.pachClient.PutFile(in.Cron.Repo, "master", "time", strings.NewReader(timeString)); err != nil {
			return err
		}
		if err := a.pachClient.FinishCommit(in.Cron.Repo, "master"); err != nil {
			return err
		}
	}
}

func (a *APIServer) jobSpawner(ctx context.Context) error {
	logger := a.getMasterLogger()
	var eg errgroup.Group

	// Spawn one goroutine per cron input, each of which commits to its cron repo
	// per its spec
	pps.VisitInput(a.pipelineInfo.Input, func(in *pps.Input) {
		if in.Cron != nil {
			eg.Go(func() error {
				return a.makeCronCommits(ctx, in)
			})
		}
	})

	// Spawn a goroutine to listen for new commits, and create jobs when they
	// arrive
	eg.Go(func() error {
		commitIter, err := a.pachClient.WithCtx(ctx).SubscribeCommit(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.OutputBranch, "")
		if err != nil {
			return err
		}
		defer commitIter.Close()
		for {
			cancel := a.spawnScaleDownGoroutine()
			commitInfo, err := commitIter.Next()
			// loop will spawn new scale-down goroutine whether we start the job or not,
			// so cancel channel to avoid leaking the current goroutine
			close(cancel)
			if err != nil {
				return err
			}
			if commitInfo.Finished != nil {
				continue
			}
			if len(commitInfo.Provenance) == 1 {
				continue
			}
			// Inspect the commit and check again if it has been finished (it may have
			// been closed since it was queued, e.g. by StopPipeline or StopJob)
			commitInfo, err = a.pachClient.WithCtx(ctx).InspectCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID)
			if err != nil {
				return err
			}
			if commitInfo.Finished != nil {
				continue
			}
			if a.pipelineInfo.ScaleDownThreshold != nil {
				if err := a.scaleUpWorkers(logger); err != nil {
					return err
				}
			}
			job, err := a.pachClient.PpsAPIClient.CreateJob(ctx, &pps.CreateJobRequest{
				Pipeline:        a.pipelineInfo.Pipeline,
				OutputCommit:    commitInfo.Commit,
				Input:           a.jobInput(commitInfo),
				Salt:            a.pipelineInfo.Salt,
				PipelineVersion: a.pipelineInfo.Version,
				EnableStats:     a.pipelineInfo.EnableStats,
				Batch:           a.pipelineInfo.Batch,
				Service:         a.pipelineInfo.Service,
				ChunkSpec:       a.pipelineInfo.ChunkSpec,
				DatumTimeout:    a.pipelineInfo.DatumTimeout,
				JobTimeout:      a.pipelineInfo.JobTimeout,
			})
			if err != nil {
				return err
			}

			jobInfo, err := a.pachClient.PpsAPIClient.InspectJob(ctx, &pps.InspectJobRequest{
				Job: job,
			})
			if err != nil {
				return err
			}

			// Now that the jobInfo is persisted, wait until all input commits are
			// ready, split the input datums into chunks and merge the results of
			// chunks as they're processed
			if err := a.waitJob(ctx, jobInfo, logger); err != nil {
				return err
			}
		}
	})

	return eg.Wait() // if anything goes wrong, retry in master()
}

func (a *APIServer) serviceSpawner(ctx context.Context) error {
	commitIter, err := a.pachClient.WithCtx(ctx).SubscribeCommit(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.OutputBranch, "")
	if err != nil {
		return err
	}
	defer commitIter.Close()
	var serviceCtx context.Context
	var serviceCancel func()
	for {
		commitInfo, err := commitIter.Next()
		if err != nil {
			return err
		}
		if commitInfo.Finished != nil {
			continue
		}

		// Create a job document matching the service's output commit
		jobInput := a.jobInput(commitInfo)
		job, err := a.pachClient.PpsAPIClient.CreateJob(ctx, &pps.CreateJobRequest{
			Pipeline:        a.pipelineInfo.Pipeline,
			Input:           jobInput,
			Salt:            a.pipelineInfo.Salt,
			PipelineVersion: a.pipelineInfo.Version,
			EnableStats:     a.pipelineInfo.EnableStats,
			Batch:           a.pipelineInfo.Batch,
			Service:         a.pipelineInfo.Service,
			ChunkSpec:       a.pipelineInfo.ChunkSpec,
			DatumTimeout:    a.pipelineInfo.DatumTimeout,
			JobTimeout:      a.pipelineInfo.JobTimeout,
		})
		if err != nil {
			return err
		}
		df, err := NewDatumFactory(ctx, a.pachClient.PfsAPIClient, jobInput)
		if err != nil {
			return err
		}
		if df.Len() != 1 {
			return fmt.Errorf("services must have a single datum")
		}
		data := df.Datum(0)
		logger, err := a.getTaggedLogger(ctx, job.ID, data, false)
		puller := filesync.NewPuller()
		dir, err := a.downloadData(logger, data, puller, nil, &pps.ProcessStats{}, nil, "")
		if err != nil {
			return err
		}
		if err := os.MkdirAll(client.PPSInputPrefix, 0666); err != nil {
			return err
		}
		if err := syscall.Unmount(client.PPSInputPrefix, syscall.MNT_DETACH); err != nil {
			logger.Logf("error unmounting %+v", err)
		}
		if err := syscall.Mount(dir, client.PPSInputPrefix, "", syscall.MS_BIND, ""); err != nil {
			return err
		}
		if serviceCancel != nil {
			serviceCancel()
		}
		serviceCtx, serviceCancel = context.WithCancel(ctx)
		defer serviceCancel()
		go func() {
			serviceCtx := serviceCtx
			if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				jobs := a.jobs.ReadWrite(stm)
				jobInfo := &pps.JobInfo{}
				if err := jobs.Get(job.ID, jobInfo); err != nil {
					return err
				}
				jobInfo.State = pps.JobState_JOB_RUNNING
				return jobs.Put(jobInfo.Job.ID, jobInfo)
			}); err != nil {
				logger.Logf("error updating job state: %+v", err)
			}
			err := a.runService(serviceCtx, logger)
			if err != nil {
				logger.Logf("error from runService: %+v", err)
			}
			select {
			case <-serviceCtx.Done():
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobInfo := &pps.JobInfo{}
					if err := jobs.Get(job.ID, jobInfo); err != nil {
						return err
					}
					jobInfo.State = pps.JobState_JOB_SUCCESS
					return jobs.Put(jobInfo.Job.ID, jobInfo)
				}); err != nil {
					logger.Logf("error updating job progress: %+v", err)
				}
				if _, err := a.pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
					Commit: commitInfo.Commit,
				}); err != nil {
					logger.Logf("could not finish output commit: %v", err)
				}
			default:
			}
		}()
	}
}

func plusDuration(x *types.Duration, y *types.Duration) (*types.Duration, error) {
	var xd time.Duration
	var yd time.Duration
	var err error
	if x != nil {
		xd, err = types.DurationFromProto(x)
		if err != nil {
			return nil, err
		}
	}
	if y != nil {
		yd, err = types.DurationFromProto(y)
		if err != nil {
			return nil, err
		}
	}
	return types.DurationProto(xd + yd), nil
}

func (a *APIServer) locks(jobID string) col.Collection {
	return col.NewCollection(a.etcdClient, path.Join(a.etcdPrefix, lockPrefix, jobID), nil, &types.Empty{}, nil)
}

// collectDatum collects the output and stats output from a datum, and merges
// it into the passed trees. It errors if it can't find the tree object for
// this datum, unless failed is true in which case it tolerates missing trees.
func (a *APIServer) collectDatum(ctx context.Context, index int, files []*Input, logger *taggedLogger,
	tree hashtree.OpenHashTree, statsTree hashtree.OpenHashTree, treeMu *sync.Mutex, failed bool) error {
	datumHash := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, files)
	datumID := a.DatumID(files)
	tag := &pfs.Tag{datumHash}
	statsTag := &pfs.Tag{datumHash + statsTagSuffix}

	var eg errgroup.Group
	var subTree hashtree.HashTree
	var statsSubtree hashtree.HashTree
	eg.Go(func() error {
		var err error
		subTree, err = a.getTreeFromTag(ctx, tag)
		if err != nil && !failed {
			return fmt.Errorf("failed to retrieve hashtree after processing for datum %v: %v", files, err)
		}
		return nil
	})
	if a.pipelineInfo.EnableStats {
		eg.Go(func() error {
			var err error
			if statsSubtree, err = a.getTreeFromTag(ctx, statsTag); err != nil {
				logger.Logf("failed to read stats tree, this is non-fatal but will result in some missing stats")
				return nil
			}
			indexObject, length, err := a.pachClient.WithCtx(ctx).PutObject(strings.NewReader(fmt.Sprint(index)))
			if err != nil {
				logger.Logf("failed to write stats tree, this is non-fatal but will result in some missing stats")
				return nil
			}
			treeMu.Lock()
			defer treeMu.Unlock()
			// Add a file to statsTree indicating the index of this
			// datum in the datum factory.
			if err := statsTree.PutFile(fmt.Sprintf("%v/index", datumID), []*pfs.Object{indexObject}, length); err != nil {
				logger.Logf("failed to write index file, this is non-fatal but will result in some missing stats")
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	treeMu.Lock()
	defer treeMu.Unlock()
	if statsSubtree != nil {
		if err := statsTree.Merge(statsSubtree); err != nil {
			logger.Logf("failed to merge into stats tree: %v", err)
		}
	}
	if subTree != nil {
		return tree.Merge(subTree)
	}
	return nil
}

func chunks(df DatumFactory, spec *pps.ChunkSpec, parallelism int) *Chunks {
	if spec == nil {
		spec = &pps.ChunkSpec{}
	}
	if spec.Number == 0 && spec.SizeBytes == 0 {
		spec.Number = int64(df.Len() / (parallelism * 10))
		if spec.Number == 0 {
			spec.Number = 1
		}
	}
	chunks := &Chunks{}
	if spec.Number != 0 {
		for i := spec.Number; i < int64(df.Len()); i += spec.Number {
			chunks.Chunks = append(chunks.Chunks, int64(i))
		}
	} else {
		size := int64(0)
		for i := 0; i < df.Len(); i++ {
			for _, input := range df.Datum(i) {
				size += int64(input.FileInfo.SizeBytes)
			}
			if size > spec.SizeBytes {
				chunks.Chunks = append(chunks.Chunks, int64(i))
			}
		}
	}
	chunks.Chunks = append(chunks.Chunks, int64(df.Len()))
	return chunks
}

func (a *APIServer) blockInputs(ctx context.Context, jobInfo *pps.JobInfo) error {
	var vistErr error
	blockCommit := func(commit *pfs.Commit) {
		if _, err := a.pachClient.PfsAPIClient.InspectCommit(ctx,
			&pfs.InspectCommitRequest{
				Commit: commit,
				Block:  true,
			}); err != nil && vistErr == nil {
			vistErr = fmt.Errorf("error blocking on commit %s/%s: %v",
				commit.Repo.Name, commit.ID, err)
		}
	}
	pps.VisitInput(jobInfo.Input, func(input *pps.Input) {
		if input.Atom != nil && input.Atom.Commit != "" {
			blockCommit(client.NewCommit(input.Atom.Repo, input.Atom.Commit))
		}
		if input.Cron != nil && input.Cron.Commit != "" {
			blockCommit(client.NewCommit(input.Cron.Repo, input.Cron.Commit))
		}
		if input.Git != nil && input.Git.Commit != "" {
			blockCommit(client.NewCommit(input.Git.Name, input.Git.Commit))
		}
	})
	return vistErr
}

// waitJob waits for the job in 'jobInfo' to finish, and then it collects the
// output from the job's workers and merges it into a commit (and may merge
// stats into a commit in the stats branch as well)
func (a *APIServer) waitJob(ctx context.Context, jobInfo *pps.JobInfo, logger *taggedLogger) error {
	ctx, cancel := context.WithCancel(ctx)

	// Watch the output commit to see if it's terminated (KILLED, FAILED, or
	// SUCCESS) and if so, cancel the current context
	go func() {
		backoff.RetryNotify(func() error {
			commitInfo, err := a.pachClient.PfsAPIClient.InspectCommit(ctx,
				&pfs.InspectCommitRequest{
					Commit: jobInfo.OutputCommit,
					Block:  true,
				})
			if err != nil {
				return err
			}
			if commitInfo.Tree == nil {
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					return a.updateJobState(stm, jobInfo, pps.JobState_JOB_KILLED, "")
				}); err != nil {
					return err
				}
				cancel()
			}
			return nil
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			return nil
		})
	}()

	backoff.RetryNotify(func() (retErr error) {
		// block until job inputs are ready
		if err := a.blockInputs(ctx, jobInfo); err != nil {
			return err
		}

		// Create a datum factory pointing at the job's inputs and split up the
		// input data into chunks
		df, err := NewDatumFactory(ctx, a.pachClient.PfsAPIClient, jobInfo.Input)
		if err != nil {
			return err
		}
		parallelism, err := ppsutil.GetExpectedNumWorkers(a.kubeClient, a.pipelineInfo.ParallelismSpec)
		if err != nil {
			return fmt.Errorf("error from GetExpectedNumWorkers: %v")
		}
		chunks := chunks(df, jobInfo.ChunkSpec, parallelism)

		// Read the job document, and either resume (if we're recovering from a
		// crash) or mark it running. Also write the input chunks calculated above
		// into chunksCol
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobID := jobInfo.Job.ID
			jobInfo := &pps.JobInfo{}
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			if jobInfo.State == pps.JobState_JOB_KILLED {
				return nil
			}
			jobInfo.DataTotal = int64(df.Len())
			if err := a.updateJobState(stm, jobInfo, pps.JobState_JOB_RUNNING, ""); err != nil {
				return err
			}
			chunksCol := a.chunks.ReadWrite(stm)
			if err := chunksCol.Get(jobID, chunks); err == nil {
				return nil
			}
			return chunksCol.Put(jobID, chunks)
		}); err != nil {
			return err
		}

		// Watch the chunk locks in order, and merge chunk outputs into commit tree
		locks := a.locks(jobInfo.Job.ID).ReadOnly(ctx)
		tree := hashtree.NewHashTree()
		var statsTree hashtree.OpenHashTree
		if jobInfo.EnableStats {
			statsTree = hashtree.NewHashTree()
		}
		var treeMu sync.Mutex
		limiter := limit.New(100)
		var failedDatumID string
		var eg errgroup.Group
		for i, high := range chunks.Chunks {
			// Watch this chunk's lock and when it's finished, handle the result
			// (merge chunk output into commit trees, fail if chunk failed, etc)
			if err := func() error {
				chunkState := &ChunkState{}
				watcher, err := locks.WatchOne(fmt.Sprint(high))
				if err != nil {
					return err
				}
				defer watcher.Close()
			EventLoop:
				for {
					select {
					case e := <-watcher.Watch():
						var key string
						if err := e.Unmarshal(&key, chunkState); err != nil {
							return err
						}
						if chunkState.State != ChunkState_RUNNING {
							if chunkState.State == ChunkState_FAILED {
								failedDatumID = chunkState.DatumID
							}
							var low int64 // chunk lower bound
							if i > 0 {
								low = chunks.Chunks[i-1]
							}
							high := high // chunk upper bound
							// merge results into output tree
							eg.Go(func() error {
								for i := low; i < high; i++ {
									i := i
									limiter.Acquire()
									eg.Go(func() error {
										defer limiter.Release()
										files := df.Datum(int(i))
										return a.collectDatum(ctx, int(i), files, logger, tree, statsTree, &treeMu, chunkState.State == ChunkState_FAILED)
									})
								}
								return nil
							})
							break EventLoop
						}
					case <-ctx.Done():
						return context.Canceled
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		if err := eg.Wait(); err != nil { // all results have been merged
			return err
		}
		// put output tree into object store
		object, err := a.putTree(ctx, tree)
		if err != nil {
			return err
		}
		// Finish the job's output commit
		_, err = a.pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
			Tree:   object,
		})
		if err != nil {
			return err
		}

		// merge stats into stats commit
		// TODO stats branch should be provenant on inputs, rather than us creating
		// the stats commit here
		var statsCommit *pfs.Commit
		if jobInfo.EnableStats {
			statsObject, err := a.putTree(ctx, statsTree)
			if err != nil {
				return err
			}
			statsCommit, err = a.pachClient.PfsAPIClient.BuildCommit(ctx, &pfs.BuildCommitRequest{
				Parent: &pfs.Commit{
					Repo: jobInfo.OutputRepo,
				},
				Branch: "stats",
				Tree:   statsObject,
			})
			if err != nil {
				return err
			}
		}

		// Handle egress
		if err := a.egress(ctx, logger, jobInfo); err != nil {
			reason := fmt.Sprintf("egress error: %v", err)
			_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				jobs := a.jobs.ReadWrite(stm)
				jobID := jobInfo.Job.ID
				jobInfo := &pps.JobInfo{}
				if err := jobs.Get(jobID, jobInfo); err != nil {
					return err
				}
				jobInfo.Finished = now()
				jobInfo.StatsCommit = statsCommit
				return a.updateJobState(stm, jobInfo, pps.JobState_JOB_FAILURE, reason)
			})
			// returning nil so we don't retry
			logger.Logf("possibly a bug -- returning \"%v\"", err)
			return err
		}

		// Record the job's output commit and 'Finished' timestamp, and mark the job
		// as a SUCCESS
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobID := jobInfo.Job.ID
			jobInfo := &pps.JobInfo{}
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			jobInfo.Finished = now()
			jobInfo.StatsCommit = statsCommit
			if failedDatumID != "" {
				return a.updateJobState(stm, jobInfo, pps.JobState_JOB_FAILURE, fmt.Sprintf("failed to process datum: %v", failedDatumID))
			}
			return a.updateJobState(stm, jobInfo, pps.JobState_JOB_SUCCESS, "")
		})
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in waitJob %v, retrying in %v", err, d)
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				if err == context.DeadlineExceeded {
					reason := fmt.Sprintf("job exceeded timeout (%v)", jobInfo.JobTimeout)
					// Mark the job as failed.
					// Workers subscribe to etcd for this state change to cancel their work
					_, err := col.NewSTM(context.Background(), a.etcdClient, func(stm col.STM) error {
						jobs := a.jobs.ReadWrite(stm)
						jobID := jobInfo.Job.ID
						jobInfo := &pps.JobInfo{}
						if err := jobs.Get(jobID, jobInfo); err != nil {
							return err
						}
						jobInfo.Finished = now()
						err = a.updateJobState(stm, jobInfo, pps.JobState_JOB_FAILURE, reason)
						if err != nil {
							return nil
						}
						return nil
					})
					if err != nil {
						return err
					}
				}
				return err
			}
			return err
		default:
		}
		// Increment the job's restart count
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobID := jobInfo.Job.ID
			jobInfo := &pps.JobInfo{}
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			jobInfo.Restart++
			return jobs.Put(jobID, jobInfo)
		})
		if err != nil {
			logger.Logf("error incrementing job %s's restart count", jobInfo.Job.ID)
		}
		return nil
	})
	return nil
}

func (a *APIServer) egress(ctx context.Context, logger *taggedLogger, jobInfo *pps.JobInfo) error {
	var egressFailureCount int
	return backoff.RetryNotify(func() (retErr error) {
		if jobInfo.Egress != nil {
			logger.Logf("Starting egress upload for job (%v)", jobInfo)
			start := time.Now()
			url, err := obj.ParseURL(jobInfo.Egress.URL)
			if err != nil {
				return err
			}
			objClient, err := obj.NewClientFromURLAndSecret(ctx, url)
			if err != nil {
				return err
			}
			client := client.APIClient{
				PfsAPIClient: a.pachClient.PfsAPIClient,
			}
			client.SetMaxConcurrentStreams(100)
			if err := pfs_sync.PushObj(client, jobInfo.OutputCommit, objClient, url.Object); err != nil {
				return err
			}
			logger.Logf("Completed egress upload for job (%v), duration (%v)", jobInfo, time.Since(start))
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		egressFailureCount++
		if egressFailureCount > 3 {
			return err
		}
		logger.Logf("egress failed: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *APIServer) runService(ctx context.Context, logger *taggedLogger) error {
	return backoff.RetryNotify(func() error {
		return a.runUserCode(ctx, logger, nil, &pps.ProcessStats{}, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return err
		default:
			logger.Logf("error running user code: %+v, retrying in: %+v", err, d)
			return nil
		}
	})
}

func (a *APIServer) aggregateProcessStats(stats []*pps.ProcessStats) (*pps.AggregateProcessStats, error) {
	var downloadTime []float64
	var processTime []float64
	var uploadTime []float64
	var downloadBytes []float64
	var uploadBytes []float64
	for _, s := range stats {
		dt, err := types.DurationFromProto(s.DownloadTime)
		if err != nil {
			return nil, err
		}
		downloadTime = append(downloadTime, float64(dt))
		pt, err := types.DurationFromProto(s.ProcessTime)
		if err != nil {
			return nil, err
		}
		processTime = append(processTime, float64(pt))
		ut, err := types.DurationFromProto(s.UploadTime)
		if err != nil {
			return nil, err
		}
		uploadTime = append(uploadTime, float64(ut))
		downloadBytes = append(downloadBytes, float64(s.DownloadBytes))
		uploadBytes = append(uploadBytes, float64(s.UploadBytes))

	}
	dtAgg, err := a.aggregate(downloadTime)
	if err != nil {
		return nil, err
	}
	ptAgg, err := a.aggregate(processTime)
	if err != nil {
		return nil, err
	}
	utAgg, err := a.aggregate(uploadTime)
	if err != nil {
		return nil, err
	}
	dbAgg, err := a.aggregate(downloadBytes)
	if err != nil {
		return nil, err
	}
	ubAgg, err := a.aggregate(uploadBytes)
	if err != nil {
		return nil, err
	}
	return &pps.AggregateProcessStats{
		DownloadTime:  dtAgg,
		ProcessTime:   ptAgg,
		UploadTime:    utAgg,
		DownloadBytes: dbAgg,
		UploadBytes:   ubAgg,
	}, nil
}

func (a *APIServer) aggregate(datums []float64) (*pps.Aggregate, error) {
	logger := a.getMasterLogger()
	mean, err := stats.Mean(datums)
	if err != nil {
		logger.Logf("error aggregating mean: %v", err)
	}
	stddev, err := stats.StandardDeviation(datums)
	if err != nil {
		logger.Logf("error aggregating std dev: %v", err)
	}
	fifth, err := stats.Percentile(datums, 5)
	if err != nil {
		logger.Logf("error aggregating 5th percentile: %v", err)
	}
	ninetyFifth, err := stats.Percentile(datums, 95)
	if err != nil {
		logger.Logf("error aggregating 95th percentile: %v", err)
	}
	return &pps.Aggregate{
		Count:                 int64(len(datums)),
		Mean:                  mean,
		Stddev:                stddev,
		FifthPercentile:       fifth,
		NinetyFifthPercentile: ninetyFifth,
	}, nil
}

func (a *APIServer) getTreeFromTag(ctx context.Context, tag *pfs.Tag) (hashtree.HashTree, error) {
	var buffer bytes.Buffer
	if err := a.pachClient.WithCtx(ctx).GetTag(tag.Name, &buffer); err != nil {
		return nil, err
	}
	return hashtree.Deserialize(buffer.Bytes())
}

func (a *APIServer) putTree(ctx context.Context, tree hashtree.OpenHashTree) (*pfs.Object, error) {
	finishedTree, err := tree.Finish()
	if err != nil {
		return nil, err
	}

	data, err := hashtree.Serialize(finishedTree)
	if err != nil {
		return nil, err
	}
	object, _, err := a.pachClient.WithCtx(ctx).PutObject(bytes.NewReader(data))
	return object, err
}

func (a *APIServer) scaleDownWorkers() error {
	rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(
		ppsutil.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version),
		metav1.GetOptions{})
	if err != nil {
		return err
	}
	*workerRc.Spec.Replicas = 1
	// When we scale down the workers, we also remove the resource
	// requirements so that the remaining master pod does not take up
	// the resource it doesn't need, since by definition when a pipeline
	// is in scale-down mode, it doesn't process any work.
	if a.pipelineInfo.ResourceRequests != nil {
		workerRc.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
	}
	_, err = rc.Update(workerRc)
	return err
}

func (a *APIServer) scaleUpWorkers(logger *taggedLogger) error {
	rc := a.kubeClient.CoreV1().ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(ppsutil.PipelineRcName(
		a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version), metav1.GetOptions{})
	if err != nil {
		return err
	}
	parallelism, err := ppsutil.GetExpectedNumWorkers(a.kubeClient, a.pipelineInfo.ParallelismSpec)
	if err != nil {
		logger.Logf("error getting number of workers, default to 1 worker: %v", err)
		parallelism = 1
	}
	if *workerRc.Spec.Replicas != int32(parallelism) {
		*workerRc.Spec.Replicas = int32(parallelism)
	}
	// Reset the resource requirements for the RC since the pipeline
	// is in scale-down mode and probably has removed its resource
	// requirements.
	if a.pipelineInfo.ResourceRequests != nil {
		requestsResourceList, err := ppsutil.GetRequestsResourceListFromPipeline(a.pipelineInfo)
		if err != nil {
			return fmt.Errorf("error parsing resource spec; this is likely a bug: %v", err)
		}
		limitsResourceList, err := ppsutil.GetRequestsResourceListFromPipeline(a.pipelineInfo)
		if err != nil {
			return fmt.Errorf("error parsing resource spec; this is likely a bug: %v", err)
		}
		workerRc.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
			Requests: *requestsResourceList,
			Limits:   *limitsResourceList,
		}
	}
	_, err = rc.Update(workerRc)
	return err
}

// getCachedDatum returns whether the given datum (identified by its hash)
// has been processed.
func (a *APIServer) getCachedDatum(hash string) bool {
	_, ok := a.datumCache.Get(hash)
	return ok
}

// setCachedDatum records that the given datum has been processed.
func (a *APIServer) setCachedDatum(hash string) {
	a.datumCache.Add(hash, struct{}{})
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func isNilBranchErr(err error) bool {
	return err != nil &&
		strings.HasPrefix(err.Error(), "the branch \"") &&
		strings.HasSuffix(err.Error(), "\" is nil")
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}

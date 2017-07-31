package worker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/montanaflynn/stats"
	"golang.org/x/sync/errgroup"
	"k8s.io/kubernetes/pkg/api"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/pool"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	pfs_sync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/util"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"google.golang.org/grpc"
)

const (
	// maximumRetriesPerDatum is the maximum number of times each datum
	// can failed to be processed before we declare that the job has failed.
	maximumRetriesPerDatum = 3

	masterLockPath = "_master_worker_lock"

	// The number of datums that will be enqueued on each worker.
	queueSize = 10
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
	b.InitialInterval = 5 * time.Second
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(ctx)

		logger.Logf("Launching worker master process")

		// Set pipeline state to running
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelineName := a.pipelineInfo.Pipeline.Name
			pipelines := a.pipelines.ReadWrite(stm)
			pipelineInfo := new(pps.PipelineInfo)
			if err := pipelines.Get(pipelineName, pipelineInfo); err != nil {
				return err
			}
			pipelineInfo.State = pps.PipelineState_PIPELINE_RUNNING
			pipelines.Put(pipelineName, pipelineInfo)
			return nil
		})
		return a.jobSpawner(ctx, logger)
	}, b, func(err error, d time.Duration) error {
		logger.Logf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

// jobSpawner spawns jobs
func (a *APIServer) jobSpawner(ctx context.Context, logger *taggedLogger) error {
	// Establish connection pool
	pool, err := pool.NewPool(a.kubeClient, a.namespace, ppsserver.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version), client.PachDialOptions()...)
	if err != nil {
		return fmt.Errorf("master: error constructing worker pool: %v; retrying in %v", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			logger.Logf("error closing pool: %v", err)
		}
	}()

	bsf, err := a.newBranchSetFactory(ctx)
	if err != nil {
		return fmt.Errorf("error constructing branch set factory: %v", err)
	}
	defer bsf.Close()
nextInput:
	for {
		// scaleDownCh is closed after we have not received a job for
		// a certain amount of time specified by ScaleDownThreshold.
		scaleDownCh := make(chan struct{})
		if a.pipelineInfo.ScaleDownThreshold != nil {
			scaleDownThreshold, err := types.DurationFromProto(a.pipelineInfo.ScaleDownThreshold)
			if err != nil {
				logger.Logf("error converting scaleDownThreshold: %v", err)
			} else {
				time.AfterFunc(scaleDownThreshold, func() {
					close(scaleDownCh)
				})
			}
		}
		var bs *branchSet
		select {
		case <-ctx.Done():
			return context.Canceled
		case bs = <-bsf.Chan():
			if bs.Err != nil {
				return fmt.Errorf("error from branch set factory: %v", bs.Err)
			}
		case <-scaleDownCh:
			if err := a.scaleDownWorkers(); err != nil {
				logger.Logf("error scaling down workers: %v", err)
			}
			continue nextInput
		}

		// Once we received a job, scale up the workers
		if a.pipelineInfo.ScaleDownThreshold != nil {
			if err := a.scaleUpWorkers(); err != nil {
				logger.Logf("error scaling up workers: %v", err)
			}
		}

		// (create JobInput for new processing job)
		jobInput := proto.Clone(a.pipelineInfo.Input).(*pps.Input)
		var visitErr error
		pps.VisitInput(jobInput, func(input *pps.Input) {
			if input.Atom != nil {
				for _, branch := range bs.Branches {
					if input.Atom.Repo == branch.Head.Repo.Name && input.Atom.Branch == branch.Name {
						input.Atom.Commit = branch.Head.ID
					}
				}
				if input.Atom.Commit == "" {
					visitErr = fmt.Errorf("didn't find input commit for %s/%s", input.Atom.Repo, input.Atom.Branch)
				}
				input.Atom.FromCommit = ""
			}
		})
		if visitErr != nil {
			return visitErr
		}

		jobsRO := a.jobs.ReadOnly(ctx)
		// Check if this input set has already been processed
		jobIter, err := jobsRO.GetByIndex(ppsdb.JobsInputIndex, jobInput)
		if err != nil {
			return err
		}

		// This is doing the same thing as the line above but for 1.4.5 style jobs
		oldJobIter, err := jobsRO.GetByIndex(ppsdb.JobsInputsIndex, untranslateJobInputs(jobInput))
		if err != nil {
			return err
		}

		// Check if any of the jobs in jobIter have been run already. If so, skip
		// this input.
		for {
			var jobID string
			var jobInfo pps.JobInfo
			ok, err := jobIter.Next(&jobID, &jobInfo)
			if err != nil {
				return err
			}
			if !ok {
				ok, err := oldJobIter.Next(&jobID, &jobInfo)
				if err != nil {
					return err
				}
				if !ok {
					break
				}
			}
			if jobInfo.Pipeline.Name == a.pipelineInfo.Pipeline.Name &&
				(jobInfo.Salt == a.pipelineInfo.Salt || (jobInfo.Salt == "" && jobInfo.PipelineVersion == a.pipelineInfo.Version)) {
				switch jobInfo.State {
				case pps.JobState_JOB_STARTING, pps.JobState_JOB_RUNNING:
					if err := a.runJob(ctx, &jobInfo, pool, logger); err != nil {
						return err
					}
				}
				continue nextInput
			}
		}

		// now we need to find the parentJob for this job. The parent job
		// is defined as the job with the same input commits except for the
		// newest input commit which triggered this job. In place of it we
		// use that commit's parent.
		newBranch := bs.Branches[bs.NewBranch]
		newCommitInfo, err := a.pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: newBranch.Head,
		})
		if err != nil {
			return err
		}
		var parentJob *pps.Job
		if newCommitInfo.ParentCommit != nil {
			// recreate our parent's inputs so we can look it up in etcd
			parentJobInput := proto.Clone(jobInput).(*pps.Input)
			pps.VisitInput(parentJobInput, func(input *pps.Input) {
				if input.Atom != nil && input.Atom.Repo == newBranch.Head.Repo.Name && input.Atom.Branch == newBranch.Name {
					input.Atom.Commit = newCommitInfo.ParentCommit.ID
				}
			})
			if visitErr != nil {
				return visitErr
			}
			jobIter, err := jobsRO.GetByIndex(ppsdb.JobsInputIndex, parentJobInput)
			if err != nil {
				return err
			}
			for {
				var jobID string
				var jobInfo pps.JobInfo
				ok, err := jobIter.Next(&jobID, &jobInfo)
				if err != nil {
					return err
				}
				if !ok {
					break
				}
				if jobInfo.Pipeline.Name == a.pipelineInfo.Pipeline.Name &&
					(jobInfo.Salt == a.pipelineInfo.Salt || (jobInfo.Salt == "" && jobInfo.PipelineVersion == a.pipelineInfo.Version)) {
					parentJob = jobInfo.Job
				}
			}
		}

		job, err := a.pachClient.PpsAPIClient.CreateJob(ctx, &pps.CreateJobRequest{
			Pipeline:        a.pipelineInfo.Pipeline,
			Input:           jobInput,
			ParentJob:       parentJob,
			NewBranch:       newBranch,
			Salt:            a.pipelineInfo.Salt,
			PipelineVersion: a.pipelineInfo.Version,
			EnableStats:     a.pipelineInfo.EnableStats,
			Batch:           a.pipelineInfo.Batch,
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

		if err := a.runJob(ctx, jobInfo, pool, logger.jobLogger(job.ID)); err != nil {
			return err
		}
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

// jobManager feeds datums to jobs
func (a *APIServer) runJob(ctx context.Context, jobInfo *pps.JobInfo, pool *pool.Pool, logger *taggedLogger) error {
	pfsClient := a.pachClient.PfsAPIClient
	ppsClient := a.pachClient.PpsAPIClient

	jobID := jobInfo.Job.ID
	var jobStopped bool
	var jobStoppedMutex sync.Mutex
	backoff.RetryNotify(func() (retErr error) {
		// We use a new context for this particular instance of the retry
		// loop, to ensure that all resources are released properly when
		// this job retries.
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if jobInfo.ParentJob != nil {
			// Wait for the parent job to finish, to ensure that output
			// commits are ordered correctly, and that this job doesn't
			// contend for workers with its parent.
			if _, err := ppsClient.InspectJob(ctx, &pps.InspectJobRequest{
				Job:        jobInfo.ParentJob,
				BlockState: true,
			}); err != nil {
				return err
			}
		}

		// Cancel the context and move on to the next job if this job
		// has been manually stopped.
		go func() {
			currentJobInfo, err := ppsClient.InspectJob(ctx, &pps.InspectJobRequest{
				Job:        jobInfo.Job,
				BlockState: true,
			})
			if err != nil {
				logger.Logf("error monitoring job state: %v", err)
				return
			}
			switch currentJobInfo.State {
			case pps.JobState_JOB_KILLED, pps.JobState_JOB_SUCCESS, pps.JobState_JOB_FAILURE:
				jobStoppedMutex.Lock()
				defer jobStoppedMutex.Unlock()
				jobStopped = true
				cancel()
			}
		}()

		var pipelineInfo *pps.PipelineInfo
		if jobInfo.Pipeline != nil {
			var err error
			pipelineInfo, err = ppsClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{
				Pipeline: jobInfo.Pipeline,
			})
			if err != nil {
				return err
			}
		}

		// Set the state of this job to 'RUNNING'
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobInfo := new(pps.JobInfo)
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			return a.updateJobState(stm, jobInfo, pps.JobState_JOB_RUNNING)
		})
		if err != nil {
			return err
		}

		failed := false
		limiter := limit.New(a.numWorkers * queueSize)
		// process all datums
		df, err := newDatumFactory(ctx, pfsClient, jobInfo.Input)
		if err != nil {
			return err
		}
		var newBranchParentCommit *pfs.Commit
		// If this is an incremental job we need to find the parent
		// commit of the new branch.
		if jobInfo.Incremental && jobInfo.NewBranch != nil {
			newBranchCommitInfo, err := pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
				Commit: jobInfo.NewBranch.Head,
			})
			if err != nil {
				return err
			}
			newBranchParentCommit = newBranchCommitInfo.ParentCommit
		}
		tree := hashtree.NewHashTree()
		var statsTree hashtree.OpenHashTree
		if jobInfo.EnableStats {
			statsTree = hashtree.NewHashTree()
		}
		var processStats []*pps.ProcessStats
		var treeMu sync.Mutex

		processedData := int64(0)
		skippedData := int64(0)
		setData := int64(0) // sum of skipped and processed data we've told etcd about
		stats := &pps.ProcessStats{}
		totalData := int64(df.Len())
		var progressMu sync.Mutex
		updateProgress := func(processed, skipped int64, newStats *pps.ProcessStats) {
			progressMu.Lock()
			defer progressMu.Unlock()
			processedData += processed
			skippedData += skipped
			totalProcessedData := processedData + skippedData
			if newStats != nil {
				var err error
				if stats.DownloadTime, err = plusDuration(stats.DownloadTime, newStats.DownloadTime); err != nil {
					logger.Logf("error adding durations: %+v", err)
				}
				if stats.ProcessTime, err = plusDuration(stats.ProcessTime, newStats.ProcessTime); err != nil {
					logger.Logf("error adding durations: %+v", err)
				}
				if stats.UploadTime, err = plusDuration(stats.UploadTime, newStats.UploadTime); err != nil {
					logger.Logf("error adding durations: %+v", err)
				}
				stats.DownloadBytes += newStats.DownloadBytes
				stats.UploadBytes += newStats.UploadBytes
			}
			// so as not to overwhelm etcd we update at most 100 times per job
			if (float64(totalProcessedData-setData)/float64(totalData)) > .01 ||
				totalProcessedData == 0 || totalProcessedData == totalData {
				// we setProcessedData even though the update below may fail,
				// if we didn't we'd retry updating the progress on the next
				// datum, this would lead to more accurate progress but
				// progress isn't that important and we don't want to overwelm
				// etcd.
				setData = totalProcessedData
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobInfo := new(pps.JobInfo)
					if err := jobs.Get(jobID, jobInfo); err != nil {
						return err
					}
					jobInfo.DataProcessed = processedData
					jobInfo.DataSkipped = skippedData
					jobInfo.DataTotal = totalData
					jobInfo.Stats = stats
					jobs.Put(jobInfo.Job.ID, jobInfo)
					return nil
				}); err != nil {
					logger.Logf("error updating job progress: %+v", err)
				}
			}
		}
		// set the initial values
		updateProgress(0, 0, nil)

		for i := 0; i < df.Len(); i++ {
			limiter.Acquire()
			files := df.Datum(i)
			var parentOutputTag *pfs.Tag
			if newBranchParentCommit != nil {
				var parentFiles []*Input
				for _, file := range files {
					parentFile := proto.Clone(file).(*Input)
					if file.FileInfo.File.Commit.Repo.Name == jobInfo.NewBranch.Head.Repo.Name && file.Branch == jobInfo.NewBranch.Name {
						parentFileInfo, err := pfsClient.InspectFile(ctx, &pfs.InspectFileRequest{
							File: client.NewFile(parentFile.FileInfo.File.Commit.Repo.Name, newBranchParentCommit.ID, parentFile.FileInfo.File.Path),
						})
						if err != nil {
							if !isNotFoundErr(err) {
								return err
							}
							// we didn't find a match for this file,
							// so we know there's no matching datum
							break
						}
						file.ParentCommit = parentFileInfo.File.Commit
						parentFile.FileInfo = parentFileInfo
					}
					parentFiles = append(parentFiles, parentFile)
				}
				if len(parentFiles) == len(files) {
					_parentOutputTag, err := HashDatum(pipelineInfo, parentFiles)
					if err != nil {
						return err
					}
					parentOutputTag = &pfs.Tag{Name: _parentOutputTag}
				}
			}
			go func() {
				userCodeFailures := 0
				defer limiter.Release()
				b := backoff.NewInfiniteBackOff()
				b.Multiplier = 1
				var resp *ProcessResponse
				datumProcessStats := &pps.ProcessStats{}
				if err := backoff.RetryNotify(func() error {
					if err := pool.Do(ctx, func(conn *grpc.ClientConn) error {
						workerClient := NewWorkerClient(conn)
						resp, err = workerClient.Process(ctx, &ProcessRequest{
							JobID:        jobInfo.Job.ID,
							Data:         files,
							ParentOutput: parentOutputTag,
							EnableStats:  jobInfo.EnableStats,
						})
						return err
					}); err != nil {
						return fmt.Errorf("Process() call failed: %v", err)
					}
					if resp.Failed {
						userCodeFailures++
						// If this is our last failure we merge in the stats
						// tree for the failed run.
						if userCodeFailures > maximumRetriesPerDatum && jobInfo.EnableStats {
							statsSubtree, err := a.getTreeFromTag(ctx, resp.StatsTag)
							if err != nil {
								logger.Logf("failed to retrieve stats hashtree after processing for datum %v: %v", files, err)
							} else {
								treeMu.Lock()
								defer treeMu.Unlock()
								if err := statsTree.Merge(statsSubtree); err != nil {
									logger.Logf("failed to merge into stats tree: %v", err)
								}
							}
						}
						return fmt.Errorf("user code failed for datum %v", files)
					}
					var eg errgroup.Group
					var subTree hashtree.HashTree
					var statsSubtree hashtree.HashTree
					eg.Go(func() error {
						subTree, err = a.getTreeFromTag(ctx, resp.Tag)
						if err != nil {
							return fmt.Errorf("failed to retrieve hashtree after processing for datum %v: %v", files, err)
						}
						return nil
					})
					if jobInfo.EnableStats {
						eg.Go(func() error {
							statsSubtree, err = a.getTreeFromTag(ctx, resp.StatsTag)
							if err != nil {
								logger.Logf("failed to retrieve stats hashtree after processing for datum %v: %v", files, err)
								return nil
							}
							nodes, err := statsSubtree.Glob("*/*/stats")
							if err != nil {
								logger.Logf("failed to retrieve process stats from hashtree for datum %v: %v", files, err)
							}
							if len(nodes) != 1 {
								logger.Logf("should have a single stats object for datum %v", files)
								return nil
							}
							var objects []string
							for _, object := range nodes[0].FileNode.Objects {
								objects = append(objects, object.Hash)
							}
							var buffer bytes.Buffer
							a.pachClient.WithCtx(ctx).GetObjects(objects, 0, 0, &buffer)
							if err := jsonpb.UnmarshalString(buffer.String(), datumProcessStats); err != nil {
								logger.Logf("error unmarshalling datum process stats for datum %v: %+v", files, err)
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
						processStats = append(processStats, datumProcessStats)
					}
					return tree.Merge(subTree)
				}, b, func(err error, d time.Duration) error {
					select {
					case <-ctx.Done():
						return err
					default:
					}
					if userCodeFailures > maximumRetriesPerDatum {
						logger.Logf("job %s failed to process datum %+v %d times failing", jobID, files, userCodeFailures)
						failed = true
						return err
					}
					logger.Logf("job %s failed to process datum %+v with: %+v, retrying in: %+v", jobID, files, err, d)
					return nil
				}); err == nil {
					if resp.Skipped {
						go updateProgress(0, 1, datumProcessStats)
					} else {
						go updateProgress(1, 0, datumProcessStats)
					}
				}
			}()
		}
		limiter.Wait()

		var statsCommit *pfs.Commit
		if jobInfo.EnableStats {
			if err := func() error {
				aggregateProcessStats, err := a.aggregateProcessStats(processStats)
				if err != nil {
					return err
				}
				marshalled, err := (&jsonpb.Marshaler{}).MarshalToString(aggregateProcessStats)
				if err != nil {
					return err
				}
				aggregateObject, _, err := a.pachClient.WithCtx(ctx).PutObject(strings.NewReader(marshalled))
				if err != nil {
					return err
				}
				return statsTree.PutFile(path.Join("/", jobID, "stats"), []*pfs.Object{aggregateObject}, int64(len(marshalled)))
			}(); err != nil {
				logger.Logf("error aggregating stats")
			}
			statsObject, err := a.putTree(ctx, statsTree)
			if err != nil {
				return err
			}

			statsCommit, err = pfsClient.BuildCommit(ctx, &pfs.BuildCommitRequest{
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

		// check if the job failed
		if failed {
			_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				jobs := a.jobs.ReadWrite(stm)
				jobInfo := new(pps.JobInfo)
				if err := jobs.Get(jobID, jobInfo); err != nil {
					return err
				}
				jobInfo.Finished = now()
				jobInfo.StatsCommit = statsCommit
				return a.updateJobState(stm, jobInfo, pps.JobState_JOB_FAILURE)
			})
			return err
		}

		object, err := a.putTree(ctx, tree)
		if err != nil {
			return err
		}

		var provenance []*pfs.Commit
		for _, commit := range pps.InputCommits(jobInfo.Input) {
			provenance = append(provenance, commit)
		}

		outputCommit, err := pfsClient.BuildCommit(ctx, &pfs.BuildCommitRequest{
			Parent: &pfs.Commit{
				Repo: jobInfo.OutputRepo,
			},
			Branch:     jobInfo.OutputBranch,
			Provenance: provenance,
			Tree:       object,
		})
		if err != nil {
			return err
		}

		if jobInfo.Egress != nil {
			logger.Logf("Starting egress upload for job (%v)\n", jobInfo)
			start := time.Now()
			objClient, err := obj.NewClientFromURLAndSecret(ctx, jobInfo.Egress.URL)
			if err != nil {
				return err
			}
			url, err := url.Parse(jobInfo.Egress.URL)
			if err != nil {
				return err
			}
			client := client.APIClient{
				PfsAPIClient: pfsClient,
			}
			client.SetMaxConcurrentStreams(100)
			if err := pfs_sync.PushObj(client, outputCommit, objClient, strings.TrimPrefix(url.Path, "/")); err != nil {
				return err
			}
			logger.Logf("Completed egress upload for job (%v), duration (%v)\n", jobInfo, time.Since(start))
		}

		// Record the job's output commit and 'Finished' timestamp, and mark the job
		// as a SUCCESS
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobInfo := new(pps.JobInfo)
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			jobInfo.OutputCommit = outputCommit
			jobInfo.Finished = now()
			// By definition, we will have processed all datums at this point
			jobInfo.DataProcessed = processedData
			jobInfo.DataSkipped = skippedData
			jobInfo.Stats = stats
			// likely already set but just in case it failed
			jobInfo.DataTotal = totalData
			jobInfo.StatsCommit = statsCommit
			return a.updateJobState(stm, jobInfo, pps.JobState_JOB_SUCCESS)
		})
		return err
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			// Exit the retry loop if context got cancelled
			return err
		default:
		}

		jobStoppedMutex.Lock()
		defer jobStoppedMutex.Unlock()
		if jobStopped {
			// If the job has been stopped, exit the retry loop
			return err
		}

		logger.Logf("error running jobManager for job %s: %v; retrying in %v", jobInfo.Job.ID, err, d)

		// Increment the job's restart count
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobInfo := new(pps.JobInfo)
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			jobInfo.Restart++
			jobs.Put(jobInfo.Job.ID, jobInfo)
			return nil
		})
		if err != nil {
			logger.Logf("error incrementing job %s's restart count", jobInfo.Job.ID)
		}

		return nil
	})
	return nil
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
		logger.Logf("error aggregating mean: %v\n", err)
	}
	stddev, err := stats.StandardDeviation(datums)
	if err != nil {
		logger.Logf("error aggregating std dev: %v\n", err)
	}
	fifth, err := stats.Percentile(datums, 5)
	if err != nil {
		logger.Logf("error aggregating 5th percentile: %v\n", err)
	}
	ninetyFifth, err := stats.Percentile(datums, 95)
	if err != nil {
		logger.Logf("error aggregating 95th percentile: %v\n", err)
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
	rc := a.kubeClient.ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(ppsserver.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version))
	if err != nil {
		return err
	}
	workerRc.Spec.Replicas = 1
	// When we scale down the workers, we also remove the resource
	// requirements so that the remaining master pod does not take up
	// the resource it doesn't need, since by definition when a pipeline
	// is in scale-down mode, it doesn't process any work.
	if a.pipelineInfo.ResourceSpec != nil {
		workerRc.Spec.Template.Spec.Containers[0].Resources = api.ResourceRequirements{}
	}
	_, err = rc.Update(workerRc)
	return err
}

func (a *APIServer) scaleUpWorkers() error {
	rc := a.kubeClient.ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(ppsserver.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version))
	if err != nil {
		return err
	}
	parallelism, err := ppsserver.GetExpectedNumWorkers(a.kubeClient, a.pipelineInfo.ParallelismSpec)
	if err != nil {
		return err
	}
	if workerRc.Spec.Replicas != int32(parallelism) {
		workerRc.Spec.Replicas = int32(parallelism)
	}
	// Reset the resource requirements for the RC since the pipeline
	// is in scale-down mode and probably has removed its resource
	// requirements.
	if a.pipelineInfo.ResourceSpec != nil {
		resourceList, err := util.GetResourceListFromPipeline(a.pipelineInfo)
		if err != nil {
			return fmt.Errorf("error parsing resource spec; this is likely a bug: %v", err)
		}
		workerRc.Spec.Template.Spec.Containers[0].Resources = api.ResourceRequirements{
			Requests: *resourceList,
		}
	}
	_, err = rc.Update(workerRc)
	return err
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

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}

package worker

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"go.pedge.io/lion/proto"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	pfs_sync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

const (
	// maximumRetriesPerDatum is the maximum number of times each datum
	// can failed to be processed before we declare that the job has failed.
	maximumRetriesPerDatum = 3

	masterLockPath = "_master_worker_lock"
)

func (a *APIServer) master() {
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		masterLock, err := dlock.NewDLock(ctx, a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.ID))
		if err != nil {
			return err
		}
		defer masterLock.Unlock()
		ctx = masterLock.Context()

		protolion.Infof("Launching worker master process")

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

		return a.jobSpawner(ctx)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		protolion.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

// jobSpawner spawns jobs
func (a *APIServer) jobSpawner(ctx context.Context) error {
	// Establish connection pool
	numWorkers, err := pps.GetExpectedNumWorkers(a.kubeClient, a.pipelineInfo.ParallelismSpec)
	if err != nil {
		return err
	}
	pool, err := grpcutil.NewPool(a.kubeClient, a.namespace, pps.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version), numWorkers, client.PachDialOptions()...)
	if err != nil {
		return fmt.Errorf("master: error constructing worker pool: %v; retrying in %v", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			protolion.Errorf("error closing pool: %v", err)
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
				protolion.Errorf("error converting scaleDownThreshold: %v", err)
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
				protolion.Errorf("error scaling down workers: %v", err)
			}
			continue nextInput
		}

		// Once we received a job, scale up the workers
		if a.pipelineInfo.ScaleDownThreshold != nil {
			if err := a.scaleUpWorkers(); err != nil {
				protolion.Errorf("error scaling up workers: %v", err)
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
			if jobInfo.PipelineID == a.pipelineInfo.ID && jobInfo.PipelineVersion == a.pipelineInfo.Version {
				switch jobInfo.State {
				case pps.JobState_JOB_STARTING, pps.JobState_JOB_RUNNING:
					if err := a.runJob(ctx, &jobInfo, pool); err != nil {
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
				if jobInfo.PipelineID == a.pipelineInfo.ID && jobInfo.PipelineVersion == a.pipelineInfo.Version {
					parentJob = jobInfo.Job
				}
			}
		}

		job, err := a.pachClient.PpsAPIClient.CreateJob(ctx, &pps.CreateJobRequest{
			Pipeline: a.pipelineInfo.Pipeline,
			Input:    jobInput,
			// TODO(derek): Note that once the pipeline restarts, the `job`
			// variable is lost and we don't know who is our parent job.
			ParentJob: parentJob,
			NewBranch: newBranch,
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

		if err := a.runJob(ctx, jobInfo, pool); err != nil {
			return err
		}
	}
}

// jobManager feeds datums to jobs
func (a *APIServer) runJob(ctx context.Context, jobInfo *pps.JobInfo, pool *grpcutil.Pool) error {
	pfsClient := a.pachClient.PfsAPIClient
	objectClient := a.pachClient.ObjectAPIClient
	ppsClient := a.pachClient.PpsAPIClient

	jobID := jobInfo.Job.ID
	var jobStopped bool
	var jobStoppedMutex sync.Mutex
	backoff.RetryNotify(func() error {
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
				protolion.Errorf("error monitoring job state: %v", err)
				return
			}
			switch currentJobInfo.State {
			case pps.JobState_JOB_STOPPED, pps.JobState_JOB_SUCCESS, pps.JobState_JOB_FAILURE:
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
		limiter := limit.New(a.numWorkers)
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
		var treeMu sync.Mutex

		processedData := int64(0)
		setProcessedData := int64(0)
		totalData := int64(df.Len())
		var progressMu sync.Mutex
		updateProgress := func(processed int64) {
			progressMu.Lock()
			defer progressMu.Unlock()
			processedData += processed
			// so as not to overwhelm etcd we update at most 100 times per job
			if (float64(processedData-setProcessedData)/float64(totalData)) > .01 ||
				processedData == 0 || processedData == totalData {
				// we setProcessedData even though the update below may fail,
				// if we didn't we'd retry updating the progress on the next
				// datum, this would lead to more accurate progress but
				// progress isn't that important and we don't want to overwelm
				// etcd.
				setProcessedData = processedData
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobInfo := new(pps.JobInfo)
					if err := jobs.Get(jobID, jobInfo); err != nil {
						return err
					}
					jobInfo.DataProcessed = processedData
					jobInfo.DataTotal = totalData
					jobs.Put(jobInfo.Job.ID, jobInfo)
					return nil
				}); err != nil {
					protolion.Errorf("error updating job progress: %+v", err)
				}
			}
		}
		// set the initial values
		updateProgress(0)

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
				if err := backoff.RetryNotify(func() error {
					conn, err := pool.Get(ctx)
					if err != nil {
						return fmt.Errorf("error from connection pool: %v", err)
					}
					workerClient := NewWorkerClient(conn)
					resp, err := workerClient.Process(ctx, &ProcessRequest{
						JobID:        jobInfo.Job.ID,
						Data:         files,
						ParentOutput: parentOutputTag,
					})
					if err != nil {
						if err := conn.Close(); err != nil {
							protolion.Errorf("error closing conn: %+v", err)
						}
						return fmt.Errorf("Process() call failed: %v", err)
					}
					defer func() {
						if err := pool.Put(conn); err != nil {
							protolion.Errorf("error Putting conn: %+v", err)
						}
					}()
					if resp.Failed {
						userCodeFailures++
						return fmt.Errorf("user code failed for datum %v", files)
					}
					getTagClient, err := objectClient.GetTag(ctx, resp.Tag)
					if err != nil {
						return fmt.Errorf("failed to retrieve hashtree after processing for datum %v: %v", files, err)
					}
					var buffer bytes.Buffer
					if err := grpcutil.WriteFromStreamingBytesClient(getTagClient, &buffer); err != nil {
						return fmt.Errorf("failed to retrieve hashtree after processing for datum %v: %v", files, err)
					}
					subTree, err := hashtree.Deserialize(buffer.Bytes())
					if err != nil {
						return fmt.Errorf("failed deserialize hashtree after processing for datum %v: %v", files, err)
					}
					treeMu.Lock()
					defer treeMu.Unlock()
					return tree.Merge(subTree)
				}, b, func(err error, d time.Duration) error {
					select {
					case <-ctx.Done():
						return err
					default:
					}
					if userCodeFailures > maximumRetriesPerDatum {
						protolion.Errorf("job %s failed to process datum %+v %d times failing", jobID, files, userCodeFailures)
						failed = true
						return err
					}
					protolion.Errorf("job %s failed to process datum %+v with: %+v, retrying in: %+v", jobID, files, err, d)
					return nil
				}); err == nil {
					go updateProgress(1)
				}
			}()
		}
		limiter.Wait()

		// check if the job failed
		if failed {
			_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				jobs := a.jobs.ReadWrite(stm)
				jobInfo := new(pps.JobInfo)
				if err := jobs.Get(jobID, jobInfo); err != nil {
					return err
				}
				jobInfo.Finished = now()
				return a.updateJobState(stm, jobInfo, pps.JobState_JOB_FAILURE)
			})
			return err
		}

		finishedTree, err := tree.Finish()
		if err != nil {
			return err
		}

		data, err := hashtree.Serialize(finishedTree)
		if err != nil {
			return err
		}

		putObjClient, err := objectClient.PutObject(ctx)
		if err != nil {
			return err
		}
		for _, chunk := range grpcutil.Chunk(data, grpcutil.MaxMsgSize/2) {
			if err := putObjClient.Send(&pfs.PutObjectRequest{
				Value: chunk,
			}); err != nil {
				return err
			}
		}
		object, err := putObjClient.CloseAndRecv()
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
			jobInfo.DataProcessed = totalData
			// likely already set but just in case it failed
			jobInfo.DataTotal = totalData
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

		protolion.Errorf("error running jobManager for job %s: %v; retrying in %v", jobInfo.Job.ID, err, d)

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
			protolion.Errorf("error incrementing job %s's restart count", jobInfo.Job.ID)
		}

		return nil
	})
	return nil
}

func (a *APIServer) scaleDownWorkers() error {
	rc := a.kubeClient.ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(pps.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version))
	if err != nil {
		return err
	}
	if workerRc.Spec.Replicas != 1 {
		workerRc.Spec.Replicas = 1
		_, err = rc.Update(workerRc)
		return err
	}
	return nil
}

func (a *APIServer) scaleUpWorkers() error {
	rc := a.kubeClient.ReplicationControllers(a.namespace)
	workerRc, err := rc.Get(pps.PipelineRcName(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Version))
	if err != nil {
		return err
	}
	parallelism, err := pps.GetExpectedNumWorkers(a.kubeClient, a.pipelineInfo.ParallelismSpec)
	if err != nil {
		return err
	}
	if workerRc.Spec.Replicas != int32(parallelism) {
		workerRc.Spec.Replicas = int32(parallelism)
		_, err = rc.Update(workerRc)
		return err
	}
	return nil
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

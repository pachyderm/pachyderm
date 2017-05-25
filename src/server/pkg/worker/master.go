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

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/input"
	"github.com/pachyderm/pachyderm/src/client/pkg/limit"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	masterLocksDir = "_master_locks"
)

func (a *APIServer) master() {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Acquire the master lock
		masterLock, err := dlock.NewDLock(ctx, a.etcdClient, path.Join(a.ppsEtcdPrefix, masterLocksDir, a.pipelineInfo.Pipeline.Name))
		if err != nil {
			return err
		}
		defer masterLock.Unlock()
		ctx = masterLock.Context()

		var provenance []*pfs.Repo
		for _, commit := range input.InputCommits(a.pipelineInfo.Input) {
			provenance = append(provenance, commit.Repo)
		}

		pfsClient := a.pachClient.PfsAPIClient
		ppsClient := a.pachClient.PpsAPIClient

		// Create the output repo; if it already exists, do nothing
		if _, err := pfsClient.CreateRepo(ctx, &pfs.CreateRepoRequest{
			Repo:       &pfs.Repo{a.pipelineInfo.Pipeline.Name},
			Provenance: provenance,
		}); err != nil {
			if !isAlreadyExistsErr(err) {
				return err
			}
		}

		branchSetFactory, err := newBranchSetFactory(ctx, pfsClient, a.pipelineInfo.Input)
		if err != nil {
			return err
		}
		defer branchSetFactory.Close()

		var job *pps.Job
	nextInput:
		for {
			branchSet := <-branchSetFactory.Chan()
			if branchSet.Err != nil {
				return err
			}

			// (create JobInput for new processing job)
			jobInput := proto.Clone(a.pipelineInfo.Input).(*pps.Input)
			var visitErr error
			input.Visit(jobInput, func(input *pps.Input) {
				if input.Atom != nil {
					for _, branch := range branchSet.Branches {
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
			jobIter, err := jobsRO.GetByIndex(jobsInputIndex, jobInput)
			if err != nil {
				return err
			}

			// Check if there's already a job with this input
			var jobInfo *pps.JobInfo
			for {
				var jobID string
				var oldJobInfo pps.JobInfo
				ok, err := jobIter.Next(&jobID, &oldJobInfo)
				if err != nil {
					return err
				}
				if !ok {
					break
				}
				if oldJobInfo.PipelineID == pipelineInfo.ID && oldJobInfo.PipelineVersion == pipelineInfo.Version {
					job = oldJobInfo.Job
					switch oldJobInfo.State {
					case pps.JobState_JOB_SUCCESS, pps.JobState_JOB_FAILURE, pps.JobState_JOB_STOPPED:
						continue nextInput
					}
					jobInfo = &oldJobInfo
					break
				}
			}

			if jobInfo == nil {
				job, err = a.CreateJob(ctx, &pps.CreateJobRequest{
					Pipeline:  pipelineInfo.Pipeline,
					Input:     jobInput,
					ParentJob: job,
				})
				if err != nil {
					return err
				}

				jobInfo, err = ppsClient.InspectJob(ctx, job)
				if err != nil {
					return err
				}
			} else {
				job = jobInfo.Job
			}

			if err := a.runJob(ctx, jobInfo); err != nil {
				return err
			}
		}
	}, b, func(err error, d time.Duration) error {
		protolion.Infof("master process failed; retrying in %s", d)
		return nil
	})
}

// runJob feeds datums to workers for a job
func (a *APIServer) runJob(ctx context.Context, jobInfo *pps.JobInfo) error {
	jobID := jobInfo.Job.ID
	b := backoff.NewInfiniteBackOff()
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
			if _, err := a.InspectJob(ctx, &pps.InspectJobRequest{
				Job:        jobInfo.ParentJob,
				BlockState: true,
			}); err != nil {
				return err
			}
		}

		pfsClient := a.pachClient.PfsAPIClient
		objectClient := a.pachClient.ObjectAPIClient

		// Create workers and output repo if 'jobInfo' belongs to an orphan job
		if jobInfo.Pipeline == nil {
			// Create output repo for this job
			var provenance []*pfs.Repo
			input.Visit(jobInfo.Input, func(input *pps.Input) {
				if input.Atom != nil {
					provenance = append(provenance, client.NewRepo(input.Atom.Repo))
				}
			})
			if _, err := pfsClient.CreateRepo(ctx, &pfs.CreateRepoRequest{
				Repo:       jobInfo.OutputRepo,
				Provenance: provenance,
			}); err != nil {
				// (if output repo already exists, do nothing)
				if !isAlreadyExistsErr(err) {
					return err
				}
			}

			// Create workers in kubernetes (i.e. create replication controller)
			if err := a.createWorkersForOrphanJob(jobInfo); err != nil {
				if !isAlreadyExistsErr(err) {
					return err
				}
			}
			go func() {
				// Clean up workers if the job gets cancelled
				<-ctx.Done()
				rcName := JobRcName(jobInfo.Job.ID)
				if err := a.deleteWorkers(rcName); err != nil {
					protolion.Errorf("error deleting workers for job: %v", jobID)
				}
				protolion.Infof("deleted workers for job: %v", jobID)
			}()
		}

		// Set the state of this job to 'RUNNING'
		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
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

		// Start worker pool
		var rcName string
		if jobInfo.Pipeline != nil {
			// We scale up the workers before we run a job, to ensure
			// that the job will have workers to use.  Note that scaling
			// a RC is idempotent: nothing happens if the workers have
			// already been scaled.
			rcName = PipelineRcName(jobInfo.Pipeline.Name, jobInfo.PipelineVersion)
			if err := a.scaleUpWorkers(ctx, rcName, jobInfo.ParallelismSpec); err != nil {
				return err
			}
		} else {
			rcName = JobRcName(jobInfo.Job.ID)
		}

		failed := false
		numWorkers, err := a.numWorkers(ctx, rcName)
		if err != nil {
			return err
		}
		limiter := limit.New(numWorkers)
		// process all datums
		df, err := newDatumFactory(ctx, pfsClient, jobInfo.Input)
		if err != nil {
			return err
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

		serviceAddr, err := a.workerServiceIP(ctx, rcName)
		if err != nil {
			return err
		}
		pool := grpcutil.NewPool(fmt.Sprintf("%s:%d", serviceAddr, client.PPSWorkerPort), numWorkers, client.PachDialOptions()...)
		defer func() {
			if err := pool.Close(); err != nil {
				protolion.Errorf("error closing pool: %+v", pool)
			}
		}()
		for i := 0; i < df.Len(); i++ {
			limiter.Acquire()
			files := df.Datum(i)
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
					workerClient := workerpkg.NewWorkerClient(conn)
					resp, err := workerClient.Process(ctx, &workerpkg.ProcessRequest{
						JobID: jobInfo.Job.ID,
						Data:  files,
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
					if userCodeFailures > MaximumRetriesPerDatum {
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
		for _, commit := range input.InputCommits(jobInfo.Input) {
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
	}, b, func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			// Exit the retry loop if context got cancelled
			return err
		default:
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
}

func isAlreadyExistsErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

package transform

import (
	"fmt"
	"io"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func jobDatumHashtreeTag(jobID string) *pfs.Tag {
	return client.NewTag(fmt.Sprintf("job-%s-datum", jobID))
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

// mergeStats merges y into x
func mergeStats(x, y *pps.ProcessStats) error {
	var err error
	if x.DownloadTime, err = plusDuration(x.DownloadTime, y.DownloadTime); err != nil {
		return err
	}
	if x.ProcessTime, err = plusDuration(x.ProcessTime, y.ProcessTime); err != nil {
		return err
	}
	if x.UploadTime, err = plusDuration(x.UploadTime, y.UploadTime); err != nil {
		return err
	}
	x.DownloadBytes += y.DownloadBytes
	x.UploadBytes += y.UploadBytes
	return nil
}

type processResult struct {
	failedDatumID   string
	datumsProcessed int64
	datumsSkipped   int64
	datumsRecovered int64
	datumsFailed    int64
	recoveredDatums *pfs.Object
}

type processFunc func(low, high int64) (*processResult, error)

// Worker does the following:
//  - claims filesystem shards as they become available
//  - watches for new jobs (jobInfos in the jobs collection)
//  - claims chunks from the chunk layout it finds in the chunks collection
//  - claims those chunks with acquireDatums
//  - processes the chunks with processDatums
//  - merges the chunks with mergeDatums
func Worker(driver driver.Driver, logger logs.TaggedLogger, task *work.Task, subtask *work.Task) error {
	jobData, err := deserializeJobData(task.Data)
	if err != nil {
		return err
	}

	logger = logger.WithJob(jobData.JobId)

	datumData, err := deserializeDatumData(subtask.Data)
	if err == nil {
		if err = handleDatumTask(driver, logger, datumData); err != nil {
			return err
		}
		subtask.Data, err = serializeDatumData(datumData)
		return err
	}

	mergeData, err := deserializeMergeData(subtask.Data)
	if err == nil {
		if err = handleMergeTask(driver, logger, mergeData); err != nil {
			return err
		}
		subtask.Data, err = serializeMergeData(mergeData)
		return err
	}

	return fmt.Errorf("worker task format unrecognized")
}

func forEachDatum(driver driver.Driver, object *pfs.Object, cb func([]*common.Input) error) (retErr error) {
	getObjectClient, err := driver.PachClient().ObjectAPIClient.GetObject(driver.PachClient().Ctx(), object)
	if err != nil {
		return err
	}

	grpcReader := grpcutil.NewStreamingBytesReader(getObjectClient, nil)
	protoReader := pbutil.NewReader(grpcReader)

	defer func() {
		if retErr == io.EOF {
			retErr = nil
		}
	}()

	datum := &DatumInputs{}

	if err := protoReader.Read(datum); err != nil {
		return err
	}

	for {
		cb(datum.Inputs)

		if err := protoReader.Read(datum); err != nil {
			return err
		}
	}
}

func handleDatumTask(driver driver.Driver, logger logs.TaggedLogger, data *DatumData) error {
	logger.Logf("transform worker datum task: %v", data)
	data.ProcessStats = &pps.ProcessStats{}
	/*
		  pps.ProcessStats process_stats = 2;
			int64 datums_processed = 4;
			int64 datums_skipped = 5;
			int64 datums_recovered = 6;
			int64 datums_failed = 7;
			pfs.Object recovered_datums = 8;
			string failed_datum_id = 3;
	*/
	limiter := limit.New(int(driver.PipelineInfo().MaxQueueSize))

	runMutex := &sync.Mutex{}
	var eg errgroup.Group
	if err := forEachDatum(driver, data.Datums, func(inputs []*common.Input) error {
		limiter.Acquire()
		eg.Go(func() {
			defer limiter.Release()
			return processDatum(driver, logger.WithData(inputs), inputs, runMutex)
		})
		return nil
	}); err != nil {
		return err
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func processDatum(driver driver.Driver, logger logs.TaggedLogger, inputs []*common.Input, runMutex *sync.Mutex) error {
	if err := backoff.RetryUntilCancel(func() error {
		subStats, err := driver.WithData(inputs, inputTree, logger, func(subStats *pps.ProcessStats) error {
			env := userCodeEnv(jobInfo.Job.ID, jobInfo.OutputCommit.ID, localDriver.InputDir(), data)
			runMutex.Lock()
			defer runMutex.Unlock()
			if err := driver.RunUserCode(logger, env, subStats, jobInfo.DatumTimeout); err != nil {
				if a.pipelineInfo.Transform.ErrCmd != nil && failures == jobInfo.DatumTries-1 {
					if err = localDriver.RunUserErrorHandlingCode(logger, env, subStats, jobInfo.DatumTimeout); err != nil {
						return fmt.Errorf("error runUserErrorHandlingCode: %v", err)
					}
					return errDatumRecovered
				}
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		b, err := driver.UploadOutput(tag, logger, data, subStats, outputTree)
		if err != nil {
			return err
		}
		// Cache datum hashtree locally
		return a.datumCache.Put(datumIdx, bytes.NewReader(b))
	}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
		failures++
		if failures >= jobInfo.DatumTries {
			logger.Logf("failed to process datum with error: %+v", err)
			if statsTree != nil {
				object, size, err := pachClient.PutObject(strings.NewReader(err.Error()))
				if err != nil {
					logger.Errf("could not put error object: %s\n", err)
				} else {
					objectInfo, err := pachClient.InspectObject(object.Hash)
					if err != nil {
						return err
					}
					h, err := pfs.DecodeHash(object.Hash)
					if err != nil {
						return err
					}
					statsTree.PutFile("failure", h, size, objectInfo.BlockRef)
				}
			}
			return err
		}
		logger.Logf("failed processing datum: %v, retrying in %v", err, d)
		return nil
	}); err == errDatumRecovered {
		// keep track of the recovered datums
		recoverMu.Lock()
		defer recoverMu.Unlock()
		recoveredDatums = append(recoveredDatums, a.DatumID(data))
		atomic.AddInt64(&result.datumsRecovered, 1)
		return nil
	} else if err != nil {
		result.failedDatumID = common.DatumID(data)
		atomic.AddInt64(&result.datumsFailed, 1)
		return nil
	}
	return nil
}

func handleMergeTask(driver driver.Driver, logger logs.TaggedLogger, data *MergeData) error {
	logger.Logf("transform worker merge task: %v", data)
	return nil
}

/*
func (a *APIServer) Worker() {
	logger := logs.NewStatlessLogger(a.pipelineInfo)

	// claim a shard if one is available or becomes available
	go a.claimShard(a.pachClient.Ctx())

	// Process incoming jobs
	backoff.RetryNotify(func() (retErr error) {
		retryCtx, retryCancel := context.WithCancel(a.pachClient.Ctx())
		defer retryCancel()
		watcher, err := a.driver.Jobs().ReadOnly(retryCtx).WatchByIndex(ppsdb.JobsPipelineIndex, a.pipelineInfo.Pipeline)
		if err != nil {
			return fmt.Errorf("error creating watch: %v", err)
		}
		defer watcher.Close()
		for e := range watcher.Watch() {
			// Clear chunk caches from previous job
			if err := a.chunkCache.Clear(); err != nil {
				logger.Logf("error clearing chunk cache: %v", err)
			}
			if err := a.chunkStatsCache.Clear(); err != nil {
				logger.Logf("error clearing chunk stats cache: %v", err)
			}
			if e.Type == watch.EventError {
				return fmt.Errorf("worker watch error: %v", e.Err)
			} else if e.Type == watch.EventDelete {
				// Job was deleted, e.g. because input commit was deleted. This is
				// handled by cancelCtxIfJobFails goro, which was spawned when job was
				// created. Nothing to do here
				continue
			}

			// 'e' is a Put event -- new job
			var jobID string
			jobPtr := &pps.EtcdJobInfo{}
			if err := e.Unmarshal(&jobID, jobPtr); err != nil {
				return fmt.Errorf("error unmarshalling: %v", err)
			}
			if ppsutil.IsTerminal(jobPtr.State) {
				// previously-created job has finished, or job was finished during backoff
				// or in the 'watcher' queue
				logger.Logf("skipping job %v as it is already in state %v", jobID, jobPtr.State)
				continue
			}

			// create new ctx for this job, and don't use retryCtx as the
			// parent. Just because another job's etcd write failed doesn't
			// mean this job shouldn't run
			if err := func() error {
				jobCtx, jobCancel := context.WithCancel(a.pachClient.Ctx())
				defer jobCancel() // cancel the job ctx
				pachClient := a.pachClient.WithCtx(jobCtx)

				//  Watch for any changes to EtcdJobInfo corresponding to jobID; if
				// the EtcdJobInfo is marked 'FAILED', call jobCancel().
				// ('watcher' above can't detect job state changes--it's watching
				// an index and so only emits when jobs are created or deleted).
				go a.cancelCtxIfJobFails(jobCtx, jobCancel, jobID)

				// Inspect the job and make sure it's relevant, as this worker may be old
				jobInfo, err := pachClient.InspectJob(jobID, false)
				if err != nil {
					if col.IsErrNotFound(err) {
						return nil
					}
					return fmt.Errorf("error from InspectJob(%v): %+v", jobID, err)
				}
				if jobInfo.PipelineVersion < a.pipelineInfo.Version {
					logger.Logf("skipping job %v as it uses old pipeline version %d", jobID, jobInfo.PipelineVersion)
					return nil
				}
				if jobInfo.PipelineVersion > a.pipelineInfo.Version {
					return fmt.Errorf("job %s's version (%d) greater than pipeline's "+
						"version (%d), this should automatically resolve when the worker "+
						"is updated", jobID, jobInfo.PipelineVersion, a.pipelineInfo.Version)
				}
				logger.Logf("processing job %v", jobID)

				// Read the chunks laid out by the master and create the datum factory
				plan := &common.Plan{}
				if err := a.driver.Plans().ReadOnly(jobCtx).GetBlock(jobInfo.Job.ID, plan); err != nil {
					return fmt.Errorf("error reading job chunks: %v", err)
				}
				dit, err := datum.NewIterator(pachClient, jobInfo.Input)
				if err != nil {
					return fmt.Errorf("error from datum.NewIterator: %v", err)
				}

				// Compute the datums to skip
				skip := make(map[string]bool)
				var useParentHashTree bool
				parentCommitInfo, err := a.getParentCommitInfo(jobCtx, pachClient, jobInfo.OutputCommit)
				if err != nil {
					return err
				}
				if parentCommitInfo != nil {
					var err error
					skip, err = a.getDatumMap(jobCtx, pachClient, parentCommitInfo.Datums)
					if err != nil {
						return err
					}
					var count int
					for i := 0; i < dit.Len(); i++ {
						files := dit.DatumN(i)
						datumHash := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, files)
						if skip[datumHash] {
							count++
						}
					}
					if len(skip) == count {
						useParentHashTree = true
					}
				}

				// Get updated job info from master
				jobInfo, err = pachClient.InspectJob(jobID, false)
				if err != nil {
					return err
				}
				eg, ctx := errgroup.WithContext(jobCtx)
				// If a datum fails, acquireDatums updates the relevant lock in
				// etcd, which causes the master to fail the job (which is
				// handled above in the JOB_FAILURE case). There's no need to
				// handle failed datums here, just failed etcd writes.
				eg.Go(func() error {
					return a.acquireDatums(
						ctx, jobID, plan, logger,
						func(low, high int64) (*processResult, error) {
							processResult, err := a.processDatums(pachClient, logger, jobInfo, dit, low, high, skip, useParentHashTree)
							if err != nil {
								return nil, err
							}
							return processResult, nil
						},
					)
				})
				eg.Go(func() error {
					return a.mergeDatums(ctx, pachClient, jobInfo, jobID, plan, logger, dit, skip, useParentHashTree)
				})
				if err := eg.Wait(); err != nil {
					if jobCtx.Err() == context.Canceled {
						return nil
					}
					return fmt.Errorf("acquire/process/merge datums for job %s exited with err: %v", jobID, err)
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return fmt.Errorf("worker: jobs.WatchByIndex(pipeline = %s) closed unexpectedly", a.pipelineInfo.Pipeline.Name)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("worker: watch closed or error running the worker process: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *APIServer) claimShard(ctx context.Context) {
	watcher, err := a.driver.Shards().ReadOnly(ctx).Watch(watch.WithFilterPut())
	if err != nil {
		log.Printf("error creating shard watcher: %v", err)
		return
	}
	for {
		// Attempt to claim a shard
		for shard := int64(0); shard < a.numShards; shard++ {
			var shardInfo common.ShardInfo
			err := a.driver.Shards().Claim(ctx, fmt.Sprint(shard), &shardInfo, func(ctx context.Context) error {
				ctx = context.WithValue(ctx, shardKey, shard)
				a.claimedShard <- ctx
				<-ctx.Done()
				return nil
			})
			if err != nil && err != col.ErrNotClaimed {
				log.Printf("error attempting to claim shard: %v", err)
				return
			}
		}
		// Wait for a deletion event (ttl expired) before attempting to claim a shard again
		select {
		case e := <-watcher.Watch():
			if e.Type == watch.EventError {
				log.Printf("shard watch error: %v", e.Err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// processDatums processes datums from low to high in dit, if a datum fails it
// returns the id of the failed datum it also may return a variety of errors
// such as network errors.
func (a *APIServer) processDatums(
	pachClient *client.APIClient,
	logger logs.TaggedLogger,
	jobInfo *pps.JobInfo,
	dit datum.Iterator,
	low, high int64,
	skip map[string]bool,
	useParentHashTree bool,
) (result *processResult, retErr error) {
	defer func() {
		if err := a.datumCache.Clear(); err != nil && retErr == nil {
			logger.Logf("error clearing datum cache: %v", err)
		}
		if err := a.datumStatsCache.Clear(); err != nil && retErr == nil {
			logger.Logf("error clearing datum stats cache: %v", err)
		}
	}()
	ctx := pachClient.Ctx()
	objClient, err := obj.NewClientFromSecret(a.hashtreeStorage)
	if err != nil {
		return nil, err
	}
	stats := &pps.ProcessStats{}
	var statsMu sync.Mutex
	result = &processResult{}
	var eg errgroup.Group
	limiter := limit.New(int(a.pipelineInfo.MaxQueueSize))
	var recoveredDatums []string
	var recoverMu sync.Mutex
	for i := low; i < high; i++ {
		datumIdx := i

		limiter.Acquire()
		atomic.AddInt64(&a.queueSize, 1)
		eg.Go(func() (retErr error) {
			defer limiter.Release()
			defer atomic.AddInt64(&a.queueSize, -1)

			data := dit.DatumN(int(datumIdx))
			logger, err := logs.NewLogger(a.pipelineInfo, pachClient)
			if err != nil {
				return err
			}
			logger = logger.WithJob(jobInfo.Job.ID).WithData(data)

			// Hash inputs
			tag := common.HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, data)
			if skip[tag] {
				if !useParentHashTree {
					if err := a.cacheHashtree(pachClient, tag, datumIdx); err != nil {
						return err
					}
				}
				atomic.AddInt64(&result.datumsSkipped, 1)
				logger.Logf("skipping datum")
				return nil
			}
			if _, err := pachClient.InspectTag(ctx, client.NewTag(tag)); err == nil {
				if err := a.cacheHashtree(pachClient, tag, datumIdx); err != nil {
					return err
				}
				atomic.AddInt64(&result.datumsSkipped, 1)
				logger.Logf("skipping datum")
				return nil
			}
			var subStats *pps.ProcessStats
			var inputTree, outputTree *hashtree.Ordered
			var statsTree *hashtree.Unordered
			if a.pipelineInfo.EnableStats {
				statsRoot := path.Join("/", common.DatumID(data))
				inputTree = hashtree.NewOrdered(path.Join(statsRoot, "pfs"))
				outputTree = hashtree.NewOrdered(path.Join(statsRoot, "pfs", "out"))
				statsTree = hashtree.NewUnordered(statsRoot)
				// Write job id to stats tree
				statsTree.PutFile(fmt.Sprintf("job:%s", jobInfo.Job.ID), nil, 0)
				// Write index in datum iterator to stats tree
				object, size, err := pachClient.PutObject(strings.NewReader(fmt.Sprint(int(datumIdx))))
				if err != nil {
					return err
				}
				objectInfo, err := pachClient.InspectObject(object.Hash)
				if err != nil {
					return err
				}
				h, err := pfs.DecodeHash(object.Hash)
				if err != nil {
					return err
				}
				statsTree.PutFile("index", h, size, objectInfo.BlockRef)
				defer func() {
					if err := a.writeStats(pachClient, objClient, tag, subStats, logger, inputTree, outputTree, statsTree, datumIdx); err != nil && retErr == nil {
						retErr = err
					}
				}()
			}

			var failures int64
			if err := backoff.RetryNotify(func() error {
				if common.IsDone(ctx) {
					return ctx.Err() // timeout or cancelled job--don't run datum
				}

				// shadow ctx and pachClient for the context of processing this one datum
				ctx, cancel := context.WithCancel(ctx)
				localDriver := driver.WithCtx(ctx)
				pachClient := pachClient.WithCtx(ctx)
				func() {
					a.statusMu.Lock()
					defer a.statusMu.Unlock()
					a.jobID = jobInfo.Job.ID
					a.data = data
					a.started = time.Now()
					a.cancel = cancel
					a.stats = stats
				}()

				subStats, err := localDriver.WithData(data, inputTree, logger, func(subStats *pps.ProcessStats) error {
					env := userCodeEnv(jobInfo.Job.ID, jobInfo.OutputCommit.ID, localDriver.InputDir(), data)
					a.runMu.Lock()
					defer a.runMu.Unlock()
					if err := localDriver.RunUserCode(logger, env, subStats, jobInfo.DatumTimeout); err != nil {
						if a.pipelineInfo.Transform.ErrCmd != nil && failures == jobInfo.DatumTries-1 {
							if err = localDriver.RunUserErrorHandlingCode(logger, env, subStats, jobInfo.DatumTimeout); err != nil {
								return fmt.Errorf("error runUserErrorHandlingCode: %v", err)
							}
							return errDatumRecovered
						}
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}

				b, err := a.driver.UploadOutput(tag, logger, data, subStats, outputTree)
				if err != nil {
					return err
				}
				// Cache datum hashtree locally
				return a.datumCache.Put(datumIdx, bytes.NewReader(b))
			}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
				if common.IsDone(ctx) {
					return ctx.Err() // timeout or cancelled job, err out and don't retry
				}
				failures++
				if failures >= jobInfo.DatumTries {
					logger.Logf("failed to process datum with error: %+v", err)
					if statsTree != nil {
						object, size, err := pachClient.PutObject(strings.NewReader(err.Error()))
						if err != nil {
							logger.Errf("could not put error object: %s\n", err)
						} else {
							objectInfo, err := pachClient.InspectObject(object.Hash)
							if err != nil {
								return err
							}
							h, err := pfs.DecodeHash(object.Hash)
							if err != nil {
								return err
							}
							statsTree.PutFile("failure", h, size, objectInfo.BlockRef)
						}
					}
					return err
				}
				logger.Logf("failed processing datum: %v, retrying in %v", err, d)
				return nil
			}); err == errDatumRecovered {
				// keep track of the recovered datums
				recoverMu.Lock()
				defer recoverMu.Unlock()
				recoveredDatums = append(recoveredDatums, a.DatumID(data))
				atomic.AddInt64(&result.datumsRecovered, 1)
				return nil
			} else if err != nil {
				result.failedDatumID = common.DatumID(data)
				atomic.AddInt64(&result.datumsFailed, 1)
				return nil
			}
			statsMu.Lock()
			defer statsMu.Unlock()
			if err := mergeStats(stats, subStats); err != nil {
				logger.Logf("failed to merge Stats: %v", err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// put the list of recovered datums in a pfs.Object, and save it as part of the result
	if len(recoveredDatums) > 0 {
		recoveredDatumsBuf := &bytes.Buffer{}
		pbw := pbutil.NewWriter(recoveredDatumsBuf)
		for _, datumHash := range recoveredDatums {
			if _, err := pbw.WriteBytes([]byte(datumHash)); err != nil {
				return nil, err
			}
		}
		recoveredDatumsObj, _, err := pachClient.PutObject(recoveredDatumsBuf)
		if err != nil {
			return nil, err
		}

		result.recoveredDatums = recoveredDatumsObj
	}

	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.driver.Jobs().ReadWrite(stm)
		jobID := jobInfo.Job.ID
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Get(jobID, jobPtr); err != nil {
			return err
		}
		if jobPtr.Stats == nil {
			jobPtr.Stats = &pps.ProcessStats{}
		}
		if err := mergeStats(jobPtr.Stats, stats); err != nil {
			logger.Logf("failed to merge Stats: %v", err)
		}
		return jobs.Put(jobID, jobPtr)
	}); err != nil {
		return nil, err
	}
	result.datumsProcessed = high - low - result.datumsSkipped - result.datumsFailed - result.datumsRecovered
	// Merge datum hashtrees into a chunk hashtree, then cache it.
	if err := a.mergeChunk(logger, high, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (a *APIServer) cacheHashtree(pachClient *client.APIClient, tag string, datumIdx int64) (retErr error) {
	buf := &bytes.Buffer{}
	if err := pachClient.GetTag(tag, buf); err != nil {
		return err
	}
	if err := a.datumCache.Put(datumIdx, buf); err != nil {
		return err
	}
	if a.pipelineInfo.EnableStats {
		buf.Reset()
		if err := pachClient.GetTag(tag+statsTagSuffix, buf); err != nil {
			// We are okay with not finding the stats hashtree.
			// This allows users to enable stats on a pipeline
			// with pre-existing jobs.
			return nil
		}
		return a.datumStatsCache.Put(datumIdx, buf)
	}
	return nil
}

func (a *APIServer) writeStats(
	pachClient *client.APIClient,
	objClient obj.Client,
	tag string,
	stats *pps.ProcessStats,
	logger logs.TaggedLogger,
	inputTree *hashtree.Ordered,
	outputTree *hashtree.Ordered,
	statsTree *hashtree.Unordered,
	datumIdx int64,
) (retErr error) {
	// Store stats and add stats file
	marshaler := &jsonpb.Marshaler{}
	statsString, err := marshaler.MarshalToString(stats)
	if err != nil {
		logger.Errf("could not serialize stats: %s\n", err)
		return err
	}
	object, size, err := pachClient.PutObject(strings.NewReader(statsString))
	if err != nil {
		logger.Errf("could not put stats object: %s\n", err)
		return err
	}
	objectInfo, err := pachClient.InspectObject(object.Hash)
	if err != nil {
		return err
	}
	h, err := pfs.DecodeHash(object.Hash)
	if err != nil {
		return err
	}
	statsTree.PutFile("stats", h, size, objectInfo.BlockRef)
	// Store logs and add logs file
	object, size, err = logger.Close()
	if err != nil {
		return err
	}
	if object != nil {
		objectInfo, err := pachClient.InspectObject(object.Hash)
		if err != nil {
			return err
		}
		h, err := pfs.DecodeHash(object.Hash)
		if err != nil {
			return err
		}
		statsTree.PutFile("logs", h, size, objectInfo.BlockRef)
	}
	// Merge stats trees (input, output, stats) and write out
	inputBuf := &bytes.Buffer{}
	inputTree.Serialize(inputBuf)
	outputBuf := &bytes.Buffer{}
	outputTree.Serialize(outputBuf)
	statsBuf := &bytes.Buffer{}
	statsTree.Ordered().Serialize(statsBuf)
	// Merge datum stats hashtree
	buf := &bytes.Buffer{}
	if err := hashtree.Merge(hashtree.NewWriter(buf), []*hashtree.Reader{
		hashtree.NewReader(inputBuf, nil),
		hashtree.NewReader(outputBuf, nil),
		hashtree.NewReader(statsBuf, nil),
	}); err != nil {
		return err
	}
	// Write datum stats hashtree to object storage
	objW, err := pachClient.PutObjectAsync([]*pfs.Tag{client.NewTag(tag + statsTagSuffix)})
	if err != nil {
		return err
	}
	defer func() {
		if err := objW.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if _, err := objW.Write(buf.Bytes()); err != nil {
		return err
	}
	// Cache datum stats hashtree locally
	return a.datumStatsCache.Put(datumIdx, bytes.NewReader(buf.Bytes()))
}

// mergeChunk merges the datum hashtrees into a chunk hashtree and stores it.
func (a *APIServer) mergeChunk(logger logs.TaggedLogger, high int64, result *processResult) (retErr error) {
	logger.Logf("starting to merge chunk")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored merging chunk after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished merging chunk after %v", time.Since(start))
		}
	}(time.Now())
	buf := &bytes.Buffer{}
	if result.datumsFailed <= 0 {
		if err := a.datumCache.Merge(hashtree.NewWriter(buf), nil, nil); err != nil {
			return err
		}
	}
	if err := a.chunkCache.Put(high, buf); err != nil {
		return err
	}
	if a.pipelineInfo.EnableStats {
		buf.Reset()
		if err := a.datumStatsCache.Merge(hashtree.NewWriter(buf), nil, nil); err != nil {
			return err
		}
		return a.chunkStatsCache.Put(high, buf)
	}
	return nil
}

func (a *APIServer) acquireDatums(ctx context.Context, jobID string, plan *common.Plan, logger logs.TaggedLogger, process processFunc) error {
	chunks := a.driver.Chunks(jobID)
	watcher, err := chunks.ReadOnly(ctx).Watch(watch.WithFilterPut())
	if err != nil {
		return fmt.Errorf("error creating chunk watcher: %v", err)
	}
	var complete bool
	for !complete {
		// We set complete to true and then unset it if we find an incomplete chunk
		complete = true
		// Attempt to claim a chunk
		low, high := int64(0), int64(0)
		for _, high = range plan.Chunks {
			var chunkState common.ChunkState
			if err := chunks.Claim(ctx, fmt.Sprint(high), &chunkState, func(ctx context.Context) error {
				return a.processChunk(ctx, jobID, low, high, process)
			}); err == col.ErrNotClaimed {
				// Check if a different worker is processing this chunk
				if chunkState.State == common.State_RUNNING {
					complete = false
				}
			} else if err != nil {
				return err
			}
			low = high
		}
		// Wait for a deletion event (ttl expired) before attempting to claim a chunk again
		select {
		case e := <-watcher.Watch():
			if e.Type == watch.EventError {
				return fmt.Errorf("chunk watch error: %v", e.Err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (a *APIServer) processChunk(ctx context.Context, jobID string, low, high int64, process processFunc) error {
	processResult, err := process(low, high)
	if err != nil {
		return err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.driver.Jobs().ReadWrite(stm)
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Update(jobID, jobPtr, func() error {
			jobPtr.DataProcessed += processResult.datumsProcessed
			jobPtr.DataSkipped += processResult.datumsSkipped
			jobPtr.DataRecovered += processResult.datumsRecovered
			jobPtr.DataFailed += processResult.datumsFailed
			return nil
		}); err != nil {
			return err
		}
		chunks := a.driver.Chunks(jobID).ReadWrite(stm)
		if processResult.failedDatumID != "" {
			return chunks.Put(fmt.Sprint(high), &common.ChunkState{
				State:   common.State_FAILURE,
				DatumID: processResult.failedDatumID,
				Address: os.Getenv(client.PPSWorkerIPEnv),
			})
		}
		return chunks.Put(fmt.Sprint(high), &common.ChunkState{
			State:           common.State_SUCCESS,
			Address:         os.Getenv(client.PPSWorkerIPEnv),
			RecoveredDatums: processResult.recoveredDatums,
		})
	})
	return err
}

func (a *APIServer) mergeDatums(
	jobCtx context.Context,
	pachClient *client.APIClient,
	jobInfo *pps.JobInfo,
	jobID string,
	plan *common.Plan,
	logger logs.TaggedLogger,
	dit datum.Iterator,
	skip map[string]bool,
	useParentHashTree bool,
) (retErr error) {
	for {
		if err := func() error {
			// if this worker is not responsible for a shard, it waits to be assigned one or for the job to finish
			if a.shardCtx == nil {
				select {
				case a.shardCtx = <-a.claimedShard:
					a.shard = a.shardCtx.Value(shardKey).(int64)
				case <-jobCtx.Done():
					return nil
				}
			}
			ctx, _ := joincontext.Join(jobCtx, a.shardCtx)
			objClient, err := obj.NewClientFromSecret(a.hashtreeStorage)
			if err != nil {
				return err
			}
			// collect hashtrees from chunks as they complete
			var failed bool
			if err := logger.LogStep("collecting chunk hashtree(s)", func() error {
				low := int64(0)
				chunks := a.driver.Chunks(jobInfo.Job.ID).ReadOnly(ctx)
				for _, high := range plan.Chunks {
					chunkState := &common.ChunkState{}
					if err := chunks.WatchOneF(fmt.Sprint(high), func(e *watch.Event) error {
						if e.Type == watch.EventError {
							return e.Err
						}
						// unmarshal and check that full key matched
						var key string
						if err := e.Unmarshal(&key, chunkState); err != nil {
							return err
						}
						if key != fmt.Sprint(high) {
							return nil
						}
						switch chunkState.State {
						case common.State_FAILURE:
							failed = true
							fallthrough
						case common.State_SUCCESS:
							if err := a.getChunk(ctx, high, chunkState.Address, failed); err != nil {
								logger.Logf("error downloading chunk %v from worker at %v (%v), falling back on object storage", high, chunkState.Address, err)
								tags := a.computeTags(dit, low, high, skip, useParentHashTree)
								// Download datum hashtrees from object storage if we run into an error getting them from the worker
								if err := a.getChunkFromObjectStorage(ctx, pachClient, objClient, tags, high, failed); err != nil {
									return err
								}
							}
							return errutil.ErrBreak
						}
						return nil
					}); err != nil {
						return err
					}
					low = high
				}
			}); err != nil {
				return err
			}
			// get parent hashtree reader if it is being used
			var parentHashtree, parentStatsHashtree io.Reader
			if useParentHashTree {
				var r io.ReadCloser
				if err := logger.LogStep("getting parent hashtree", func() error {
					var err error
					r, err = a.getParentHashTree(ctx, pachClient, objClient, jobInfo.OutputCommit, a.shard)
					return err
				}); err != nil {
					return err
				}
				defer func() {
					if err := r.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				parentHashtree = bufio.NewReaderSize(r, parentTreeBufSize)
				// get parent stats hashtree reader if it is being used
				if a.pipelineInfo.EnableStats {
					var r io.ReadCloser
					if err := logger.LogStep("getting parent stats hashtree", func() error {
						var err error
						r, err = a.getParentHashTree(ctx, pachClient, objClient, jobInfo.StatsCommit, a.shard)
						return err
					}); err != nil {
						return err
					}
					defer func() {
						if err := r.Close(); err != nil && retErr == nil {
							retErr = err
						}
					}()
					parentStatsHashtree = bufio.NewReaderSize(r, parentTreeBufSize)
				}
			}
			// merging output tree(s)
			var tree, statsTree *pfs.Object
			var size, statsSize uint64
			if err := logger.LogStep("merging output", func() error {
				if a.pipelineInfo.EnableStats {
					statsTree, statsSize, err = a.merge(pachClient, objClient, true, parentStatsHashtree)
					if err != nil {
						return err
					}
				}
				if !failed {
					tree, size, err = a.merge(pachClient, objClient, false, parentHashtree)
					if err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
			// mark merge as complete
			_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				merges := a.driver.Merges(jobID).ReadWrite(stm)
				mergeState := &common.MergeState{
					State:          common.State_SUCCESS,
					Tree:           tree,
					SizeBytes:      size,
					StatsTree:      statsTree,
					StatsSizeBytes: statsSize,
				}
				return merges.Put(fmt.Sprint(a.shard), mergeState)
			})
			return err
		}(); err != nil {
			if a.shardCtx.Err() == context.Canceled {
				a.shardCtx = nil
				continue
			}
			return err
		}
		return nil
	}
}

func (a *APIServer) getChunk(ctx context.Context, id int64, address string, failed bool) error {
	// If this worker processed the chunk, then it is already in the chunk cache
	if address == os.Getenv(client.PPSWorkerIPEnv) {
		return nil
	}
	if _, ok := a.clients[address]; !ok {
		client, err := NewClient(address)
		if err != nil {
			return err
		}
		a.clients[address] = client
	}
	client := a.clients[address]
	// Get chunk hashtree and store in chunk cache if the chunk succeeded
	if !failed {
		c, err := client.GetChunk(ctx, &GetChunkRequest{
			Id:    id,
			Shard: a.shard,
		})
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := grpcutil.WriteFromStreamingBytesClient(c, buf); err != nil {
			return err
		}
		if err := a.chunkCache.Put(id, buf); err != nil {
			return err
		}
	}
	// Get chunk stats hashtree and store in chunk stats cache if applicable
	if a.pipelineInfo.EnableStats {
		c, err := client.GetChunk(ctx, &GetChunkRequest{
			Id:    id,
			Shard: a.shard,
			Stats: true,
		})
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := grpcutil.WriteFromStreamingBytesClient(c, buf); err != nil {
			return err
		}
		if err := a.chunkStatsCache.Put(id, buf); err != nil {
			return err
		}
	}
	return nil
}

func (a *APIServer) getChunkFromObjectStorage(ctx context.Context, pachClient *client.APIClient, objClient obj.Client, tags []*pfs.Tag, id int64, failed bool) error {
	// Download, merge, and cache datum hashtrees for a chunk if it succeeded
	if !failed {
		ts, err := a.getHashtrees(ctx, pachClient, objClient, tags, hashtree.NewFilter(a.numShards, a.shard))
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := hashtree.Merge(hashtree.NewWriter(buf), ts); err != nil {
			return err
		}
		if err := a.chunkCache.Put(id, buf); err != nil {
			return err
		}
	}
	// Download, merge, and cache datum stats hashtrees for a chunk if applicable
	if a.pipelineInfo.EnableStats {
		var statsTags []*pfs.Tag
		for _, tag := range tags {
			statsTags = append(statsTags, client.NewTag(tag.Name+statsTagSuffix))
		}
		ts, err := a.getHashtrees(ctx, pachClient, objClient, statsTags, hashtree.NewFilter(a.numShards, a.shard))
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := hashtree.Merge(hashtree.NewWriter(buf), ts); err != nil {
			return err
		}
		if err := a.chunkStatsCache.Put(id, buf); err != nil {
			return err
		}
	}
	return nil
}

func (a *APIServer) merge(pachClient *client.APIClient, objClient obj.Client, stats bool, parent io.Reader) (*pfs.Object, uint64, error) {
	var tree *pfs.Object
	var size uint64
	if err := func() (retErr error) {
		objW, err := pachClient.PutObjectAsync(nil)
		if err != nil {
			return err
		}
		w := hashtree.NewWriter(objW)
		filter := hashtree.NewFilter(a.numShards, a.shard)
		if stats {
			err = a.chunkStatsCache.Merge(w, parent, filter)
		} else {
			err = a.chunkCache.Merge(w, parent, filter)
		}
		size = w.Size()
		if err != nil {
			objW.Close()
			return err
		}
		// Get object hash for hashtree
		if err := objW.Close(); err != nil {
			return err
		}
		tree, err = objW.Object()
		if err != nil {
			return err
		}
		// Get index and write it out
		idx, err := w.Index()
		if err != nil {
			return err
		}
		return writeIndex(pachClient, objClient, tree, idx)
	}(); err != nil {
		return nil, 0, err
	}
	return tree, size, nil
}

func (a *APIServer) computeTags(dit datum.Iterator, low, high int64, skip map[string]bool, useParentHashTree bool) []*pfs.Tag {
	var tags []*pfs.Tag
	for i := low; i < high; i++ {
		files := dit.DatumN(int(i))
		datumHash := common.HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, files)
		// Skip datum if it is in the parent hashtree and the parent hashtree is being used in the merge
		if skip[datumHash] && useParentHashTree {
			continue
		}
		tags = append(tags, client.NewTag(datumHash))
	}
	return tags
}

func (a *APIServer) getHashtrees(ctx context.Context, pachClient *client.APIClient, objClient obj.Client, tags []*pfs.Tag, filter hashtree.Filter) ([]*hashtree.Reader, error) {
	limiter := limit.New(hashtree.DefaultMergeConcurrency)
	var eg errgroup.Group
	var mu sync.Mutex
	var rs []*hashtree.Reader
	for _, tag := range tags {
		tag := tag
		limiter.Acquire()
		eg.Go(func() (retErr error) {
			defer limiter.Release()
			// Get datum hashtree info
			info, err := pachClient.InspectTag(ctx, tag)
			if err != nil {
				return err
			}
			path, err := obj.BlockPathFromEnv(info.BlockRef.Block)
			if err != nil {
				return err
			}
			// Read the full datum hashtree in memory
			objR, err := objClient.Reader(ctx, path, 0, 0)
			if err != nil {
				return err
			}
			defer func() {
				if err := objR.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			fullTree, err := ioutil.ReadAll(objR)
			if err != nil {
				return err
			}
			// Filter out unecessary keys
			filteredTree := &bytes.Buffer{}
			w := hashtree.NewWriter(filteredTree)
			r := hashtree.NewReader(bytes.NewBuffer(fullTree), filter)
			if err := w.Copy(r); err != nil {
				return err
			}
			// Add it to the list of readers
			mu.Lock()
			defer mu.Unlock()
			rs = append(rs, hashtree.NewReader(filteredTree, nil))
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return rs, nil
}

func (a *APIServer) getParentHashTree(
	ctx context.Context,
	pachClient *client.APIClient,
	objClient obj.Client,
	commit *pfs.Commit,
	merge int64,
) (io.ReadCloser, error) {
	parentCommitInfo, err := a.getParentCommitInfo(ctx, pachClient, commit)
	if err != nil {
		return nil, err
	}
	if parentCommitInfo == nil {
		return ioutil.NopCloser(&bytes.Buffer{}), nil
	}
	info, err := pachClient.InspectObject(parentCommitInfo.Trees[merge].Hash)
	if err != nil {
		return nil, err
	}
	path, err := obj.BlockPathFromEnv(info.BlockRef.Block)
	if err != nil {
		return nil, err
	}
	return objClient.Reader(ctx, path, 0, 0)
}

func writeIndex(
	pachClient *client.APIClient,
	objClient obj.Client,
	tree *pfs.Object,
	idx []byte,
) (retErr error) {
	info, err := pachClient.InspectObject(tree.Hash)
	if err != nil {
		return err
	}
	path, err := obj.BlockPathFromEnv(info.BlockRef.Block)
	if err != nil {
		return err
	}
	idxW, err := objClient.Writer(pachClient.Ctx(), path+hashtree.IndexPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := idxW.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err = idxW.Write(idx)
	return err
}

// cancelCtxIfJobFails watches jobID's JobPtr, and if its state is changed to a
// terminal state (KILLED, FAILURE, or SUCCESS) cancel the jobCtx so we kill any
// user processes
func (a *APIServer) cancelCtxIfJobFails(jobCtx context.Context, jobCancel func(), jobID string) {
	logger := logs.NewStatlessLogger(a.pipelineInfo)

	backoff.RetryNotify(func() error {
		// Check if job was cancelled while backoff was sleeping
		if common.IsDone(jobCtx) {
			return nil
		}

		// Start watching for job state changes
		watcher, err := a.driver.Jobs().ReadOnly(jobCtx).WatchOne(jobID)
		if err != nil {
			if col.IsErrNotFound(err) {
				jobCancel() // job deleted before we started--cancel the job ctx
				return nil
			}
			return fmt.Errorf("worker: could not create state watcher for job %s, err is %v", jobID, err)
		}

		// If any job events indicate that the job is done, cancel jobCtx
	Outer:
		for {
			select {
			case e := <-watcher.Watch():
				switch e.Type {
				case watch.EventPut:
					var jobID string
					jobPtr := &pps.EtcdJobInfo{}
					if err := e.Unmarshal(&jobID, jobPtr); err != nil {
						return fmt.Errorf("worker: error unmarshalling while watching job state (%v)", err)
					}
					if ppsutil.IsTerminal(jobPtr.State) {
						logger.Logf("job %q put in terminal state %q; cancelling", jobID, jobPtr.State)
						jobCancel() // cancel the job
					}
				case watch.EventDelete:
					logger.Logf("job %q deleted; cancelling", jobID)
					jobCancel() // cancel the job
				case watch.EventError:
					return fmt.Errorf("job state watch error: %v", e.Err)
				}
			case <-jobCtx.Done():
				break Outer
			}
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if jobCtx.Err() == context.Canceled {
			return err // worker is done, nothing else to do
		}
		logger.Logf("worker: error watching job %s (%v); retrying in %v", jobID, err, d)
		return nil
	})
}
*/

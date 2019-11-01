package driver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/go-playground/webhooks.v5/github"
	"gopkg.in/src-d/go-git.v4"
	gitPlumbing "gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/exec"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/pachyderm/pachyderm/src/server/worker/stats"
)

const (
	// The maximum number of concurrent download/upload operations
	concurrency = 100

	chunkPrefix = "/chunk"
	mergePrefix = "/merge"
	planPrefix  = "/plan"
	shardPrefix = "/shard"
)

// Driver provides an interface for common functions needed by worker code, and
// captures the relevant objects necessary to provide these functions so that
// users do not need to keep track of as many variables.  In addition, this
// interface can be used to mock out external calls to make unit-testing
// simpler.
type Driver interface {
	Jobs() col.Collection
	Pipelines() col.Collection
	Plans() col.Collection
	Shards() col.Collection
	Chunks(jobID string) col.Collection
	Merges(jobID string) col.Collection

	// Returns the path that will contain the input filesets for the job
	InputDir() string

	// Returns the number of workers to be used based on what can be determined from kubernetes
	GetExpectedNumWorkers() (int, error)

	// WithCtx clones the current driver and applies the context to its pachClient
	WithCtx(context.Context) Driver

	// WithData prepares the current node the code is running on to run a piece
	// of user code by downloading the specified data, and cleans up afterwards.
	WithData(context.Context, []*common.Input, *hashtree.Ordered, logs.TaggedLogger, func(*pps.ProcessStats) error) (*pps.ProcessStats, error)

	// RunUserCode runs the pipeline's configured code
	RunUserCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error

	// RunUserErrorHandlingCode runs the pipeline's configured error handling code
	RunUserErrorHandlingCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error

	// TODO: provide a more generic interface for modifying jobs/plans/etc, and
	// some quality-of-life functions for common operations.
	DeleteJob(col.STM, *pps.EtcdJobInfo) error
	UpdateJobState(context.Context, string, *pfs.Commit, pps.JobState, string) error

	// TODO: figure out how to not expose this
	ReportUploadStats(time.Time, *pps.ProcessStats, logs.TaggedLogger)

	// TODO: figure out how to not expose this - currenly only used for a few
	// operations in the map spawner
	NewSTM(context.Context, func(col.STM) error) (*etcd.TxnResponse, error)
}

type driver struct {
	pipelineInfo *pps.PipelineInfo
	pachClient   *client.APIClient
	kubeWrapper  KubeWrapper
	etcdClient   *etcd.Client
	etcdPrefix   string

	jobs col.Collection

	pipelines col.Collection

	plans col.Collection

	// Stores available filesystem shards for a pipeline, workers will claim these
	shards col.Collection

	// User and group IDs used for running user code, determined in the constructor
	uid *uint32
	gid *uint32

	// We only export application statistics if enterprise is enabled
	exportStats bool

	// The directory to store input data - this is typically static but can be
	// overridden by tests
	inputDir string
}

// NewDriver constructs a Driver object using the given clients and pipeline
// settings.  It makes blocking calls to determine the user/group to use with
// the user code on the current worker node, as well as determining if
// enterprise features are activated (for exporting stats).
func NewDriver(
	pipelineInfo *pps.PipelineInfo,
	pachClient *client.APIClient,
	kubeWrapper KubeWrapper,
	etcdClient *etcd.Client,
	etcdPrefix string,
) (Driver, error) {
	result := &driver{
		pipelineInfo: pipelineInfo,
		pachClient:   pachClient,
		kubeWrapper:  kubeWrapper,
		etcdClient:   etcdClient,
		etcdPrefix:   etcdPrefix,
		jobs:         ppsdb.Jobs(etcdClient, etcdPrefix),
		pipelines:    ppsdb.Pipelines(etcdClient, etcdPrefix),
		shards:       col.NewCollection(etcdClient, path.Join(etcdPrefix, shardPrefix, pipelineInfo.Pipeline.Name), nil, &common.ShardInfo{}, nil, nil),
		plans:        col.NewCollection(etcdClient, path.Join(etcdPrefix, planPrefix), nil, &common.Plan{}, nil, nil),
		inputDir:     client.PPSInputPrefix,
	}

	if pipelineInfo.Transform.User != "" {
		user, err := lookupDockerUser(pipelineInfo.Transform.User)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		// If `user` is `nil`, `uid` and `gid` will get set, and we won't
		// customize the user that executes the worker process.
		if user != nil { // user is nil when os.IsNotExist(err) is true in which case we use root
			uid, err := strconv.ParseUint(user.Uid, 10, 32)
			if err != nil {
				return nil, err
			}
			uid32 := uint32(uid)
			result.uid = &uid32
			gid, err := strconv.ParseUint(user.Gid, 10, 32)
			if err != nil {
				return nil, err
			}
			gid32 := uint32(gid)
			result.gid = &gid32
		}
	}

	if resp, err := pachClient.Enterprise.GetState(context.Background(), &enterprise.GetStateRequest{}); err != nil {
		logs.NewStatlessLogger(pipelineInfo).Logf("failed to get enterprise state with error: %v\n", err)
	} else {
		result.exportStats = resp.State == enterprise.State_ACTIVE
	}

	return result, nil
}

// lookupDockerUser looks up users given the argument to a Dockerfile USER directive.
// According to Docker's docs this directive looks like:
// USER <user>[:<group>] or
// USER <UID>[:<GID>]
func lookupDockerUser(userArg string) (_ *user.User, retErr error) {
	userParts := strings.Split(userArg, ":")
	userOrUID := userParts[0]
	groupOrGID := ""
	if len(userParts) > 1 {
		groupOrGID = userParts[1]
	}
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := passwd.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] == userOrUID || parts[2] == userOrUID {
			result := &user.User{
				Username: parts[0],
				Uid:      parts[2],
				Gid:      parts[3],
				Name:     parts[4],
				HomeDir:  parts[5],
			}
			if groupOrGID != "" {
				if parts[0] == userOrUID {
					// groupOrGid is a group
					group, err := lookupGroup(groupOrGID)
					if err != nil {
						return nil, err
					}
					result.Gid = group.Gid
				} else {
					// groupOrGid is a gid
					result.Gid = groupOrGID
				}
			}
			return result, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil, fmt.Errorf("user %s not found", userArg)
}

func lookupGroup(group string) (_ *user.Group, retErr error) {
	groupFile, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := groupFile.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	scanner := bufio.NewScanner(groupFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] == group {
			return &user.Group{
				Gid:  parts[2],
				Name: parts[0],
			}, nil
		}
	}
	return nil, fmt.Errorf("group %s not found", group)
}

func (d *driver) WithCtx(ctx context.Context) Driver {
	result := &driver{}
	*result = *d
	result.pachClient = result.pachClient.WithCtx(ctx)
	return result
}

func (d *driver) Jobs() col.Collection {
	return d.jobs
}

func (d *driver) Pipelines() col.Collection {
	return d.pipelines
}

func (d *driver) Plans() col.Collection {
	return d.plans
}

func (d *driver) Shards() col.Collection {
	return d.shards
}

func (d *driver) Chunks(jobID string) col.Collection {
	return col.NewCollection(d.etcdClient, path.Join(d.etcdPrefix, chunkPrefix, jobID), nil, &common.ChunkState{}, nil, nil)
}

func (d *driver) Merges(jobID string) col.Collection {
	return col.NewCollection(d.etcdClient, path.Join(d.etcdPrefix, mergePrefix, jobID), nil, &common.MergeState{}, nil, nil)
}

func (d *driver) GetExpectedNumWorkers() (int, error) {
	return d.kubeWrapper.GetExpectedNumWorkers(d.pipelineInfo.ParallelismSpec)
}

func (d *driver) InputDir() string {
	return d.inputDir
}

func (d *driver) NewSTM(ctx context.Context, cb func(col.STM) error) (*etcd.TxnResponse, error) {
	return col.NewSTM(ctx, d.etcdClient, cb)
}

func (d *driver) WithData(
	ctx context.Context,
	data []*common.Input,
	inputTree *hashtree.Ordered,
	logger logs.TaggedLogger,
	cb func(*pps.ProcessStats) error,
) (retStats *pps.ProcessStats, retErr error) {
	puller := filesync.NewPuller()
	stats := &pps.ProcessStats{}

	// Download input data into a temporary directory
	// TODO: pass down ctx and use it for interruption
	dir, err := d.downloadData(d.pachClient, logger, data, puller, stats, inputTree)
	// We run these cleanup functions no matter what, so that if
	// downloadData partially succeeded, we still clean up the resources.
	defer func() {
		if err := os.RemoveAll(dir); err != nil && retErr == nil {
			retErr = err
		}
	}()
	// It's important that we run puller.CleanUp before os.RemoveAll,
	// because otherwise puller.Cleanup might try to open pipes that have
	// been deleted.
	defer func() {
		if _, err := puller.CleanUp(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("error downloadData: %v", err)
	}
	if err := os.MkdirAll(d.inputDir, 0777); err != nil {
		return nil, err
	}
	if err := d.linkData(data, dir); err != nil {
		return nil, fmt.Errorf("error linkData: %v", err)
	}
	defer func() {
		if err := d.unlinkData(data); err != nil && retErr == nil {
			retErr = fmt.Errorf("error unlinkData: %v", err)
		}
	}()
	// If the pipeline spec set a custom user to execute the process, make sure
	// the input directory and its content are owned by it
	if d.uid != nil && d.gid != nil {
		filepath.Walk(d.inputDir, func(name string, info os.FileInfo, err error) error {
			if err == nil {
				err = os.Chown(name, int(*d.uid), int(*d.gid))
			}
			return err
		})
	}

	if err := cb(stats); err != nil {
		return nil, err
	}

	// CleanUp is idempotent so we can call it however many times we want.
	// The reason we are calling it here is that the puller could've
	// encountered an error as it was lazily loading files, in which case
	// the output might be invalid since as far as the user's code is
	// concerned, they might've just seen an empty or partially completed
	// file.
	// TODO: do we really need two puller.CleanUps?
	downSize, err := puller.CleanUp()
	if err != nil {
		logger.Logf("puller encountered an error while cleaning up: %+v", err)
		return nil, err
	}

	atomic.AddUint64(&stats.DownloadBytes, uint64(downSize))
	d.reportDownloadSizeStats(float64(downSize), logger)
	return stats, nil
}

func (d *driver) downloadData(
	pachClient *client.APIClient,
	logger logs.TaggedLogger,
	inputs []*common.Input,
	puller *filesync.Puller,
	stats *pps.ProcessStats,
	statsTree *hashtree.Ordered,
) (_ string, retErr error) {
	defer d.reportDownloadTimeStats(time.Now(), stats, logger)
	logger.Logf("starting to download data")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored downloading data after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished downloading data after %v", time.Since(start))
		}
	}(time.Now())

	// The scratch space is where Pachyderm stores downloaded and output data, which is
	// then symlinked into place for the pipeline.
	dir := filepath.Join(d.inputDir, client.PPSScratchSpace, uuid.NewWithoutDashes())
	// Create output directory (currently /pfs/out)
	outPath := filepath.Join(dir, "out")
	if d.pipelineInfo.Spout != nil {
		// Spouts need to create a named pipe at /pfs/out
		if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
			return "", fmt.Errorf("mkdirall :%v", err)
		}
		// TODO: this doesn't build on windows
		/*
			if err := syscall.Mkfifo(outPath, 0666); err != nil {
				return "", fmt.Errorf("mkfifo :%v", err)
			}
		*/
	} else {
		if err := os.MkdirAll(outPath, 0777); err != nil {
			return "", err
		}
	}
	for _, input := range inputs {
		if input.GitURL != "" {
			if err := d.downloadGitData(pachClient, dir, input); err != nil {
				return "", err
			}
			continue
		}
		file := input.FileInfo.File
		root := filepath.Join(dir, input.Name, file.Path)
		var statsRoot string
		if statsTree != nil {
			statsRoot = path.Join(input.Name, file.Path)
			parent, _ := path.Split(statsRoot)
			statsTree.MkdirAll(parent)
		}
		if err := puller.Pull(pachClient, root, file.Commit.Repo.Name, file.Commit.ID, file.Path, input.Lazy, input.EmptyFiles, concurrency, statsTree, statsRoot); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func (d *driver) downloadGitData(pachClient *client.APIClient, dir string, input *common.Input) error {
	file := input.FileInfo.File
	pachydermRepoName := input.Name
	var rawJSON bytes.Buffer
	err := pachClient.GetFile(pachydermRepoName, file.Commit.ID, "commit.json", 0, 0, &rawJSON)
	if err != nil {
		return err
	}
	var payload github.PushPayload
	err = json.Unmarshal(rawJSON.Bytes(), &payload)
	if err != nil {
		return err
	}
	sha := payload.After
	// Clone checks out a reference, not a SHA
	r, err := git.PlainClone(
		filepath.Join(dir, pachydermRepoName),
		false,
		&git.CloneOptions{
			URL:           payload.Repository.CloneURL,
			SingleBranch:  true,
			ReferenceName: gitPlumbing.ReferenceName(payload.Ref),
		},
	)
	if err != nil {
		return fmt.Errorf("error cloning repo %v at SHA %v from URL %v: %v", pachydermRepoName, sha, input.GitURL, err)
	}
	hash := gitPlumbing.NewHash(sha)
	wt, err := r.Worktree()
	if err != nil {
		return err
	}
	err = wt.Checkout(&git.CheckoutOptions{Hash: hash})
	if err != nil {
		return fmt.Errorf("error checking out SHA %v from repo %v: %v", sha, pachydermRepoName, err)
	}
	return nil
}

func (d *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that previously symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.inputDir, input.Name)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return os.Symlink(filepath.Join(dir, "out"), filepath.Join(d.inputDir, "out"))
}

func (d *driver) unlinkData(inputs []*common.Input) error {
	entries, err := ioutil.ReadDir(d.inputDir)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == client.PPSScratchSpace {
			continue // don't delete scratch space
		}
		if err := os.RemoveAll(filepath.Join(d.inputDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (d *driver) RunUserCode(ctx context.Context, logger logs.TaggedLogger, environ []string, procStats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	d.reportUserCodeStats(logger)
	defer func(start time.Time) { d.reportDeferredUserCodeStats(retErr, start, procStats, logger) }(time.Now())
	logger.Logf("beginning to run user code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user code after %v", time.Since(start))
		}
	}(time.Now())
	if rawDatumTimeout != nil {
		datumTimeout, err := types.DurationFromProto(rawDatumTimeout)
		if err != nil {
			return err
		}
		datumTimeoutCtx, cancel := context.WithTimeout(ctx, datumTimeout)
		defer cancel()
		ctx = datumTimeoutCtx
	}

	// Run user code
	cmd := exec.CommandContext(ctx, d.pipelineInfo.Transform.Cmd[0], d.pipelineInfo.Transform.Cmd[1:]...)
	if d.pipelineInfo.Transform.Stdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(d.pipelineInfo.Transform.Stdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	// TODO: this doesn't work on windows
	/*
		if d.uid != nil && d.gid != nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{
					Uid: *d.uid,
					Gid: *d.gid,
				},
			}
		}
	*/
	cmd.Dir = d.pipelineInfo.Transform.WorkingDir
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error cmd.Start: %v", err)
	}
	// A context w a deadline will successfully cancel/kill
	// the running process (minus zombies)
	state, err := cmd.Process.Wait()
	if err != nil {
		return fmt.Errorf("error cmd.Wait: %v", err)
	}
	if common.IsDone(ctx) {
		if err = ctx.Err(); err != nil {
			return err
		}
	}

	// Because of this issue: https://github.com/golang/go/issues/18874
	// We forked os/exec so that we can call just the part of cmd.Wait() that
	// happens after blocking on the process. Unfortunately calling
	// cmd.Process.Wait() then cmd.Wait() will produce an error. So instead we
	// close the IO using this helper
	err = cmd.WaitIO(state, err)
	// We ignore broken pipe errors, these occur very occasionally if a user
	// specifies Stdin but their process doesn't actually read everything from
	// Stdin. This is a fairly common thing to do, bash by default ignores
	// broken pipe errors.
	if err != nil && !strings.Contains(err.Error(), "broken pipe") {
		// (if err is an acceptable return code, don't return err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range d.pipelineInfo.Transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						return nil
					}
				}
			}
		}
		return fmt.Errorf("error cmd.WaitIO: %v", err)
	}
	return nil
}

// Run user error code and return the combined output of stdout and stderr.
func (d *driver) RunUserErrorHandlingCode(ctx context.Context, logger logs.TaggedLogger, environ []string, procStats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	logger.Logf("beginning to run user error handling code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user error handling code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user error handling code after %v", time.Since(start))
		}
	}(time.Now())

	cmd := exec.CommandContext(ctx, d.pipelineInfo.Transform.ErrCmd[0], d.pipelineInfo.Transform.ErrCmd[1:]...)
	if d.pipelineInfo.Transform.ErrStdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(d.pipelineInfo.Transform.ErrStdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	// TODO: this doesn't work on windows
	/*
		if d.uid != nil && d.gid != nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{
					Uid: *d.uid,
					Gid: *d.gid,
				},
			}
		}
	*/
	cmd.Dir = d.pipelineInfo.Transform.WorkingDir
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error cmd.Start: %v", err)
	}
	// A context w a deadline will successfully cancel/kill
	// the running process (minus zombies)
	state, err := cmd.Process.Wait()
	if err != nil {
		return fmt.Errorf("error cmd.Wait: %v", err)
	}
	if common.IsDone(ctx) {
		if err = ctx.Err(); err != nil {
			return err
		}
	}
	// Because of this issue: https://github.com/golang/go/issues/18874
	// We forked os/exec so that we can call just the part of cmd.Wait() that
	// happens after blocking on the process. Unfortunately calling
	// cmd.Process.Wait() then cmd.Wait() will produce an error. So instead we
	// close the IO using this helper
	err = cmd.WaitIO(state, err)
	// We ignore broken pipe errors, these occur very occasionally if a user
	// specifies Stdin but their process doesn't actually read everything from
	// Stdin. This is a fairly common thing to do, bash by default ignores
	// broken pipe errors.
	if err != nil && !strings.Contains(err.Error(), "broken pipe") {
		// (if err is an acceptable return code, don't return err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range d.pipelineInfo.Transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						return nil
					}
				}
			}
		}
		return fmt.Errorf("error cmd.WaitIO: %v", err)
	}
	return nil
}

func (d *driver) UpdateJobState(ctx context.Context, jobID string, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		jobPtr := &pps.EtcdJobInfo{}
		if err := d.Jobs().ReadWrite(stm).Get(jobID, jobPtr); err != nil {
			return err
		}
		// TODO: move this out
		if jobPtr.StatsCommit == nil {
			jobPtr.StatsCommit = statsCommit
		}

		return ppsutil.UpdateJobState(d.Pipelines().ReadWrite(stm), d.Jobs().ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

// DeleteJob is identical to updateJobState, except that jobPtr points to a job
// that should be deleted rather than marked failed. Jobs may be deleted if
// their output commit is deleted.
func (d *driver) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	pipelinePtr := &pps.EtcdPipelineInfo{}
	if err := d.Pipelines().ReadWrite(stm).Update(jobPtr.Pipeline.Name, pipelinePtr, func() error {
		if pipelinePtr.JobCounts == nil {
			pipelinePtr.JobCounts = make(map[int32]int32)
		}
		if pipelinePtr.JobCounts[int32(jobPtr.State)] != 0 {
			pipelinePtr.JobCounts[int32(jobPtr.State)]--
		}
		return nil
	}); err != nil {
		return err
	}
	return d.Jobs().ReadWrite(stm).Delete(jobPtr.Job.ID)
}

func (d *driver) updateCounter(
	stat *prometheus.CounterVec,
	logger logs.TaggedLogger,
	state string,
	cb func(prometheus.Counter),
) {
	labels := []string{d.pipelineInfo.ID, logger.JobID()}
	if state != "" {
		labels = append(labels, state)
	}
	if counter, err := stat.GetMetricWithLabelValues(labels...); err != nil {
		logger.Logf("failed to get counter with labels (%v): %v", labels, err)
	} else {
		cb(counter)
	}
}

func (d *driver) updateHistogram(
	stat *prometheus.HistogramVec,
	logger logs.TaggedLogger,
	state string,
	cb func(prometheus.Observer),
) {
	labels := []string{d.pipelineInfo.ID, logger.JobID()}
	if state != "" {
		labels = append(labels, state)
	}
	if hist, err := stat.GetMetricWithLabelValues(labels...); err != nil {
		logger.Logf("failed to get histogram with labels (%v): %v", labels, err)
	} else {
		cb(hist)
	}
}

func (d *driver) reportUserCodeStats(logger logs.TaggedLogger) {
	if d.exportStats {
		d.updateCounter(stats.DatumCount, logger, "started", func(counter prometheus.Counter) {
			counter.Add(1)
		})
	}
}

func (d *driver) reportDeferredUserCodeStats(err error, start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
	if d.exportStats {
		duration := time.Since(start)
		procStats.ProcessTime = types.DurationProto(duration)

		state := "errored"
		if err == nil {
			state = "finished"
		}

		d.updateCounter(stats.DatumCount, logger, state, func(counter prometheus.Counter) {
			counter.Add(1)
		})
		d.updateHistogram(stats.DatumProcTime, logger, state, func(hist prometheus.Observer) {
			hist.Observe(duration.Seconds())
		})
		d.updateCounter(stats.DatumProcSecondsCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(duration.Seconds())
		})
	}
}

func (d *driver) ReportUploadStats(start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
	if d.exportStats {
		duration := time.Since(start)
		procStats.UploadTime = types.DurationProto(duration)

		d.updateHistogram(stats.DatumUploadTime, logger, "", func(hist prometheus.Observer) {
			hist.Observe(duration.Seconds())
		})
		d.updateCounter(stats.DatumUploadSecondsCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(duration.Seconds())
		})
		d.updateHistogram(stats.DatumUploadSize, logger, "", func(hist prometheus.Observer) {
			hist.Observe(float64(procStats.UploadBytes))
		})
		d.updateCounter(stats.DatumUploadBytesCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(float64(procStats.UploadBytes))
		})
	}
}

func (d *driver) reportDownloadSizeStats(downSize float64, logger logs.TaggedLogger) {
	if d.exportStats {
		d.updateHistogram(stats.DatumDownloadSize, logger, "", func(hist prometheus.Observer) {
			hist.Observe(downSize)
		})
		d.updateCounter(stats.DatumDownloadBytesCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(downSize)
		})
	}
}

func (d *driver) reportDownloadTimeStats(start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
	if d.exportStats {
		duration := time.Since(start)
		procStats.DownloadTime = types.DurationProto(duration)

		d.updateHistogram(stats.DatumDownloadTime, logger, "", func(hist prometheus.Observer) {
			hist.Observe(duration.Seconds())
		})
		d.updateCounter(stats.DatumDownloadSecondsCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(duration.Seconds())
		})
	}
}

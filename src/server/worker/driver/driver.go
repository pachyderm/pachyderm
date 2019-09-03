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
	kube "k8s.io/client-go/kubernetes"

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
)

type Driver interface {
	Jobs() col.Collection
	Pipelines() col.Collection
	Plans() col.Collection
	Chunks(jobID string) col.Collection
	Merges(jobID string) col.Collection

	GetExpectedNumWorkers() (int, error)

	WithProvisionedNode(context.Context, []*common.Input, *hashtree.Ordered, logs.TaggedLogger, func(*pps.ProcessStats) error) (*pps.ProcessStats, error)
	RunUserCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error
	RunUserErrorHandlingCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error

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
	kubeClient   *kube.Clientset
	etcdClient   *etcd.Client
	etcdPrefix   string
	jobs         col.Collection
	pipelines    col.Collection
	plans        col.Collection

	uid *uint32
	gid *uint32

	// We only export application statistics if enterprise is enabled
	exportStats bool
}

func NewDriver(
	pipelineInfo *pps.PipelineInfo,
	pachClient *client.APIClient,
	kubeClient *kube.Clientset,
	etcdClient *etcd.Client,
	etcdPrefix string,
) (Driver, error) {
	result := &driver{
		pipelineInfo: pipelineInfo,
		pachClient:   pachClient,
		kubeClient:   kubeClient,
		etcdClient:   etcdClient,
		etcdPrefix:   etcdPrefix,
		jobs:         ppsdb.Jobs(etcdClient, etcdPrefix),
		pipelines:    ppsdb.Pipelines(etcdClient, etcdPrefix),
		plans:        col.NewCollection(etcdClient, path.Join(etcdPrefix, planPrefix), nil, &common.Plan{}, nil, nil),
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

func (u *driver) Jobs() col.Collection {
	return u.jobs
}

func (u *driver) Pipelines() col.Collection {
	return u.pipelines
}

func (u *driver) Plans() col.Collection {
	return u.plans
}

func (u *driver) Chunks(jobID string) col.Collection {
	return col.NewCollection(u.etcdClient, path.Join(u.etcdPrefix, chunkPrefix, jobID), nil, &common.ChunkState{}, nil, nil)
}

func (u *driver) Merges(jobID string) col.Collection {
	return col.NewCollection(u.etcdClient, path.Join(u.etcdPrefix, mergePrefix, jobID), nil, &common.MergeState{}, nil, nil)
}

func (u *driver) GetExpectedNumWorkers() (int, error) {
	return ppsutil.GetExpectedNumWorkers(u.kubeClient, u.pipelineInfo.ParallelismSpec)
}

func (u *driver) NewSTM(ctx context.Context, cb func(col.STM) error) (*etcd.TxnResponse, error) {
	return col.NewSTM(ctx, u.etcdClient, cb)
}

// Create puller *
// downloadData *
// defer removeAll *
// defer puller.Cleanup
// lock runMutex
// lock status mutex, set some dumb values
// mkdir
// linkData *
// defer unlinkData
// walk pfs and chown
// run user code and user error handling code
// puller.Cleanup
// report stats
// uploadOutput

func (u *driver) WithProvisionedNode(
	ctx context.Context,
	data []*common.Input,
	inputTree *hashtree.Ordered,
	logger logs.TaggedLogger,
	cb func(*pps.ProcessStats) error,
) (retStats *pps.ProcessStats, retErr error) {
	puller := filesync.NewPuller()
	stats := &pps.ProcessStats{}

	// Download input data
	dir, err := u.downloadData(u.pachClient, logger, data, puller, stats, inputTree)
	// We run these cleanup functions no matter what, so that if
	// downloadData partially succeeded, we still clean up the resources.
	defer func() {
		if err := os.RemoveAll(dir); err != nil && retErr == nil {
			retErr = err
		}
	}()
	// It's important that we run puller.CleanUp before os.RemoveAll,
	// because otherwise puller.Cleanup might try tp open pipes that have
	// been deleted.
	defer func() {
		if _, err := puller.CleanUp(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("error downloadData: %v", err)
	}
	if err := os.MkdirAll(client.PPSInputPrefix, 0777); err != nil {
		return nil, err
	}
	if err := u.linkData(data, dir); err != nil {
		return nil, fmt.Errorf("error linkData: %v", err)
	}
	defer func() {
		if err := u.unlinkData(data); err != nil && retErr == nil {
			retErr = fmt.Errorf("error unlinkData: %v", err)
		}
	}()
	// If the pipeline spec set a custom user to execute the
	// process, make sure `/pfs` and its content are owned by it
	if u.uid != nil && u.gid != nil {
		filepath.Walk("/pfs", func(name string, info os.FileInfo, err error) error {
			if err == nil {
				err = os.Chown(name, int(*u.uid), int(*u.gid))
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
	downSize, err := puller.CleanUp()
	if err != nil {
		logger.Logf("puller encountered an error while cleaning up: %+v", err)
		return nil, err
	}

	atomic.AddUint64(&stats.DownloadBytes, uint64(downSize))
	u.reportDownloadSizeStats(float64(downSize), logger)
	return stats, nil
}

func (u *driver) downloadData(pachClient *client.APIClient, logger logs.TaggedLogger, inputs []*common.Input, puller *filesync.Puller, stats *pps.ProcessStats, statsTree *hashtree.Ordered) (_ string, retErr error) {
	defer u.reportDownloadTimeStats(time.Now(), stats, logger)
	logger.Logf("starting to download data")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored downloading data after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished downloading data after %v", time.Since(start))
		}
	}(time.Now())
	dir := filepath.Join(client.PPSInputPrefix, client.PPSScratchSpace, uuid.NewWithoutDashes())
	// Create output directory (currently /pfs/out)
	outPath := filepath.Join(dir, "out")
	if u.pipelineInfo.Spout != nil {
		// Spouts need to create a named pipe at /pfs/out
		if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
			return "", fmt.Errorf("mkdirall :%v", err)
		}
		if err := syscall.Mkfifo(outPath, 0666); err != nil {
			return "", fmt.Errorf("mkfifo :%v", err)
		}
	} else {
		if err := os.MkdirAll(outPath, 0777); err != nil {
			return "", err
		}
	}
	for _, input := range inputs {
		if input.GitURL != "" {
			if err := u.downloadGitData(pachClient, dir, input); err != nil {
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

func (u *driver) downloadGitData(pachClient *client.APIClient, dir string, input *common.Input) error {
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

func (u *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that previously symlinked outputs are removed.
	if err := u.unlinkData(inputs); err != nil {
		return err
	}
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(client.PPSInputPrefix, input.Name)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return os.Symlink(filepath.Join(dir, "out"), filepath.Join(client.PPSInputPrefix, "out"))
}

func (u *driver) unlinkData(inputs []*common.Input) error {
	dirs, err := ioutil.ReadDir(client.PPSInputPrefix)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir: %v", err)
	}
	for _, d := range dirs {
		if d.Name() == client.PPSScratchSpace {
			continue // don't delete scratch space
		}
		if err := os.RemoveAll(filepath.Join(client.PPSInputPrefix, d.Name())); err != nil {
			return err
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (u *driver) RunUserCode(ctx context.Context, logger logs.TaggedLogger, environ []string, procStats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	u.reportUserCodeStats(logger)
	defer func(start time.Time) { u.reportDeferredUserCodeStats(retErr, start, procStats, logger) }(time.Now())
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
	cmd := exec.CommandContext(ctx, u.pipelineInfo.Transform.Cmd[0], u.pipelineInfo.Transform.Cmd[1:]...)
	if u.pipelineInfo.Transform.Stdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(u.pipelineInfo.Transform.Stdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	if u.uid != nil && u.gid != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: *u.uid,
				Gid: *u.gid,
			},
		}
	}
	cmd.Dir = u.pipelineInfo.Transform.WorkingDir
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
				for _, returnCode := range u.pipelineInfo.Transform.AcceptReturnCode {
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
func (u *driver) RunUserErrorHandlingCode(ctx context.Context, logger logs.TaggedLogger, environ []string, procStats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	logger.Logf("beginning to run user error handling code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user error handling code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user error handling code after %v", time.Since(start))
		}
	}(time.Now())

	cmd := exec.CommandContext(ctx, u.pipelineInfo.Transform.ErrCmd[0], u.pipelineInfo.Transform.ErrCmd[1:]...)
	if u.pipelineInfo.Transform.ErrStdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(u.pipelineInfo.Transform.ErrStdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	if u.uid != nil && u.gid != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: *u.uid,
				Gid: *u.gid,
			},
		}
	}
	cmd.Dir = u.pipelineInfo.Transform.WorkingDir
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
				for _, returnCode := range u.pipelineInfo.Transform.AcceptReturnCode {
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

func (u *driver) UpdateJobState(ctx context.Context, jobID string, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
	_, err := col.NewSTM(ctx, u.etcdClient, func(stm col.STM) error {
		jobs := u.jobs.ReadWrite(stm)
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Get(jobID, jobPtr); err != nil {
			return err
		}
		// TODO: move this out
		if jobPtr.StatsCommit == nil {
			jobPtr.StatsCommit = statsCommit
		}

		return ppsutil.UpdateJobState(u.pipelines.ReadWrite(stm), u.jobs.ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

// DeleteJob is identical to updateJobState, except that jobPtr points to a job
// that should be deleted rather than marked failed. Jobs may be deleted if
// their output commit is deleted.
func (u *driver) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	pipelinePtr := &pps.EtcdPipelineInfo{}
	if err := u.pipelines.ReadWrite(stm).Update(jobPtr.Pipeline.Name, pipelinePtr, func() error {
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
	return u.jobs.ReadWrite(stm).Delete(jobPtr.Job.ID)
}

func (u *driver) updateCounter(
	stat *prometheus.CounterVec,
	logger logs.TaggedLogger,
	state string,
	cb func(prometheus.Counter),
) {
	labels := []string{u.pipelineInfo.ID, logger.JobID()}
	if state != "" {
		labels = append(labels, state)
	}
	if counter, err := stat.GetMetricWithLabelValues(labels...); err != nil {
		logger.Logf("failed to get counter with labels (%v): %v", labels, err)
	} else {
		cb(counter)
	}
}

func (u *driver) updateHistogram(
	stat *prometheus.HistogramVec,
	logger logs.TaggedLogger,
	state string,
	cb func(prometheus.Observer),
) {
	labels := []string{u.pipelineInfo.ID, logger.JobID()}
	if state != "" {
		labels = append(labels, state)
	}
	if hist, err := stat.GetMetricWithLabelValues(labels...); err != nil {
		logger.Logf("failed to get histogram with labels (%v): %v", labels, err)
	} else {
		cb(hist)
	}
}

func (u *driver) reportUserCodeStats(logger logs.TaggedLogger) {
	if u.exportStats {
		u.updateCounter(stats.DatumCount, logger, "started", func(counter prometheus.Counter) {
			counter.Add(1)
		})
	}
}

func (u *driver) reportDeferredUserCodeStats(err error, start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
	if u.exportStats {
		duration := time.Since(start)
		procStats.ProcessTime = types.DurationProto(duration)

		state := "errored"
		if err == nil {
			state = "finished"
		}

		u.updateCounter(stats.DatumCount, logger, state, func(counter prometheus.Counter) {
			counter.Add(1)
		})
		u.updateHistogram(stats.DatumProcTime, logger, state, func(hist prometheus.Observer) {
			hist.Observe(duration.Seconds())
		})
		u.updateCounter(stats.DatumProcSecondsCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(duration.Seconds())
		})
	}
}

func (u *driver) ReportUploadStats(start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
	if u.exportStats {
		duration := time.Since(start)
		procStats.UploadTime = types.DurationProto(duration)

		u.updateHistogram(stats.DatumUploadTime, logger, "", func(hist prometheus.Observer) {
			hist.Observe(duration.Seconds())
		})
		u.updateCounter(stats.DatumUploadSecondsCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(duration.Seconds())
		})
		u.updateHistogram(stats.DatumUploadSize, logger, "", func(hist prometheus.Observer) {
			hist.Observe(float64(procStats.UploadBytes))
		})
		u.updateCounter(stats.DatumUploadBytesCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(float64(procStats.UploadBytes))
		})
	}
}

func (u *driver) reportDownloadSizeStats(downSize float64, logger logs.TaggedLogger) {
	if u.exportStats {
		u.updateHistogram(stats.DatumDownloadSize, logger, "", func(hist prometheus.Observer) {
			hist.Observe(downSize)
		})
		u.updateCounter(stats.DatumDownloadBytesCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(downSize)
		})
	}
}

func (u *driver) reportDownloadTimeStats(start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
	if u.exportStats {
		duration := time.Since(start)
		procStats.DownloadTime = types.DurationProto(duration)

		u.updateHistogram(stats.DatumDownloadTime, logger, "", func(hist prometheus.Observer) {
			hist.Observe(duration.Seconds())
		})
		u.updateCounter(stats.DatumDownloadSecondsCount, logger, "", func(counter prometheus.Counter) {
			counter.Add(duration.Seconds())
		})
	}
}

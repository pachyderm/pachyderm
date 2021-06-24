package driver

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/exec"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/work"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// TODO(2.0 optional):
// Prometheus stats? (refer to old driver code and tests)
// capture logs (refer to old driver code and tests).
// In general, need to spend some time walking through the old driver
// tests to see what can be reused.

// WorkNamespace returns the namespace used by the work package for this
// pipeline.
func WorkNamespace(pipelineInfo *pps.PipelineInfo) string {
	return fmt.Sprintf("/pipeline-%s/v%d", pipelineInfo.Pipeline.Name, pipelineInfo.Version)
}

// Driver provides an interface for common functions needed by worker code, and
// captures the relevant objects necessary to provide these functions so that
// users do not need to keep track of as many variables.  In addition, this
// interface can be used to mock out external calls to make unit-testing
// simpler.
type Driver interface {
	Jobs() col.PostgresCollection
	Pipelines() col.PostgresCollection

	NewTaskWorker() *work.Worker
	NewTaskQueue() (*work.TaskQueue, error)

	// Returns the PipelineInfo for the pipeline that this worker belongs to
	PipelineInfo() *pps.PipelineInfo

	// Returns the kubernetes namespace that the worker is deployed in
	Namespace() string

	// Returns the path that will contain the input filesets for the job
	InputDir() string

	// Returns the pachd API client for the driver
	PachClient() *client.APIClient

	// Returns the number of workers to be used
	ExpectedNumWorkers() (int64, error)

	// WithContext clones the current driver and applies the context to its
	// pachClient. The pachClient context will be used for other blocking
	// operations as well.
	WithContext(context.Context) Driver

	// WithActiveData swaps the given scratch directory into the 'active' input
	// directory used when running user code. This also locks a mutex so that no
	// two datums can be active concurrently.
	WithActiveData([]*common.Input, string, func() error) error

	// UserCodeEnv returns the set of environment variables to construct when
	// launching the configured user process.
	UserCodeEnv(string, *pfs.Commit, []*common.Input) []string

	RunUserCode(context.Context, logs.TaggedLogger, []string) error

	RunUserErrorHandlingCode(context.Context, logs.TaggedLogger, []string) error

	// TODO: provide a more generic interface for modifying jobs, and
	// some quality-of-life functions for common operations.
	DeleteJob(*sqlx.Tx, *pps.JobInfo) error
	UpdateJobState(*pps.Job, pps.JobState, string) error

	// TODO: figure out how to not expose this - currently only used for a few
	// operations in the map spawner
	NewSQLTx(func(*sqlx.Tx) error) error
}

type driver struct {
	env             serviceenv.ServiceEnv
	ctx             context.Context
	pachClient      *client.APIClient
	pipelineInfo    *pps.PipelineInfo
	activeDataMutex *sync.Mutex

	jobs      col.PostgresCollection
	pipelines col.PostgresCollection

	// User and group IDs used for running user code, determined in the constructor
	uid *uint32
	gid *uint32

	// The root directory to use when setting the user code path. This is normally
	// "/", but can be overridden by tests.
	rootDir string

	// The directory to store input data - this is typically static but can be
	// overridden by tests.
	inputDir string
}

// NewDriver constructs a Driver object using the given clients and pipeline
// settings.  It makes blocking calls to determine the user/group to use with
// the user code on the current worker node, as well as determining if
// enterprise features are activated (for exporting stats).
func NewDriver(
	env serviceenv.ServiceEnv,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	rootPath string,
) (Driver, error) {
	pfsPath := filepath.Join(rootPath, client.PPSInputPrefix)
	if err := os.MkdirAll(pfsPath, 0777); err != nil {
		return nil, errors.EnsureStack(err)
	}
	jobs := ppsdb.Jobs(env.GetDBClient(), env.GetPostgresListener())
	pipelines := ppsdb.Pipelines(env.GetDBClient(), env.GetPostgresListener())
	result := &driver{
		env:             env,
		ctx:             env.Context(),
		pachClient:      pachClient,
		pipelineInfo:    pipelineInfo,
		activeDataMutex: &sync.Mutex{},
		jobs:            jobs,
		pipelines:       pipelines,
		rootDir:         rootPath,
		inputDir:        pfsPath,
	}
	if pipelineInfo.Details.Transform.User != "" {
		user, err := lookupDockerUser(pipelineInfo.Details.Transform.User)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, errors.EnsureStack(err)
		}
		// If `user` is `nil`, `uid` and `gid` will get set, and we won't
		// customize the user that executes the worker process.
		if user != nil { // user is nil when err is an os.ErrNotExist is true in which case we use root
			uid, err := strconv.ParseUint(user.Uid, 10, 32)
			if err != nil {
				return nil, errors.EnsureStack(err)
			}
			uid32 := uint32(uid)
			result.uid = &uid32
			gid, err := strconv.ParseUint(user.Gid, 10, 32)
			if err != nil {
				return nil, errors.EnsureStack(err)
			}
			gid32 := uint32(gid)
			result.gid = &gid32
		}
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
		return nil, errors.EnsureStack(err)
	}
	defer func() {
		if err := passwd.Close(); err != nil && retErr == nil {
			retErr = errors.EnsureStack(err)
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
						return nil, errors.EnsureStack(err)
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
	return nil, errors.Errorf("user %s not found", userArg)
}

func lookupGroup(group string) (_ *user.Group, retErr error) {
	groupFile, err := os.Open("/etc/group")
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer func() {
		if err := groupFile.Close(); err != nil && retErr == nil {
			retErr = errors.EnsureStack(err)
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
	return nil, errors.Errorf("group %s not found", group)
}

func (d *driver) WithContext(ctx context.Context) Driver {
	result := &driver{}
	*result = *d
	result.ctx = ctx
	return result
}

func (d *driver) Jobs() col.PostgresCollection {
	return d.jobs
}

func (d *driver) Pipelines() col.PostgresCollection {
	return d.pipelines
}

func (d *driver) NewTaskWorker() *work.Worker {
	return work.NewWorker(d.env.GetEtcdClient(), d.env.Config().PPSEtcdPrefix, WorkNamespace(d.pipelineInfo))
}

func (d *driver) NewTaskQueue() (*work.TaskQueue, error) {
	return work.NewTaskQueue(d.ctx, d.env.GetEtcdClient(), d.env.Config().PPSEtcdPrefix, WorkNamespace(d.pipelineInfo))
}

func (d *driver) ExpectedNumWorkers() (int64, error) {
	latestPipelineInfo := &pps.PipelineInfo{}
	if err := d.Pipelines().ReadOnly(d.ctx).Get(d.PipelineInfo().Pipeline.Name, latestPipelineInfo); err != nil {
		return 0, errors.EnsureStack(err)
	}
	numWorkers := latestPipelineInfo.Parallelism
	if numWorkers == 0 {
		numWorkers = 1
	}
	return int64(numWorkers), nil
}

func (d *driver) PipelineInfo() *pps.PipelineInfo {
	return d.pipelineInfo
}

func (d *driver) Namespace() string {
	return d.env.Config().Namespace
}

func (d *driver) InputDir() string {
	return d.inputDir
}

func (d *driver) PachClient() *client.APIClient {
	return d.pachClient.WithCtx(d.ctx)
}

func (d *driver) NewSQLTx(cb func(*sqlx.Tx) error) error {
	return col.NewSQLTx(d.ctx, d.env.GetDBClient(), cb)
}

func (d *driver) RunUserCode(
	ctx context.Context,
	logger logs.TaggedLogger,
	environ []string,
) (retErr error) {
	logger.Logf("beginning to run user code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user code after %v", time.Since(start))
		}
	}(time.Now())
	if len(d.pipelineInfo.Details.Transform.Cmd) == 0 {
		return errors.New("invalid pipeline transform, no command specified")
	}

	// Run user code
	cmd := exec.CommandContext(ctx, d.pipelineInfo.Details.Transform.Cmd[0], d.pipelineInfo.Details.Transform.Cmd[1:]...)
	if d.pipelineInfo.Details.Transform.Stdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(d.pipelineInfo.Details.Transform.Stdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	if d.uid != nil && d.gid != nil {
		cmd.SysProcAttr = makeCmdCredentials(*d.uid, *d.gid)
	}

	// By default PWD will be the working dir for the container, so we don't need to set Dir explicitly.
	// If the pipeline or worker config explicitly sets the value, then override the container working dir.
	if d.pipelineInfo.Details.Transform.WorkingDir != "" || d.rootDir != "/" {
		cmd.Dir = filepath.Join(d.rootDir, d.pipelineInfo.Details.Transform.WorkingDir)
	}
	err := cmd.Start()
	if err != nil {
		return errors.EnsureStack(err)
	}
	// A context with a deadline will successfully cancel/kill
	// the running process (minus zombies)
	state, err := cmd.Process.Wait()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if common.IsDone(ctx) {
		if err = ctx.Err(); err != nil {
			return errors.EnsureStack(err)
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
		exiterr := &exec.ExitError{}
		if errors.As(err, &exiterr) {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range d.pipelineInfo.Details.Transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						return nil
					}
				}
			}
		}
		return errors.EnsureStack(err)
	}
	return nil
}

func (d *driver) RunUserErrorHandlingCode(
	ctx context.Context,
	logger logs.TaggedLogger,
	environ []string,
) (retErr error) {
	logger.Logf("beginning to run user error handling code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user error handling code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user error handling code after %v", time.Since(start))
		}
	}(time.Now())

	cmd := exec.CommandContext(ctx, d.pipelineInfo.Details.Transform.ErrCmd[0], d.pipelineInfo.Details.Transform.ErrCmd[1:]...)
	if d.pipelineInfo.Details.Transform.ErrStdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(d.pipelineInfo.Details.Transform.ErrStdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	if d.uid != nil && d.gid != nil {
		cmd.SysProcAttr = makeCmdCredentials(*d.uid, *d.gid)
	}
	cmd.Dir = d.pipelineInfo.Details.Transform.WorkingDir
	err := cmd.Start()
	if err != nil {
		return errors.EnsureStack(err)
	}
	// A context w a deadline will successfully cancel/kill
	// the running process (minus zombies)
	state, err := cmd.Process.Wait()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if common.IsDone(ctx) {
		if err = ctx.Err(); err != nil {
			return errors.EnsureStack(err)
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
		exiterr := &exec.ExitError{}
		if errors.As(err, &exiterr) {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range d.pipelineInfo.Details.Transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						return nil
					}
				}
			}
		}
		return errors.EnsureStack(err)
	}
	return nil
}

func (d *driver) UpdateJobState(job *pps.Job, state pps.JobState, reason string) error {
	return d.NewSQLTx(func(sqlTx *sqlx.Tx) error {
		jobInfo := &pps.JobInfo{}
		if err := d.Jobs().ReadWrite(sqlTx).Get(ppsdb.JobKey(job), jobInfo); err != nil {
			return err
		}
		return errors.EnsureStack(ppsutil.UpdateJobState(d.Pipelines().ReadWrite(sqlTx), d.Jobs().ReadWrite(sqlTx), jobInfo, state, reason))
	})
}

// DeleteJob is identical to updateJobState, except that jobInfo points to a job
// that should be deleted rather than marked failed.  Jobs may be deleted if
// their output commit is deleted.
func (d *driver) DeleteJob(sqlTx *sqlx.Tx, jobInfo *pps.JobInfo) error {
	pipelineInfo := &pps.PipelineInfo{}
	if err := d.Pipelines().ReadWrite(sqlTx).Update(jobInfo.Job.Pipeline.Name, pipelineInfo, func() error {
		if pipelineInfo.JobCounts == nil {
			pipelineInfo.JobCounts = make(map[int32]int32)
		}
		if pipelineInfo.JobCounts[int32(jobInfo.State)] != 0 {
			pipelineInfo.JobCounts[int32(jobInfo.State)]--
		}
		return nil
	}); err != nil {
		return err
	}
	return d.Jobs().ReadWrite(sqlTx).Delete(ppsdb.JobKey(jobInfo.Job))
}

func (d *driver) unlinkData(inputs []*common.Input) error {
	entries, err := ioutil.ReadDir(d.InputDir())
	if err != nil {
		return errors.EnsureStack(err)
	}
	for _, entry := range entries {
		if entry.Name() == client.PPSScratchSpace {
			continue // don't delete scratch space
		}
		if err := os.RemoveAll(filepath.Join(d.InputDir(), entry.Name())); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func (d *driver) UserCodeEnv(
	jobID string,
	outputCommit *pfs.Commit,
	inputs []*common.Input,
) []string {
	result := os.Environ()

	for _, input := range inputs {
		result = append(result, fmt.Sprintf("%s=%s", input.Name, filepath.Join(d.InputDir(), input.Name, input.FileInfo.File.Path)))
		result = append(result, fmt.Sprintf("%s_COMMIT=%s", input.Name, input.FileInfo.File.Commit.ID))
	}

	if jobID != "" {
		result = append(result, fmt.Sprintf("%s=%s", client.JobIDEnv, jobID))
		if ppsutil.ContainsS3Inputs(d.PipelineInfo().Details.Input) || d.PipelineInfo().Details.S3Out {
			// TODO(msteffen) Instead of reading S3GATEWAY_PORT directly, worker/main.go
			// should pass its ServiceEnv to worker.NewAPIServer, which should store it
			// in 'a'. However, requiring worker.APIServer to have a ServiceEnv would
			// break the worker.APIServer initialization in newTestAPIServer (in
			// worker/worker_test.go), which uses mock clients but has no good way to
			// mock a ServiceEnv. Once we can create mock ServiceEnvs, we should store
			// a ServiceEnv in worker.APIServer, rewrite newTestAPIServer and
			// NewAPIServer, and then change this code.
			result = append(
				result,
				fmt.Sprintf("S3_ENDPOINT=http://%s.%s:%s",
					ppsutil.SidecarS3GatewayService(jobID),
					d.Namespace(),
					os.Getenv("S3GATEWAY_PORT"),
				),
			)
		}
	}

	if outputCommit != nil {
		result = append(result, fmt.Sprintf("%s=%s", client.OutputCommitIDEnv, outputCommit.ID))
	}

	return result
}

package worker

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/src-d/go-git.v4"
	gitPlumbing "gopkg.in/src-d/go-git.v4/plumbing"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/exec"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	// The maximum number of concurrent download/upload operations
	concurrency = 100
	logBuffer   = 25

	planPrefix        = "/plan"
	chunkPrefix       = "/chunk"
	mergePrefix       = "/merge"
	parentTreeBufSize = 50 * (1 << (10 * 2))
)

var (
	errSpecialFile = errors.New("cannot upload special file")
	statsTagSuffix = "_stats"
)

// APIServer implements the worker API
type APIServer struct {
	pachClient *client.APIClient
	kubeClient *kube.Clientset
	etcdClient *etcd.Client
	etcdPrefix string

	// Information needed to process input data and upload output
	pipelineInfo *pps.PipelineInfo

	// Information attached to log lines
	logMsgTemplate pps.LogMessage

	// The k8s pod name of this worker
	workerName string

	statusMu sync.Mutex

	// The currently running job ID
	jobID string
	// The currently running data
	data []*Input
	// The time we started the currently running
	started time.Time
	// Func to cancel the currently running datum
	cancel func()
	// Stats about the execution of the job
	stats *pps.ProcessStats
	// queueSize is the number of items enqueued
	queueSize int64

	// The total number of workers for this pipeline
	numWorkers int
	// The namespace in which pachyderm is deployed
	namespace string
	// The jobs collection
	jobs col.Collection
	// The pipelines collection
	pipelines col.Collection
	// The progress collection
	plans col.Collection

	// Only one datum can be running at a time because they need to be
	// accessing /pfs, runMu enforces this
	runMu sync.Mutex

	// We only export application statistics if enterprise is enabled
	exportStats bool

	uid uint32
	gid uint32

	// hashtreeStorage is the where we store on disk hashtrees
	hashtreeStorage string
}

type putObjectResponse struct {
	object *pfs.Object
	size   int64
	err    error
}

type taggedLogger struct {
	template     pps.LogMessage
	stderrLog    log.Logger
	marshaler    *jsonpb.Marshaler
	buffer       bytes.Buffer
	putObjClient pfs.ObjectAPI_PutObjectClient
	objSize      int64
	msgCh        chan string
	eg           errgroup.Group
}

// DatumID computes the id for a datum, this value is used in ListDatum and
// InspectDatum.
func (a *APIServer) DatumID(data []*Input) string {
	hash := sha256.New()
	for _, d := range data {
		hash.Write([]byte(d.FileInfo.File.Path))
		hash.Write(d.FileInfo.Hash)
	}
	// InputFileID is a single string id for the data from this input, it's used in logs and in
	// the statsTree
	return hex.EncodeToString(hash.Sum(nil))
}

func (a *APIServer) getTaggedLogger(pachClient *client.APIClient, jobID string, data []*Input, enableStats bool) (*taggedLogger, error) {
	result := &taggedLogger{
		template:  a.logMsgTemplate, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
		msgCh:     make(chan string, logBuffer),
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line

	// Add Job ID to log metadata
	result.template.JobID = jobID

	// Add inputs' details to log metadata, so we can find these logs later
	for _, d := range data {
		result.template.Data = append(result.template.Data, &pps.InputFile{
			Path: d.FileInfo.File.Path,
			Hash: d.FileInfo.Hash,
		})
	}
	// InputFileID is a single string id for the data from this input, it's used in logs and in
	// the statsTree
	result.template.DatumID = a.DatumID(data)
	if enableStats {
		putObjClient, err := pachClient.ObjectAPIClient.PutObject(pachClient.Ctx())
		if err != nil {
			return nil, err
		}
		result.putObjClient = putObjClient
		result.eg.Go(func() error {
			for msg := range result.msgCh {
				for _, chunk := range grpcutil.Chunk([]byte(msg), grpcutil.MaxMsgSize/2) {
					if err := result.putObjClient.Send(&pfs.PutObjectRequest{
						Value: chunk,
					}); err != nil && err != io.EOF {
						return err
					}
				}
				result.objSize += int64(len(msg))
			}
			return nil
		})
	}
	return result, nil
}

// Logf logs the line Sprintf(formatString, args...), but formatted as a json
// message and annotated with all of the metadata stored in 'loginfo'.
//
// Note: this is not thread-safe, as it modifies fields of 'logger.template'
func (logger *taggedLogger) Logf(formatString string, args ...interface{}) {
	logger.template.Message = fmt.Sprintf(formatString, args...)
	if ts, err := types.TimestampProto(time.Now()); err == nil {
		logger.template.Ts = ts
	} else {
		logger.stderrLog.Printf("could not generate logging timestamp: %s\n", err)
		return
	}
	msg, err := logger.marshaler.MarshalToString(&logger.template)
	if err != nil {
		logger.stderrLog.Printf("could not marshal %v for logging: %s\n", &logger.template, err)
		return
	}
	fmt.Println(msg)
	if logger.putObjClient != nil {
		logger.msgCh <- msg + "\n"
	}
}

func (logger *taggedLogger) Write(p []byte) (_ int, retErr error) {
	// never errors
	logger.buffer.Write(p)
	r := bufio.NewReader(&logger.buffer)
	for {
		message, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.buffer.Write([]byte(message))
				return len(p), nil
			}
			// this shouldn't technically be possible to hit io.EOF should be
			// the only error bufio.Reader can return when using a buffer.
			return 0, fmt.Errorf("error ReadString: %v", err)
		}
		// We don't want to make this call as:
		// logger.Logf(message)
		// because if the message has format characters like %s in it those
		// will result in errors being logged.
		logger.Logf("%s", strings.TrimSuffix(message, "\n"))
	}
}

func (logger *taggedLogger) Close() (*pfs.Object, int64, error) {
	close(logger.msgCh)
	if logger.putObjClient != nil {
		if err := logger.eg.Wait(); err != nil {
			return nil, 0, err
		}
		object, err := logger.putObjClient.CloseAndRecv()
		// we set putObjClient to nil so that future calls to Logf won't send
		// msg down logger.msgCh as we've just closed that channel.
		logger.putObjClient = nil
		return object, logger.objSize, err
	}
	return nil, 0, nil
}

func (logger *taggedLogger) clone() *taggedLogger {
	return &taggedLogger{
		template:     logger.template, // Copy struct
		stderrLog:    log.Logger{},
		marshaler:    &jsonpb.Marshaler{},
		putObjClient: logger.putObjClient,
		msgCh:        logger.msgCh,
	}
}

func (logger *taggedLogger) userLogger() *taggedLogger {
	result := logger.clone()
	result.template.User = true
	return result
}

// NewAPIServer creates an APIServer for a given pipeline
func NewAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, etcdPrefix string, pipelineInfo *pps.PipelineInfo, workerName string, namespace string, hashtreeStorage string) (*APIServer, error) {
	initPrometheus()
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kube.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	server := &APIServer{
		pachClient:   pachClient,
		kubeClient:   kubeClient,
		etcdClient:   etcdClient,
		etcdPrefix:   etcdPrefix,
		pipelineInfo: pipelineInfo,
		logMsgTemplate: pps.LogMessage{
			PipelineName: pipelineInfo.Pipeline.Name,
			WorkerID:     os.Getenv(client.PPSPodNameEnv),
		},
		workerName:      workerName,
		namespace:       namespace,
		jobs:            ppsdb.Jobs(etcdClient, etcdPrefix),
		pipelines:       ppsdb.Pipelines(etcdClient, etcdPrefix),
		plans:           col.NewCollection(etcdClient, path.Join(etcdPrefix, planPrefix), nil, &Plan{}, nil, nil),
		hashtreeStorage: hashtreeStorage,
	}
	logger, err := server.getTaggedLogger(pachClient, "", nil, false)
	if err != nil {
		return nil, err
	}
	resp, err := pachClient.Enterprise.GetState(context.Background(), &enterprise.GetStateRequest{})
	if err != nil {
		logger.Logf("failed to get enterprise state with error: %v\n", err)
	} else {
		server.exportStats = resp.State == enterprise.State_ACTIVE
	}
	numWorkers, err := ppsutil.GetExpectedNumWorkers(kubeClient, pipelineInfo.ParallelismSpec)
	if err != nil {
		logger.Logf("error getting number of workers, default to 1 worker: %v", err)
		numWorkers = 1
	}
	server.numWorkers = numWorkers
	var noDocker bool
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		noDocker = true
	}
	if pipelineInfo.Transform.Image != "" && !noDocker {
		docker, err := docker.NewClientFromEnv()
		if err != nil {
			return nil, err
		}
		image, err := docker.InspectImage(pipelineInfo.Transform.Image)
		if err != nil {
			return nil, fmt.Errorf("error inspecting image %s: %+v", pipelineInfo.Transform.Image, err)
		}
		if pipelineInfo.Transform.User == "" {
			pipelineInfo.Transform.User = image.Config.User
		}
		if pipelineInfo.Transform.WorkingDir == "" {
			pipelineInfo.Transform.WorkingDir = image.Config.WorkingDir
		}
		if server.pipelineInfo.Transform.Cmd == nil {
			server.pipelineInfo.Transform.Cmd = image.Config.Entrypoint
		}
	}
	if pipelineInfo.Transform.User != "" {
		user, err := lookupDockerUser(pipelineInfo.Transform.User)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		// if User == "" then uid, and gid don't get set which
		// means they default to a value of 0 which means we run the code as
		// root which is the only sane default.
		if user != nil { // user is nil when os.IsNotExist(err) is true in which case we use root
			uid, err := strconv.ParseUint(user.Uid, 10, 32)
			if err != nil {
				return nil, err
			}
			server.uid = uint32(uid)
			gid, err := strconv.ParseUint(user.Gid, 10, 32)
			if err != nil {
				return nil, err
			}
			server.gid = uint32(gid)
		}
	}
	if pipelineInfo.Service == nil {
		go server.master()
	} else {
		go server.serviceMaster()
	}
	go server.worker()
	return server, nil
}

func (a *APIServer) downloadGitData(pachClient *client.APIClient, dir string, input *Input) error {
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

func (a *APIServer) reportDownloadSizeStats(downSize float64, logger *taggedLogger) {

	if a.exportStats {
		if hist, err := datumDownloadSize.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(downSize)
		}
		if counter, err := datumDownloadBytesCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(downSize)
		}
	}
}

func (a *APIServer) reportDownloadTimeStats(start time.Time, stats *pps.ProcessStats, logger *taggedLogger) {
	duration := time.Since(start)
	stats.DownloadTime = types.DurationProto(duration)
	if a.exportStats {
		if hist, err := datumDownloadTime.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(duration.Seconds())
		}
		if counter, err := datumDownloadSecondsCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(duration.Seconds())
		}
	}
}

func (a *APIServer) downloadData(pachClient *client.APIClient, logger *taggedLogger, inputs []*Input, puller *filesync.Puller, stats *pps.ProcessStats, statsTree *hashtree.Ordered) (_ string, retErr error) {
	defer a.reportDownloadTimeStats(time.Now(), stats, logger)
	logger.Logf("starting to download data")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored downloading data after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished downloading data after %v", time.Since(start))
		}
	}(time.Now())
	dir := filepath.Join(client.PPSScratchSpace, uuid.NewWithoutDashes())
	// Create output directory (currently /pfs/out)
	if err := os.MkdirAll(filepath.Join(dir, "out"), 0777); err != nil {
		return "", err
	}
	for _, input := range inputs {
		if input.GitURL != "" {
			if err := a.downloadGitData(pachClient, dir, input); err != nil {
				return "", err
			}
			continue
		}
		file := input.FileInfo.File
		root := filepath.Join(dir, input.Name, file.Path)
		var statsRoot string
		if statsTree != nil {
			statsTree.PutDir(input.Name)
			statsRoot = path.Join(input.Name, file.Path)
		}
		if err := puller.Pull(pachClient, root, file.Commit.Repo.Name, file.Commit.ID, file.Path, input.Lazy, input.EmptyFiles, concurrency, statsTree, statsRoot); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func (a *APIServer) linkData(inputs []*Input, dir string) error {
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(client.PPSInputPrefix, input.Name)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return os.Symlink(filepath.Join(dir, "out"), filepath.Join(client.PPSInputPrefix, "out"))
}

func (a *APIServer) unlinkData(inputs []*Input) error {
	for _, input := range inputs {
		if err := os.RemoveAll(filepath.Join(client.PPSInputPrefix, input.Name)); err != nil {
			return err
		}
	}
	return os.RemoveAll(filepath.Join(client.PPSInputPrefix, "out"))
}

func (a *APIServer) reportUserCodeStats(logger *taggedLogger) {
	if a.exportStats {
		if counter, err := datumCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID, "started"); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(1)
		}
	}
}

func (a *APIServer) reportDeferredUserCodeStats(err error, start time.Time, stats *pps.ProcessStats, logger *taggedLogger) {
	duration := time.Since(start)
	stats.ProcessTime = types.DurationProto(duration)
	if a.exportStats {
		state := "errored"
		if err == nil {
			state = "finished"
		}
		if counter, err := datumCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID, state); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) state (%v) with error %v", a.pipelineInfo.ID, a.jobID, state, err)
		} else {
			counter.Add(1)
		}
		if hist, err := datumProcTime.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID, state); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) state (%v) with error %v", a.pipelineInfo.ID, a.jobID, state, err)
		} else {
			hist.Observe(duration.Seconds())
		}
		if counter, err := datumProcSecondsCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(duration.Seconds())
		}
	}
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context, logger *taggedLogger, environ []string, stats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	a.reportUserCodeStats(logger)
	defer func(start time.Time) { a.reportDeferredUserCodeStats(retErr, start, stats, logger) }(time.Now())
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
	cmd := exec.CommandContext(ctx, a.pipelineInfo.Transform.Cmd[0], a.pipelineInfo.Transform.Cmd[1:]...)
	if a.pipelineInfo.Transform.Stdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(a.pipelineInfo.Transform.Stdin, "\n") + "\n")
	}
	cmd.Stdout = logger.userLogger()
	cmd.Stderr = logger.userLogger()
	cmd.Env = environ
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: a.uid,
			Gid: a.gid,
		},
	}
	cmd.Dir = a.pipelineInfo.Transform.WorkingDir
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
	if isDone(ctx) {
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
				for _, returnCode := range a.pipelineInfo.Transform.AcceptReturnCode {
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

func (a *APIServer) reportUploadStats(start time.Time, stats *pps.ProcessStats, logger *taggedLogger) {
	duration := time.Since(start)
	stats.UploadTime = types.DurationProto(duration)
	if a.exportStats {
		if hist, err := datumUploadTime.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(duration.Seconds())
		}
		if counter, err := datumUploadSecondsCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(duration.Seconds())
		}
		if hist, err := datumUploadSize.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(float64(stats.UploadBytes))
		}
		if counter, err := datumUploadBytesCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(float64(stats.UploadBytes))
		}
	}
}

func (a *APIServer) uploadOutput(pachClient *client.APIClient, dir string, tag string, logger *taggedLogger, inputs []*Input, stats *pps.ProcessStats, statsTree *hashtree.Ordered) (retErr error) {
	defer a.reportUploadStats(time.Now(), stats, logger)
	logger.Logf("starting to upload output")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored uploading output after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished uploading output after %v", time.Since(start))
		}
	}(time.Now())
	// Setup client for writing file data
	putObjsClient, err := pachClient.ObjectAPIClient.PutObjects(pachClient.Ctx())
	if err != nil {
		return err
	}
	block := &pfs.Block{Hash: uuid.NewWithoutDashes()}
	if err := putObjsClient.Send(&pfs.PutObjectRequest{
		Block: block,
	}); err != nil {
		return err
	}
	outputPath := filepath.Join(dir, "out")
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	var offset uint64
	var tree *hashtree.Ordered
	// Upload all files in output directory
	if err := filepath.Walk(outputPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filePath == outputPath {
			tree = hashtree.NewOrdered("/")
			return nil
		}
		relPath, err := filepath.Rel(outputPath, filePath)
		if err != nil {
			return err
		}
		// Put directory. Even if the directory is empty, that may be useful to
		// users
		// TODO(msteffen) write a test pipeline that outputs an empty directory and
		// make sure it's preserved
		if info.IsDir() {
			tree.PutDir(relPath)
			if statsTree != nil {
				statsTree.PutDir(relPath)
			}
			return nil
		}
		// Under some circumstances, the user might have copied
		// some pipes from the input directory to the output directory.
		// Reading from these files will result in job blocking.  Thus
		// we preemptively detect if the file is a named pipe.
		if (info.Mode() & os.ModeNamedPipe) > 0 {
			logger.Logf("cannot upload named pipe: %v", relPath)
			return errSpecialFile
		}
		// If the output file is a symlink to an input file, we can skip
		// the uploading.
		if (info.Mode() & os.ModeSymlink) > 0 {
			realPath, err := os.Readlink(filePath)
			if err != nil {
				return err
			}
			if strings.HasPrefix(realPath, client.PPSInputPrefix) {
				var pathWithInput string
				var err error
				if strings.HasPrefix(realPath, dir) {
					pathWithInput, err = filepath.Rel(dir, realPath)
				} else {
					pathWithInput, err = filepath.Rel(client.PPSInputPrefix, realPath)
				}
				if err == nil {
					// We can only skip the upload if the real path is
					// under /pfs, meaning that it's a file that already
					// exists in PFS.

					// The name of the input
					inputName := strings.Split(pathWithInput, string(os.PathSeparator))[0]
					var input *Input
					for _, i := range inputs {
						if i.Name == inputName {
							input = i
						}
					}
					// this changes realPath from `/pfs/input/...` to `/scratch/<id>/input/...`
					realPath = filepath.Join(dir, pathWithInput)
					if input != nil {
						return filepath.Walk(realPath, func(filePath string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							rel, err := filepath.Rel(realPath, filePath)
							if err != nil {
								return err
							}
							subRelPath := filepath.Join(relPath, rel)
							// The path of the input file
							pfsPath, err := filepath.Rel(filepath.Join(dir, input.Name), filePath)
							if err != nil {
								return err
							}
							if info.IsDir() {
								tree.PutDir(subRelPath)
								if statsTree != nil {
									statsTree.PutDir(subRelPath)
								}
								return nil
							}
							fc := input.FileInfo.File.Commit
							fileInfo, err := pachClient.InspectFile(fc.Repo.Name, fc.ID, pfsPath)
							if err != nil {
								return err
							}
							var blockRefs []*pfs.BlockRef
							for _, object := range fileInfo.Objects {
								objectInfo, err := pachClient.InspectObject(object.Hash)
								if err != nil {
									return err
								}
								blockRefs = append(blockRefs, objectInfo.BlockRef)
							}
							blockRefs = append(blockRefs, fileInfo.BlockRefs...)
							n := &hashtree.FileNodeProto{BlockRefs: blockRefs}
							tree.PutFile(subRelPath, fileInfo.Hash, int64(fileInfo.SizeBytes), n)
							if statsTree != nil {
								statsTree.PutFile(subRelPath, fileInfo.Hash, int64(fileInfo.SizeBytes), n)
							}
							return nil
						})
					}
				}
			}
		}
		// Open local file that is being uploaded
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("os.Open(%s): %v", filePath, err)
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		var size int64
		h := pfs.NewHash()
		r := io.TeeReader(f, h)
		// Write local file to object storage block
		for {
			n, err := r.Read(buf)
			if n == 0 && err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if err := putObjsClient.Send(&pfs.PutObjectRequest{
				Value: buf[:n],
			}); err != nil {
				return err
			}
			size += int64(n)
		}
		n := &hashtree.FileNodeProto{
			BlockRefs: []*pfs.BlockRef{
				&pfs.BlockRef{
					Block: block,
					Range: &pfs.ByteRange{
						Lower: offset,
						Upper: offset + uint64(size),
					},
				},
			},
		}
		hash := h.Sum(nil)
		tree.PutFile(relPath, hash, size, n)
		if statsTree != nil {
			statsTree.PutFile(relPath, hash, size, n)
		}
		offset += uint64(size)
		stats.UploadBytes += uint64(size)
		return nil
	}); err != nil {
		return fmt.Errorf("error walking output: %v", err)
	}
	if err := putObjsClient.CloseSend(); err != nil {
		return err
	}
	w, err := pachClient.PutObjectAsync([]*pfs.Tag{client.NewTag(tag)})
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr != nil {
			retErr = err
		}
	}()
	if err := tree.Serialize(w); err != nil {
		return err
	}
	return nil
}

// HashDatum computes and returns the hash of datum + pipeline, with a
// pipeline-specific prefix.
func HashDatum(pipelineName string, pipelineSalt string, data []*Input) string {
	hash := sha256.New()
	for _, datum := range data {
		hash.Write([]byte(datum.Name))
		hash.Write([]byte(datum.FileInfo.File.Path))
		hash.Write(datum.FileInfo.Hash)
	}

	hash.Write([]byte(pipelineName))
	hash.Write([]byte(pipelineSalt))

	return client.DatumTagPrefix(pipelineSalt) + hex.EncodeToString(hash.Sum(nil))
}

// HashDatum15 computes and returns the hash of datum + pipeline for version <= 1.5.0, with a
// pipeline-specific prefix.
func HashDatum15(pipelineInfo *pps.PipelineInfo, data []*Input) (string, error) {
	hash := sha256.New()
	for _, datum := range data {
		hash.Write([]byte(datum.Name))
		hash.Write([]byte(datum.FileInfo.File.Path))
		hash.Write(datum.FileInfo.Hash)
	}

	// We set env to nil because if env contains more than one elements,
	// since it's a map, the output of Marshal() can be non-deterministic.
	env := pipelineInfo.Transform.Env
	pipelineInfo.Transform.Env = nil
	defer func() {
		pipelineInfo.Transform.Env = env
	}()
	bytes, err := pipelineInfo.Transform.Marshal()
	if err != nil {
		return "", err
	}
	hash.Write(bytes)
	hash.Write([]byte(pipelineInfo.Pipeline.Name))
	hash.Write([]byte(pipelineInfo.ID))
	hash.Write([]byte(strconv.Itoa(int(pipelineInfo.Version))))

	// Note in 1.5.0 this function was called HashPipelineID, it's now called
	// HashPipelineName but it has the same implementation.
	return client.DatumTagPrefix(pipelineInfo.ID) + hex.EncodeToString(hash.Sum(nil)), nil
}

// Status returns the status of the current worker.
func (a *APIServer) Status(ctx context.Context, _ *types.Empty) (*pps.WorkerStatus, error) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	started, err := types.TimestampProto(a.started)
	if err != nil {
		return nil, err
	}
	result := &pps.WorkerStatus{
		JobID:     a.jobID,
		WorkerID:  a.workerName,
		Started:   started,
		Data:      a.datum(),
		QueueSize: atomic.LoadInt64(&a.queueSize),
	}
	return result, nil
}

// Cancel cancels the currently running datum
func (a *APIServer) Cancel(ctx context.Context, request *CancelRequest) (*CancelResponse, error) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	if request.JobID != a.jobID {
		return &CancelResponse{Success: false}, nil
	}
	if !MatchDatum(request.DataFilters, a.datum()) {
		return &CancelResponse{Success: false}, nil
	}
	a.cancel()
	// clear the status since we're no longer processing this datum
	a.jobID = ""
	a.data = nil
	a.started = time.Time{}
	a.cancel = nil
	return &CancelResponse{Success: true}, nil
}

func (a *APIServer) datum() []*pps.InputFile {
	var result []*pps.InputFile
	for _, datum := range a.data {
		result = append(result, &pps.InputFile{
			Path: datum.FileInfo.File.Path,
			Hash: datum.FileInfo.Hash,
		})
	}
	return result
}

func (a *APIServer) userCodeEnv(jobID string, outputCommitID string, data []*Input) []string {
	result := os.Environ()
	for _, input := range data {
		result = append(result, fmt.Sprintf("%s=%s", input.Name, filepath.Join(client.PPSInputPrefix, input.Name, input.FileInfo.File.Path)))
		result = append(result, fmt.Sprintf("%s_COMMIT=%s", input.Name, input.FileInfo.File.Commit.ID))
	}
	result = append(result, fmt.Sprintf("%s=%s", client.JobIDEnv, jobID))
	result = append(result, fmt.Sprintf("%s=%s", client.OutputCommitIDEnv, outputCommitID))
	return result
}

// deleteJob is identical to updateJobState, except that jobPtr points to a job
// that should be deleted rather than marked failed. Jobs may be deleted if
// their output commit is deleted.
func (a *APIServer) deleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	pipelinePtr := &pps.EtcdPipelineInfo{}
	if err := a.pipelines.ReadWrite(stm).Update(jobPtr.Pipeline.Name, pipelinePtr, func() error {
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
	return a.jobs.ReadWrite(stm).Delete(jobPtr.Job.ID)
}

type processResult struct {
	failedDatumID   string
	datumsProcessed int64
	datumsSkipped   int64
	datumsFailed    int64
}

type processFunc func(low, high int64) (*processResult, error)

func (a *APIServer) acquireDatums(ctx context.Context, jobID string, plan *Plan, logger *taggedLogger, process processFunc) error {
	complete := false
	for !complete {
		// func to defer cancel in
		if err := func() error {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			var found bool
			low, high := int64(0), int64(0)
			// we set complete to true and then unset it if we find an incomplete chunk
			complete = true
			for _, high = range plan.Chunks {
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					found = false
					chunks := a.chunks(jobID).ReadWrite(stm)
					var chunkState ChunkState
					if err := chunks.Get(fmt.Sprint(high), &chunkState); err != nil {
						if col.IsErrNotFound(err) {
							found = true
						} else {
							return err
						}
					}
					// This gets triggered either if we found a chunk that wasn't
					// complete or if we didn't find a chunk at all.
					if chunkState.State == State_RUNNING {
						complete = false
					}
					if found {
						return chunks.PutTTL(fmt.Sprint(high), &ChunkState{State: State_RUNNING}, ttl)
					}
					return nil
				}); err != nil {
					return err
				}
				if found {
					break
				}
				low = high
			}
			if found {
				go func() {
				Renew:
					for {
						select {
						case <-time.After((time.Second * time.Duration(ttl)) / 2):
						case <-ctx.Done():
							break Renew
						}
						if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
							chunks := a.chunks(jobID).ReadWrite(stm)
							var chunkState ChunkState
							if err := chunks.Get(fmt.Sprint(high), &chunkState); err != nil {
								return err
							}
							if chunkState.State == State_RUNNING {
								return chunks.PutTTL(fmt.Sprint(high), &chunkState, ttl)
							}
							return nil
						}); err != nil {
							cancel()
							logger.Logf("failed to renew lock: %v", err)
						}
					}
				}()
				// process the datums in newRange
				processResult, err := process(low, high)
				if err != nil {
					return err
				}

				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobPtr := &pps.EtcdJobInfo{}
					if err := jobs.Update(jobID, jobPtr, func() error {
						jobPtr.DataProcessed += processResult.datumsProcessed
						jobPtr.DataSkipped += processResult.datumsSkipped
						jobPtr.DataFailed += processResult.datumsFailed
						return nil
					}); err != nil {
						return err
					}
					chunks := a.chunks(jobID).ReadWrite(stm)
					if processResult.failedDatumID != "" {
						return chunks.Put(fmt.Sprint(high), &ChunkState{
							State:   State_FAILED,
							DatumID: processResult.failedDatumID,
						})
					}
					return chunks.Put(fmt.Sprint(high), &ChunkState{State: State_COMPLETE})
				}); err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}

func (a *APIServer) mergeDatums(ctx context.Context, pachClient *client.APIClient, jobInfo *pps.JobInfo, jobID string, plan *Plan, logger *taggedLogger, tags []*pfs.Tag, useParentHashTree bool) error {
	complete := false
	for !complete {
		// func to defer cancel in
		if err := func() (retErr error) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			var merge int64
			var found bool
			if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				// Reinitialize closed upon variables.
				found = false
				merges := a.merges(jobID).ReadWrite(stm)
				// we set complete to true and then unset it if we find an incomplete chunk
				complete = true
				for merge = 0; merge < plan.Merges; merge++ {
					var mergeState MergeState
					if err := merges.Get(fmt.Sprint(merge), &mergeState); err != nil {
						if col.IsErrNotFound(err) {
							found = true
						} else {
							return err
						}
					}
					// This gets triggered either if we found a chunk that wasn't
					// complete or if we didn't find a chunk at all.
					if mergeState.State == State_RUNNING {
						complete = false
					}
					if found {
						break
					}
				}
				if found {
					return merges.PutTTL(fmt.Sprint(merge), &MergeState{State: State_RUNNING}, ttl)
				}
				return nil
			}); err != nil {
				return err
			}
			if found {
				go func() {
				Renew:
					for {
						select {
						case <-time.After((time.Second * time.Duration(ttl)) / 2):
						case <-ctx.Done():
							break Renew
						}
						if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
							merges := a.merges(jobID).ReadWrite(stm)
							var mergeState MergeState
							if err := merges.Get(fmt.Sprint(merge), &mergeState); err != nil {
								return err
							}
							if mergeState.State == State_RUNNING {
								return merges.PutTTL(fmt.Sprint(merge), &mergeState, ttl)
							}
							return nil
						}); err != nil {
							cancel()
							logger.Logf("failed to renew lock: %v", err)
						}
					}
				}()
				// Merge
				objClient, err := obj.NewClientFromEnv(ctx, a.hashtreeStorage)
				if err != nil {
					return err
				}
				var tree *pfs.Object
				var size uint64
				if jobInfo.DataFailed == 0 {
					rs, err := a.getHashtrees(ctx, pachClient, objClient, tags, hashtree.NewFilter(plan.Merges, merge))
					if err != nil {
						return err
					}
					if useParentHashTree {
						r, err := a.getParentHashTree(ctx, pachClient, objClient, jobInfo.OutputCommit, merge)
						if err != nil {
							return err
						}
						defer func() {
							if err := r.Close(); err != nil && retErr == nil {
								retErr = err
							}
						}()
						rs = append([]*hashtree.Reader{hashtree.NewReader(bufio.NewReaderSize(r, parentTreeBufSize), nil)}, rs...)
					}
					tree, size, err = a.merge(pachClient, objClient, nil, rs)
					if err != nil {
						return err
					}
				}
				// Merge stats
				var statsTree *pfs.Object
				var statsSize uint64
				if a.pipelineInfo.EnableStats {
					var statsTags []*pfs.Tag
					for _, tag := range tags {
						statsTags = append(statsTags, client.NewTag(tag.Name+statsTagSuffix))
					}
					rs, err := a.getHashtrees(ctx, pachClient, objClient, statsTags, hashtree.NewFilter(plan.Merges, merge))
					if err != nil {
						return err
					}
					if useParentHashTree {
						r, err := a.getParentHashTree(ctx, pachClient, objClient, jobInfo.StatsCommit, merge)
						if err != nil {
							return err
						}
						defer func() {
							if err := r.Close(); err != nil && retErr == nil {
								retErr = err
							}
						}()
						rs = append([]*hashtree.Reader{hashtree.NewReader(bufio.NewReaderSize(r, parentTreeBufSize), nil)}, rs...)
					}
					statsTree, statsSize, err = a.merge(pachClient, objClient, nil, rs)
					if err != nil {
						return err
					}
				}
				// Finish commit
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					merges := a.merges(jobID).ReadWrite(stm)
					return merges.Put(fmt.Sprint(merge), &MergeState{State: State_COMPLETE, Tree: tree, SizeBytes: size, StatsTree: statsTree, StatsSizeBytes: statsSize})
				}); err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil

}

func (a *APIServer) merge(pachClient *client.APIClient, objClient obj.Client, tags []*pfs.Tag, rs []*hashtree.Reader) (*pfs.Object, uint64, error) {
	var tree *pfs.Object
	var size uint64
	if err := func() error {
		objW, err := pachClient.PutObjectAsync(tags)
		if err != nil {
			return err
		}
		w := hashtree.NewWriter(objW)
		size, err = hashtree.Merge(w, rs)
		if err != nil {
			return err
		}
		// Get object hash for hashtree
		// (bryce) make this deferred
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

func (a *APIServer) getParentCommitInfo(ctx context.Context, pachClient *client.APIClient, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	commitInfo, err := pachClient.PfsAPIClient.InspectCommit(ctx,
		&pfs.InspectCommitRequest{
			Commit: commit,
		})
	if err != nil {
		return nil, err
	}
	for commitInfo.ParentCommit != nil {
		parentCommitInfo, err := pachClient.PfsAPIClient.InspectCommit(ctx,
			&pfs.InspectCommitRequest{
				Commit:     commitInfo.ParentCommit,
				BlockState: pfs.CommitState_FINISHED,
			})
		if err != nil {
			return nil, err
		}
		if parentCommitInfo.Trees != nil {
			return parentCommitInfo, nil
		}
		commitInfo = parentCommitInfo
	}
	return nil, nil
}

func (a *APIServer) getHashtrees(ctx context.Context, pachClient *client.APIClient, objClient obj.Client, tags []*pfs.Tag, filter func(k []byte) (bool, error)) ([]*hashtree.Reader, error) {
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
			objR, err := objClient.Reader(path, 0, 0)
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

func (a *APIServer) getCommitDatums(ctx context.Context, pachClient *client.APIClient, commitInfo *pfs.CommitInfo) (datums map[string]struct{}, retErr error) {
	r, err := pachClient.GetObjectReader(commitInfo.Datums.Hash)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil && retErr != nil {
			retErr = err
		}
	}()
	pbr := pbutil.NewReader(r)
	datums = make(map[string]struct{})
	for {
		k, err := pbr.ReadBytes()
		if err != nil {
			if err == io.EOF {
				return
			}
			return nil, err
		}
		datums[string(k)] = struct{}{}
	}
}

func (a *APIServer) getParentHashTree(ctx context.Context, pachClient *client.APIClient, objClient obj.Client, commit *pfs.Commit, merge int64) (io.ReadCloser, error) {
	parentCommitInfo, err := a.getParentCommitInfo(ctx, pachClient, commit)
	if err != nil {
		return nil, err
	}
	info, err := pachClient.InspectObject(parentCommitInfo.Trees[merge].Hash)
	if err != nil {
		return nil, err
	}
	path, err := obj.BlockPathFromEnv(info.BlockRef.Block)
	if err != nil {
		return nil, err
	}
	return objClient.Reader(path, 0, 0)
}

func writeIndex(pachClient *client.APIClient, objClient obj.Client, tree *pfs.Object, idx []byte) (retErr error) {
	info, err := pachClient.InspectObject(tree.Hash)
	if err != nil {
		return err
	}
	path, err := obj.BlockPathFromEnv(info.BlockRef.Block)
	if err != nil {
		return err
	}
	idxW, err := objClient.Writer(path + hashtree.IndexPath)
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

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// cancelCtxIfJobFails watches jobID's JobPtr, and if its state is changed to a
// terminal state (KILLED, FAILED, or SUCCESS) cancel the jobCtx so we kill any
// user processes
func (a *APIServer) cancelCtxIfJobFails(jobCtx context.Context, jobCancel func(), jobID string) {
	logger := a.getWorkerLogger() // this worker's formatting logger

	backoff.RetryNotify(func() error {
		// Check if job was cancelled while backoff was sleeping
		if isDone(jobCtx) {
			return nil
		}

		// Start watching for job state changes
		watcher, err := a.jobs.ReadOnly(jobCtx).WatchOne(jobID)
		if err != nil {
			if col.IsErrNotFound(err) {
				jobCancel() // job deleted before we started--cancel the job ctx
				return nil
			}
			return fmt.Errorf("worker: could not create state watcher for job %s, err is %v", jobID, err)
		}

		// If any job events indicate that the job is done, cancel jobCtx
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
						jobCancel() // cancel the job
					}
				case watch.EventDelete:
					jobCancel() // cancel the job
				case watch.EventError:
					return fmt.Errorf("job state watch error: %v", e.Err)
				}
			case <-jobCtx.Done():
				break
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		if jobCtx.Err() == context.Canceled {
			return err // worker is done, nothing else to do
		}
		logger.Logf("worker: error watching job %s (%v); retrying in %v", jobID, err, d)
		return nil
	})
}

// worker does the following:
//  - watches for new jobs (jobInfos in the jobs collection)
//  - claims chunks from the chunk layout it finds in the chunks collection
//  - claims those chunks with acquireDatums
//  - processes the chunks with processDatums
func (a *APIServer) worker() {
	logger := a.getWorkerLogger() // this worker's formatting logger

	// Process incoming jobs
	backoff.RetryNotify(func() (retErr error) {
		retryCtx, retryCancel := context.WithCancel(a.pachClient.Ctx())
		defer retryCancel()
		watcher, err := a.jobs.ReadOnly(retryCtx).WatchByIndex(ppsdb.JobsPipelineIndex, a.pipelineInfo.Pipeline)
		if err != nil {
			return fmt.Errorf("error creating watch: %v", err)
		}
		defer watcher.Close()
	NextJob:
		for e := range watcher.Watch() {
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
				continue NextJob
			}

			// create new ctx for this job, and don't use retryCtx as the
			// parent. Just because another job's etcd write failed doesn't
			// mean this job shouldn't run
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
					continue NextJob // job was deleted--no sense retrying
				}
				return fmt.Errorf("error from InspectJob(%v): %+v", jobID, err)
			}
			if jobInfo.PipelineVersion < a.pipelineInfo.Version {
				continue
			}
			if jobInfo.PipelineVersion > a.pipelineInfo.Version {
				return fmt.Errorf("job %s's version (%d) greater than pipeline's "+
					"version (%d), this should automatically resolve when the worker "+
					"is updated", jobID, jobInfo.PipelineVersion, a.pipelineInfo.Version)
			}

			// Read the chunks laid out by the master and create the datum factory
			plan := &Plan{}
			if err := a.plans.ReadOnly(jobCtx).GetBlock(jobInfo.Job.ID, plan); err != nil {
				return err
			}
			df, err := NewDatumFactory(pachClient, jobInfo.Input)
			if err != nil {
				return fmt.Errorf("error from NewDatumFactory: %v", err)
			}

			// Compute the datums to skip
			skip := make(map[string]struct{})
			var useParentHashTree bool
			parentCommitInfo, err := a.getParentCommitInfo(jobCtx, pachClient, jobInfo.OutputCommit)
			if err != nil {
				return err
			}
			if parentCommitInfo != nil {
				var err error
				skip, err = a.getCommitDatums(jobCtx, pachClient, parentCommitInfo)
				if err != nil {
					return err
				}
				var count int
				for i := 0; i < df.Len(); i++ {
					files := df.Datum(i)
					datumHash := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, files)
					if _, ok := skip[datumHash]; ok {
						count++
					}
				}
				if len(skip) == count {
					useParentHashTree = true
				}
			}
			// If a datum fails, acquireDatums updates the relevant lock in
			// etcd, which causes the master to fail the job (which is
			// handled above in the JOB_FAILURE case). There's no need to
			// handle failed datums here, just failed etcd writes.
			if err := a.acquireDatums(
				jobCtx, jobID, plan, logger,
				func(low, high int64) (*processResult, error) {
					processResult, err := a.processDatums(pachClient, logger, jobInfo, df, low, high, skip)
					if err != nil {
						return nil, err
					}
					return processResult, nil
				},
			); err != nil {
				if jobCtx.Err() == context.Canceled {
					continue NextJob // job cancelled--don't restart, just wait for next job
				}
				return fmt.Errorf("acquire/process datums for job %s exited with err: %v", jobID, err)
			}
			var tags []*pfs.Tag
			for i := 0; i < df.Len(); i++ {
				files := df.Datum(int(i))
				datumHash := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, files)
				if _, ok := skip[datumHash]; ok && useParentHashTree {
					continue
				}
				tags = append(tags, client.NewTag(datumHash))
			}
			skip = nil
			jobInfo, err = pachClient.InspectJob(jobID, false)
			if err != nil {
				return err
			}
			if err := a.mergeDatums(jobCtx, pachClient, jobInfo, jobID, plan, logger, tags, useParentHashTree); err != nil {
				if jobCtx.Err() == context.Canceled {
					continue NextJob // job cancelled--don't restart, just wait for next job
				}
				return fmt.Errorf("merge datums for job %s exited with err: %v", jobID, err)
			}
		}
		return fmt.Errorf("worker: jobs.WatchByIndex(pipeline = %s) closed unexpectedly", a.pipelineInfo.Pipeline.Name)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("worker: watch closed or error running the worker process: %v; retrying in %v", err, d)
		return nil
	})
}

// processDatums processes datums from low to high in df, if a datum fails it
// returns the id of the failed datum it also may return a variety of errors
// such as network errors.
func (a *APIServer) processDatums(pachClient *client.APIClient, logger *taggedLogger, jobInfo *pps.JobInfo, df DatumFactory, low, high int64, skip map[string]struct{}) (*processResult, error) {
	ctx := pachClient.Ctx()
	objClient, err := obj.NewClientFromEnv(ctx, a.hashtreeStorage)
	if err != nil {
		return nil, err
	}
	stats := &pps.ProcessStats{}
	var statsMu sync.Mutex
	result := &processResult{}
	var eg errgroup.Group
	limiter := limit.New(int(a.pipelineInfo.MaxQueueSize))
	for i := low; i < high; i++ {
		i := i

		limiter.Acquire()
		atomic.AddInt64(&a.queueSize, 1)
		eg.Go(func() (retErr error) {
			defer limiter.Release()
			defer atomic.AddInt64(&a.queueSize, -1)

			data := df.Datum(int(i))
			logger, err := a.getTaggedLogger(pachClient, jobInfo.Job.ID, data, a.pipelineInfo.EnableStats)
			if err != nil {
				return err
			}
			// Hash inputs
			tag := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, data)
			if _, ok := skip[tag]; ok {
				atomic.AddInt64(&result.datumsSkipped, 1)
				logger.Logf("skipping datum")
				return nil
			}
			if _, err := pachClient.InspectTag(ctx, client.NewTag(tag)); err == nil {
				atomic.AddInt64(&result.datumsSkipped, 1)
				logger.Logf("skipping datum")
				return nil
			}
			subStats := &pps.ProcessStats{}
			var inputTree, outputTree *hashtree.Ordered
			var statsTree *hashtree.Unordered
			if a.pipelineInfo.EnableStats {
				statsRoot := path.Join("/", logger.template.DatumID)
				inputTree = hashtree.NewOrdered(path.Join(statsRoot, "pfs"))
				outputTree = hashtree.NewOrdered(path.Join(statsRoot, "pfs", "out"))
				statsTree = hashtree.NewUnordered(statsRoot)
				// Write job id to stats tree
				statsTree.PutFile(fmt.Sprintf("job:%s", jobInfo.Job.ID), nil, 0)
				// Write index in datum factory to stats tree
				object, size, err := pachClient.PutObject(strings.NewReader(fmt.Sprint(int(i))))
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
					if err := a.writeStats(pachClient, objClient, tag, subStats, logger, inputTree, outputTree, statsTree); err != nil && retErr == nil {
						retErr = err
					}
				}()
			}

			env := a.userCodeEnv(jobInfo.Job.ID, jobInfo.OutputCommit.ID, data)
			var dir string
			var failures int64
			if err := backoff.RetryNotify(func() error {
				if isDone(ctx) {
					return ctx.Err() // timeout or cancelled job--don't run datum
				}
				// Download input data
				puller := filesync.NewPuller()
				// TODO parent tag shouldn't be nil
				var err error
				dir, err = a.downloadData(pachClient, logger, data, puller, subStats, inputTree)
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
					return fmt.Errorf("error downloadData: %v", err)
				}
				a.runMu.Lock()
				defer a.runMu.Unlock()
				// shadow ctx and pachClient for the context of processing this one datum
				ctx, cancel := context.WithCancel(ctx)
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
				if err := os.MkdirAll(client.PPSInputPrefix, 0777); err != nil {
					return err
				}
				if err := a.linkData(data, dir); err != nil {
					return fmt.Errorf("error linkData: %v", err)
				}
				defer func() {
					if err := a.unlinkData(data); err != nil && retErr == nil {
						retErr = fmt.Errorf("error unlinkData: %v", err)
					}
				}()
				if a.pipelineInfo.Transform.User != "" {
					filepath.Walk("/pfs", func(name string, info os.FileInfo, err error) error {
						if err == nil {
							err = os.Chown(name, int(a.uid), int(a.gid))
						}
						return err
					})
				}
				if err := a.runUserCode(ctx, logger, env, subStats, jobInfo.DatumTimeout); err != nil {
					return fmt.Errorf("error runUserCode: %v", err)
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
					return err
				}
				atomic.AddUint64(&subStats.DownloadBytes, uint64(downSize))
				a.reportDownloadSizeStats(float64(downSize), logger)
				return a.uploadOutput(pachClient, dir, tag, logger, data, subStats, outputTree)
			}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
				if isDone(ctx) {
					return ctx.Err() // timeout or cancelled job, err out and don't retry
				}
				failures++
				if failures >= jobInfo.DatumTries {
					logger.Logf("failed to process datum with error: %+v", err)
					if statsTree != nil {
						object, size, err := pachClient.PutObject(strings.NewReader(err.Error()))
						if err != nil {
							logger.stderrLog.Printf("could not put error object: %s\n", err)
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
			}); err != nil {
				result.failedDatumID = a.DatumID(data)
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
	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.jobs.ReadWrite(stm)
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
	result.datumsProcessed = high - low - result.datumsSkipped - result.datumsFailed
	return result, nil
}

func (a *APIServer) writeStats(pachClient *client.APIClient, objClient obj.Client, tag string, stats *pps.ProcessStats, logger *taggedLogger, inputTree, outputTree *hashtree.Ordered, statsTree *hashtree.Unordered) error {
	// Store stats and add stats file
	marshaler := &jsonpb.Marshaler{}
	statsString, err := marshaler.MarshalToString(stats)
	if err != nil {
		logger.stderrLog.Printf("could not serialize stats: %s\n", err)
		return err
	}
	object, size, err := pachClient.PutObject(strings.NewReader(statsString))
	if err != nil {
		logger.stderrLog.Printf("could not put stats object: %s\n", err)
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
	objW, err := pachClient.PutObjectAsync([]*pfs.Tag{client.NewTag(tag + statsTagSuffix)})
	if err != nil {
		return err
	}
	defer objW.Close()
	_, err = hashtree.Merge(hashtree.NewWriter(objW), []*hashtree.Reader{
		hashtree.NewReader(inputBuf, nil),
		hashtree.NewReader(outputBuf, nil),
		hashtree.NewReader(statsBuf, nil),
	})
	return err
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

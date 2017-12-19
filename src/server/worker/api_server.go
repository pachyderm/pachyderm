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
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/src-d/go-git.v4"
	gitPlumbing "gopkg.in/src-d/go-git.v4/plumbing"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/exec"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

const (
	// The maximum number of concurrent download/upload operations
	concurrency = 10
	logBuffer   = 25

	chunksPrefix = "/chunks"
	lockPrefix   = "/locks"

	maxRetries = 3
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
	chunks col.Collection

	// Only one datum can be running at a time because they need to be
	// accessing /pfs, runMu enforces this
	runMu sync.Mutex

	// datumCache is used by the master to keep track of the datums that
	// have already been processed.
	datumCache *lru.Cache

	uid        uint32
	gid        uint32
	workingDir string
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

func (a *APIServer) getTaggedLogger(ctx context.Context, jobID string, data []*Input, enableStats bool) (*taggedLogger, error) {
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
		putObjClient, err := a.pachClient.ObjectAPIClient.PutObject(auth.In2Out(ctx))
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
	bytes, err := logger.marshaler.MarshalToString(&logger.template)
	if err != nil {
		logger.stderrLog.Printf("could not marshal %v for logging: %s\n", &logger.template, err)
		return
	}
	bytes += "\n"
	fmt.Printf(bytes)
	if logger.putObjClient != nil {
		logger.msgCh <- bytes
	}
}

func (logger *taggedLogger) Write(p []byte) (_ int, retErr error) {
	// never errors
	logger.buffer.Write(p)
	r := bufio.NewReader(&logger.buffer)
	for {
		message, err := r.ReadString('\n')
		if err != nil {
			message = strings.TrimSuffix(message, "\n") // remove delimiter
			if err == io.EOF {
				logger.buffer.Write([]byte(message))
				return len(p), nil
			}
			// this shouldn't technically be possible to hit io.EOF should be
			// the only error bufio.Reader can return when using a buffer.
			return 0, err
		}
		logger.Logf(message)
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
func NewAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, etcdPrefix string, pipelineInfo *pps.PipelineInfo, workerName string, namespace string) (*APIServer, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kube.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	datumCache, err := lru.New(numCachedDatums)
	if err != nil {
		return nil, fmt.Errorf("error creating datum cache: %v", err)
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
		workerName: workerName,
		namespace:  namespace,
		jobs:       ppsdb.Jobs(etcdClient, etcdPrefix),
		pipelines:  ppsdb.Pipelines(etcdClient, etcdPrefix),
		chunks:     col.NewCollection(etcdClient, path.Join(etcdPrefix, chunksPrefix), []col.Index{}, &Chunks{}, nil),
		datumCache: datumCache,
	}
	logger, err := server.getTaggedLogger(context.Background(), "", nil, false)
	if err != nil {
		return nil, err
	}
	numWorkers, err := ppsutil.GetExpectedNumWorkers(kubeClient, pipelineInfo.ParallelismSpec)
	if err != nil {
		logger.Logf("error getting number of workers, default to 1 worker: %v", err)
		numWorkers = 1
	}
	server.numWorkers = numWorkers
	if pipelineInfo.Transform.Image != "" {
		docker, err := docker.NewClientFromEnv()
		if err != nil {
			return nil, err
		}
		image, err := docker.InspectImage(pipelineInfo.Transform.Image)
		if err != nil {
			return nil, fmt.Errorf("error inspecting image %s: %+v", pipelineInfo.Transform.Image, err)
		}
		// if image.Config.User == "" then uid, and gid don't get set which
		// means they default to a value of 0 which means we run the code as
		// root which is the only sane default.
		if image.Config.User != "" {
			user, err := lookupUser(image.Config.User)
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
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
		server.workingDir = image.Config.WorkingDir
		if server.pipelineInfo.Transform.Cmd == nil {
			server.pipelineInfo.Transform.Cmd = image.Config.Entrypoint
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

func (a *APIServer) downloadData(logger *taggedLogger, inputs []*Input, puller *filesync.Puller, parentTag *pfs.Tag, stats *pps.ProcessStats, statsTree hashtree.OpenHashTree, statsPath string) (string, error) {
	defer func(start time.Time) {
		stats.DownloadTime = types.DurationProto(time.Since(start))
	}(time.Now())
	logger.Logf("input has not been processed, downloading data")
	defer func(start time.Time) {
		logger.Logf("input data download took (%v)", time.Since(start))
	}(time.Now())
	dir := filepath.Join(client.PPSScratchSpace, uuid.NewWithoutDashes())
	for _, input := range inputs {
		file := input.FileInfo.File
		root := filepath.Join(dir, input.Name, file.Path)
		treeRoot := path.Join(statsPath, input.Name, file.Path)
		if a.pipelineInfo.Incremental && input.ParentCommit != nil {
			if err := puller.PullDiff(a.pachClient, root,
				file.Commit.Repo.Name, file.Commit.ID, file.Path,
				input.ParentCommit.Repo.Name, input.ParentCommit.ID, file.Path,
				true, input.Lazy, concurrency, statsTree, treeRoot); err != nil {
				return "", err
			}
		}
		if input.GitURL != "" {
			pachydermRepoName := input.Name
			var rawJSON bytes.Buffer
			err := a.pachClient.GetFile(pachydermRepoName, file.Commit.ID, "commit.json", 0, 0, &rawJSON)
			if err != nil {
				return "", err
			}
			var payload github.PushPayload
			err = json.Unmarshal(rawJSON.Bytes(), &payload)
			if err != nil {
				return "", err
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
				return "", fmt.Errorf("error cloning repo %v at SHA %v from URL %v: %v", pachydermRepoName, sha, input.GitURL, err)
			}
			hash := gitPlumbing.NewHash(sha)
			wt, err := r.Worktree()
			if err != nil {
				return "", err
			}
			err = wt.Checkout(&git.CheckoutOptions{Hash: hash})
			if err != nil {
				return "", fmt.Errorf("error checking out SHA %v from repo %v: %v", sha, pachydermRepoName, err)
			}
		} else {
			if err := puller.Pull(a.pachClient, root, file.Commit.Repo.Name, file.Commit.ID, file.Path, input.Lazy, concurrency, statsTree, treeRoot); err != nil {
				return "", err
			}
		}
	}
	if parentTag != nil {
		var buffer bytes.Buffer
		if err := a.pachClient.GetTag(parentTag.Name, &buffer); err != nil {
			return "", fmt.Errorf("error getting parent for datum %v: %v", inputs, err)
		}
		tree, err := hashtree.Deserialize(buffer.Bytes())
		if err != nil {
			return "", fmt.Errorf("failed to deserialize parent hashtree: %v", err)
		}
		if err := puller.PullTree(a.pachClient, path.Join(dir, "out"), tree, false, concurrency); err != nil {
			return "", fmt.Errorf("error pulling output tree: %+v", err)
		}
	}
	return dir, nil
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context, logger *taggedLogger, environ []string, stats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {

	defer func(start time.Time) {
		stats.ProcessTime = types.DurationProto(time.Since(start))
	}(time.Now())
	logger.Logf("beginning to run user code")
	defer func(start time.Time) {
		logger.Logf("finished running user code - took (%v) - with error (%v)", time.Since(start), retErr)
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
	cmd.Stdin = strings.NewReader(strings.Join(a.pipelineInfo.Transform.Stdin, "\n") + "\n")
	cmd.Stdout = logger.userLogger()
	cmd.Stderr = logger.userLogger()
	cmd.Env = environ
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: a.uid,
			Gid: a.gid,
		},
	}
	cmd.Dir = a.workingDir
	err := cmd.Start()
	if err != nil {
		return err
	}
	// A context w a deadline will successfully cancel/kill
	// the running process (minus zombies)
	state, err := cmd.Process.Wait()
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		if err = ctx.Err(); err != nil {
			return err
		}
	default:
	}

	// Because of this issue: https://github.com/golang/go/issues/18874
	// We forked os/exec so that we can call just the part of cmd.Wait() that
	// happens after blocking on the process. Unfortunately calling
	// cmd.Process.Wait() then cmd.Wait() will produce an error. So instead we
	// close the IO using this helper
	err = cmd.WaitIO(state, err)
	if err != nil {
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
		return err
	}
	return nil
}

func (a *APIServer) uploadOutput(ctx context.Context, dir string, tag string, logger *taggedLogger, inputs []*Input, stats *pps.ProcessStats, statsTree hashtree.OpenHashTree, statsRoot string) error {
	defer func(start time.Time) {
		stats.UploadTime = types.DurationProto(time.Since(start))
	}(time.Now())
	logger.Logf("starting to upload output")
	defer func(start time.Time) {
		logger.Logf("finished uploading output - took %v", time.Since(start))
	}(time.Now())
	// hashtree is not thread-safe--guard with 'lock'
	var lock sync.Mutex
	tree := hashtree.NewHashTree()
	outputPath := filepath.Join(dir, "out")

	// Upload all files in output directory
	var g errgroup.Group
	limiter := limit.New(concurrency)
	if err := filepath.Walk(outputPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		g.Go(func() (retErr error) {
			limiter.Acquire()
			defer limiter.Release()
			if filePath == outputPath {
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
				lock.Lock()
				defer lock.Unlock()
				tree.PutDir(relPath)
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
				pathWithInput, err := filepath.Rel(client.PPSInputPrefix, realPath)
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
								lock.Lock()
								defer lock.Unlock()
								tree.PutDir(subRelPath)
								return nil
							}

							fileInfo, err := a.pachClient.PfsAPIClient.InspectFile(auth.In2Out(ctx), &pfs.InspectFileRequest{
								File: &pfs.File{
									Commit: input.FileInfo.File.Commit,
									Path:   pfsPath,
								},
							})
							if err != nil {
								return err
							}

							lock.Lock()
							defer lock.Unlock()
							if statsTree != nil {
								if err := statsTree.PutFile(path.Join(statsRoot, subRelPath), fileInfo.Objects, int64(fileInfo.SizeBytes)); err != nil {
									return err
								}
							}
							return tree.PutFile(subRelPath, fileInfo.Objects, int64(fileInfo.SizeBytes))
						})
					}
				}
			}

			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()

			putObjClient, err := a.pachClient.ObjectAPIClient.PutObject(auth.In2Out(ctx))
			if err != nil {
				return err
			}
			size, err := grpcutil.ChunkReader(f, func(chunk []byte) error {
				return putObjClient.Send(&pfs.PutObjectRequest{
					Value: chunk,
				})
			})
			if err != nil {
				return err
			}
			object, err := putObjClient.CloseAndRecv()
			if err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()
			atomic.AddUint64(&stats.UploadBytes, uint64(size))
			if statsTree != nil {
				if err := statsTree.PutFile(path.Join(statsRoot, relPath), []*pfs.Object{object}, int64(size)); err != nil {
					return err
				}
			}
			return tree.PutFile(relPath, []*pfs.Object{object}, int64(size))
		})
		return nil
	}); err != nil {
		return err
	}

	if err := g.Wait(); err != nil {
		return err
	}

	finTree, err := tree.Finish()
	if err != nil {
		return err
	}

	treeBytes, err := hashtree.Serialize(finTree)
	if err != nil {
		return err
	}

	if _, _, err := a.pachClient.PutObject(bytes.NewReader(treeBytes), tag); err != nil {
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
		QueueSize: a.queueSize,
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

func (a *APIServer) userCodeEnv(jobID string, data []*Input) []string {
	result := os.Environ()
	for _, input := range data {
		result = append(result, fmt.Sprintf("%s=%s", input.Name, filepath.Join(client.PPSInputPrefix, input.Name, input.FileInfo.File.Path)))
	}
	result = append(result, fmt.Sprintf("PACH_JOB_ID=%s", jobID))
	return result
}

func (a *APIServer) updateJobState(stm col.STM, jobInfo *pps.JobInfo, state pps.JobState, reason string) error {
	// Update job counts
	if jobInfo.Pipeline != nil {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelinePtr := &pps.EtcdPipelineInfo{}
		if err := pipelines.Get(jobInfo.Pipeline.Name, pipelinePtr); err != nil {
			return err
		}
		if pipelinePtr.JobCounts == nil {
			pipelinePtr.JobCounts = make(map[int32]int32)
		}
		if pipelinePtr.JobCounts[int32(jobInfo.State)] != 0 {
			pipelinePtr.JobCounts[int32(jobInfo.State)]--
		}
		pipelinePtr.JobCounts[int32(state)]++
		pipelines.Put(jobInfo.Pipeline.Name, pipelinePtr)
	}
	jobInfo.State = state
	jobInfo.Reason = reason
	jobs := a.jobs.ReadWrite(stm)
	jobs.Put(jobInfo.Job.ID, jobInfo)
	return nil
}

type acquireDatumsFunc func(low, high int64) (failedDatumID string, _ error)

func (a *APIServer) acquireDatums(ctx context.Context, jobID string, chunks *Chunks, logger *taggedLogger, process acquireDatumsFunc) error {
	complete := false
	for !complete {
		// func to defer cancel in
		if err := func() error {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			var low int64
			var high int64
			var found bool
			if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				found = false
				locks := a.locks(jobID).ReadWrite(stm)
				// we set complete to true and then unset it if we find an incomplete chunk
				complete = true
				for _, high = range chunks.Chunks {
					var chunkState ChunkState
					if err := locks.Get(fmt.Sprint(high), &chunkState); err != nil {
						if col.IsErrNotFound(err) {
							found = true
						} else {
							return err
						}
					}
					// This gets triggered either if we found a chunk that wasn't
					// complete or if we didn't find a chunk at all.
					if chunkState.State == ChunkState_RUNNING {
						complete = false
					}
					if found {
						break
					}
					low = high
				}
				if found {
					return locks.PutTTL(fmt.Sprint(high), &ChunkState{State: ChunkState_RUNNING}, ttl)
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
							locks := a.locks(jobID).ReadWrite(stm)
							var chunkState ChunkState
							if err := locks.Get(fmt.Sprint(high), &chunkState); err != nil {
								return err
							}
							if chunkState.State == ChunkState_RUNNING {
								return locks.PutTTL(fmt.Sprint(high), &chunkState, ttl)
							}
							return nil
						}); err != nil {
							cancel()
							logger.Logf("failed to renew lock: %v", err)
						}
					}
				}()
				// process the datums in newRange
				failedDatumID, err := process(low, high)
				if err != nil {
					return err
				}

				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					locks := a.locks(jobID).ReadWrite(stm)
					if failedDatumID != "" {
						return locks.Put(fmt.Sprint(high), &ChunkState{
							State:   ChunkState_FAILED,
							DatumID: failedDatumID,
						})
					}
					return locks.Put(fmt.Sprint(high), &ChunkState{State: ChunkState_COMPLETE})
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

// worker does the following:
//  - watches for new jobs (jobInfos in the jobs collection)
//  - claims chunks from the chunk layout it finds in the chunks collection
//  - claims those chunks with acquireDatums
//  - processes the chunks with processDatums
func (a *APIServer) worker() {
	logger := a.getWorkerLogger()
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(a.pachClient.AddMetadata(context.Background()))
		defer cancel()
		jobs := a.jobs.ReadOnly(ctx)
		watcher, err := jobs.WatchByIndex(ppsdb.JobsPipelineIndex, a.pipelineInfo.Pipeline)
		if err != nil {
			return fmt.Errorf("error creating watch: %v", err)
		}
		defer watcher.Close()
		for e := range watcher.Watch() {
			var jobID string
			jobInfo := &pps.JobInfo{}
			if err := e.Unmarshal(&jobID, jobInfo); err != nil {
				return fmt.Errorf("error unmarshalling: %v", err)
			}
			chunks := &Chunks{}
			if err := a.chunks.ReadOnly(ctx).GetBlock(jobInfo.Job.ID, chunks); err != nil {
				return err
			}
			df, err := NewDatumFactory(ctx, a.pachClient.PfsAPIClient, jobInfo.Input)
			if err != nil {
				return fmt.Errorf("error from NewDatumFactory: %v", err)
			}
			if err := a.acquireDatums(ctx, jobInfo.Job.ID, chunks, logger, func(low, high int64) (string, error) {
				failedDatumID, err := a.processDatums(ctx, logger, jobInfo, df, low, high)
				if err != nil {
					return "", err
				}
				return failedDatumID, nil
			}); err != nil {
				return fmt.Errorf("error from acquireDatums: %v", err)
			}
		}
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("worker: error running the worker process: %v; retrying in %v", err, d)
		return nil
	})
}

// processDatums processes datums from low to high in df, if a datum fails it
// returns the id of the failed datum it also may return a variety of errors
// such as network errors.
func (a *APIServer) processDatums(ctx context.Context, logger *taggedLogger, jobInfo *pps.JobInfo, df DatumFactory, low, high int64) (string, error) {
	stats := &pps.ProcessStats{}
	var statsMu sync.Mutex
	var failedDatumID string
	var eg errgroup.Group
	var skipped int64
	var failed int64
	for i := low; i < high; i++ {
		i := i
		eg.Go(func() (retErr error) {
			data := df.Datum(int(i))
			logger, err := a.getTaggedLogger(ctx, jobInfo.Job.ID, data, a.pipelineInfo.EnableStats)
			if err != nil {
				return err
			}
			atomic.AddInt64(&a.queueSize, 1)
			// Hash inputs
			tag := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, data)
			tag15, err := HashDatum15(a.pipelineInfo, data)
			if err != nil {
				return err
			}
			foundTag := false
			foundTag15 := false
			var object *pfs.Object
			var eg errgroup.Group
			eg.Go(func() error {
				if _, err := a.pachClient.InspectTag(auth.In2Out(ctx), &pfs.Tag{tag}); err == nil {
					foundTag = true
				}
				return nil
			})
			eg.Go(func() error {
				if objectInfo, err := a.pachClient.InspectTag(auth.In2Out(ctx), &pfs.Tag{tag15}); err == nil {
					foundTag15 = true
					object = objectInfo.Object
				}
				return nil
			})
			if err := eg.Wait(); err != nil {
				return err
			}
			var statsTag *pfs.Tag
			if a.pipelineInfo.EnableStats {
				statsTag = &pfs.Tag{tag + statsTagSuffix}
			}
			if foundTag15 && !foundTag {
				if _, err := a.pachClient.ObjectAPIClient.TagObject(auth.In2Out(ctx),
					&pfs.TagObjectRequest{
						Object: object,
						Tags:   []*pfs.Tag{&pfs.Tag{tag}},
					}); err != nil {
					return err
				}
				if _, err := a.pachClient.ObjectAPIClient.DeleteTags(auth.In2Out(ctx),
					&pfs.DeleteTagsRequest{
						Tags: []string{tag15},
					}); err != nil {
					return err
				}
			}
			if foundTag15 || foundTag {
				skipped++
				return nil
			}
			subStats := &pps.ProcessStats{}
			statsPath := path.Join("/", logger.template.DatumID)
			var statsTree hashtree.OpenHashTree
			if a.pipelineInfo.EnableStats {
				statsTree = hashtree.NewHashTree()
				if err := statsTree.PutFile(path.Join(statsPath, fmt.Sprintf("job:%s", jobInfo.Job.ID)), nil, 0); err != nil {
					logger.stderrLog.Printf("error from hashtree.PutFile for job object: %s\n", err)
				}
				defer func() {
					if retErr != nil {
						return
					}
					finStatsTree, err := statsTree.Finish()
					if err != nil {
						retErr = err
						return
					}
					statsTreeBytes, err := hashtree.Serialize(finStatsTree)
					if err != nil {
						retErr = err
						return
					}
					if _, _, err := a.pachClient.PutObject(bytes.NewReader(statsTreeBytes), statsTag.Name); err != nil {
						retErr = err
						return
					}
				}()
				defer func() {
					object, size, err := logger.Close()
					if err != nil && retErr == nil {
						retErr = err
						return
					}
					if object != nil && a.pipelineInfo.EnableStats {
						if err := statsTree.PutFile(path.Join(statsPath, "logs"), []*pfs.Object{object}, size); err != nil && retErr == nil {
							retErr = err
							return
						}
					}
				}()
				defer func() {
					marshaler := &jsonpb.Marshaler{}
					statsString, err := marshaler.MarshalToString(subStats)
					if err != nil {
						logger.stderrLog.Printf("could not serialize stats: %s\n", err)
						return
					}
					object, size, err := a.pachClient.PutObject(strings.NewReader(statsString))
					if err != nil {
						logger.stderrLog.Printf("could not put stats object: %s\n", err)
						return
					}
					if err := statsTree.PutFile(path.Join(statsPath, "stats"), []*pfs.Object{object}, size); err != nil {
						logger.stderrLog.Printf("could not put-file stats object: %s\n", err)
						return
					}
				}()
			}
			parentTag, err := a.parentTag(ctx, jobInfo, data)
			if err != nil {
				return err
			}

			env := a.userCodeEnv(jobInfo.Job.ID, data)
			atomic.AddInt64(&a.queueSize, -1)
			var dir string
			var retries int
			if err := backoff.RetryNotify(func() error {
				// If the context is already cancelled (timeout, cancelled job), don't run datum
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				// Download input data
				puller := filesync.NewPuller()
				// TODO parent tag shouldn't be nil
				var err error
				dir, err = a.downloadData(logger, data, puller, parentTag, subStats, statsTree, path.Join(statsPath, "pfs"))
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
					return err
				}
				a.runMu.Lock()
				defer a.runMu.Unlock()
				ctx, cancel := context.WithCancel(ctx)
				func() {
					a.statusMu.Lock()
					defer a.statusMu.Unlock()
					a.jobID = jobInfo.Job.ID
					a.data = data
					a.started = time.Now()
					a.cancel = cancel
					a.stats = stats
				}()
				if err := os.MkdirAll(client.PPSInputPrefix, 0666); err != nil {
					return err
				}
				// Create output directory (currently /pfs/out) and run user code
				if err := os.MkdirAll(filepath.Join(dir, "out"), 0666); err != nil {
					return err
				}
				if err := syscall.Mount(dir, client.PPSInputPrefix, "", syscall.MS_BIND, ""); err != nil {
					return err
				}
				defer func() {
					if err := syscall.Unmount(client.PPSInputPrefix, syscall.MNT_DETACH); err != nil && retErr == nil {
						retErr = err
					}
				}()
				if err := a.runUserCode(ctx, logger, env, subStats, jobInfo.DatumTimeout); err != nil {
					return err
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
				return a.uploadOutput(ctx, dir, tag, logger, data, subStats, statsTree, path.Join(statsPath, "pfs", "out"))
			}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
				// If the context is already cancelled (timeout, cancelled job),
				// err out and don't retry
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				retries++
				if retries >= maxRetries {
					logger.Logf("failed to process datum with error: %+v", err)
					if statsTree != nil {
						object, size, err := a.pachClient.PutObject(strings.NewReader(err.Error()))
						if err != nil {
							logger.stderrLog.Printf("could not put error object: %s\n", err)
						} else {
							if err := statsTree.PutFile(path.Join(statsPath, "failure"), []*pfs.Object{object}, size); err != nil {
								logger.stderrLog.Printf("could not put-file error object: %s\n", err)
							}
						}
					}
					return err
				}
				logger.Logf("failed processing datum: %v, retrying in %v", err, d)
				return nil
			}); err != nil {
				failedDatumID = a.DatumID(data)
				failed++
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
		return "", err
	}
	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.jobs.ReadWrite(stm)
		jobID := jobInfo.Job.ID
		jobInfo := &pps.JobInfo{}
		if err := jobs.Get(jobID, jobInfo); err != nil {
			return err
		}
		jobInfo.DataProcessed += high - low - skipped
		jobInfo.DataSkipped += skipped
		jobInfo.DataFailed += failed
		if jobInfo.Stats == nil {
			jobInfo.Stats = &pps.ProcessStats{}
		}
		if err := mergeStats(jobInfo.Stats, stats); err != nil {
			logger.Logf("failed to merge Stats: %v", err)
		}
		return jobs.Put(jobID, jobInfo)
	}); err != nil {
		return "", err
	}
	return failedDatumID, nil
}

func (a *APIServer) parentTag(ctx context.Context, jobInfo *pps.JobInfo, files []*Input) (*pfs.Tag, error) {
	if !jobInfo.Incremental {
		return nil, nil // don't bother downloading the parent for non-incremental jobs
	}
	ci, err := a.pachClient.InspectCommit(jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get output commit's commitInfo: %v", err)
	}
	if ci.ParentCommit == nil {
		return nil, nil
	}
	parentCi, err := a.pachClient.InspectCommit(ci.ParentCommit.Repo.Name, ci.ParentCommit.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get output commit's parent's commitInfo: %v", err)
	}

	// compare provenance of ci and parentCi to figure out which commit is new
	key := path.Join
	parentProv := make(map[string]*pfs.Commit)
	for i, commit := range parentCi.Provenance {
		parentProv[key(commit.Repo.Name, parentCi.BranchProvenance[i].Name)] = commit
	}
	var newInputCommit, newInputCommitParent *pfs.Commit
	for i, c := range ci.Provenance {
		pc, ok := parentProv[key(c.Repo.Name, ci.BranchProvenance[i].Name)]
		if !ok {
			return nil, nil // 'c' has no parent, so there's no parent tag
		}
		if pc.ID != c.ID {
			if newInputCommit != nil {
				// Multiple new commits have arrived since last run. Process everything
				// from scratch
				return nil, nil
			}
			newInputCommit = c
			newInputCommitParent = pc
		}
	}
	var parentFiles []*Input // the equivalent datum to 'files', which the parent job processed
	for _, file := range files {
		parentFile := proto.Clone(file).(*Input)
		if file.FileInfo.File.Commit.Repo.Name == newInputCommit.Repo.Name && file.FileInfo.File.Commit.ID == newInputCommit.ID {
			// 'file' from datumFactory is in the new input commit
			parentFileInfo, err := a.pachClient.WithCtx(ctx).InspectFile(
				newInputCommitParent.Repo.Name, newInputCommitParent.ID, file.FileInfo.File.Path)
			if err != nil {
				if !isNotFoundErr(err) {
					return nil, err
				}
				// we didn't find a match for this file,
				// so we know there's no matching datum
				break
			}
			parentFile.FileInfo = parentFileInfo
			// also tell downloadData to make a diff for 'file'
			// side effect of the main work (derive _parentOutputTag)
			file.ParentCommit = newInputCommitParent
		}
		parentFiles = append(parentFiles, parentFile)
	}

	// We have derived what files the parent saw -- compute the tag
	if len(parentFiles) == len(files) {
		_parentOutputTag := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, parentFiles)
		return &pfs.Tag{Name: _parentOutputTag}, nil
	}
	return nil, nil
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

// lookupUser is a reimplementation of user.Lookup that doesn't require cgo.
func lookupUser(name string) (_ *user.User, retErr error) {
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
		if parts[0] == name {
			return &user.User{
				Username: parts[0],
				Uid:      parts[2],
				Gid:      parts[3],
				Name:     parts[4],
				HomeDir:  parts[5],
			}, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil, fmt.Errorf("user %s not found", name)
}

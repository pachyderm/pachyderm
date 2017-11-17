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
	"os/exec"
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
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/util"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
)

const (
	// The maximum number of concurrent download/upload operations
	concurrency = 10
	logBuffer   = 25
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
		datumCache: datumCache,
	}
	logger, err := server.getTaggedLogger(context.Background(), "", nil, false)
	if err != nil {
		return nil, err
	}
	numWorkers, err := ppsserver.GetExpectedNumWorkers(kubeClient, pipelineInfo.ParallelismSpec)
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
			user, err := util.LookupUser(image.Config.User)
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
		} else if input.GitURL != "" {
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
func (a *APIServer) runUserCode(ctx context.Context, logger *taggedLogger, environ []string, stats *pps.ProcessStats) (retErr error) {
	defer func(start time.Time) {
		stats.ProcessTime = types.DurationProto(time.Since(start))
	}(time.Now())
	logger.Logf("beginning to run user code")
	defer func(start time.Time) {
		logger.Logf("finished running user code - took (%v) - with error (%v)", time.Since(start), retErr)
	}(time.Now())
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
	err := cmd.Run()

	// Return result
	if err == nil {
		return nil
	}
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
							atomic.AddUint64(&stats.UploadBytes, fileInfo.SizeBytes)
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

// Process processes a datum.
func (a *APIServer) Process(ctx context.Context, req *ProcessRequest) (resp *ProcessResponse, retErr error) {
	// Set the auth parameters for the context
	ctx = a.pachClient.AddMetadata(ctx)

	logger, err := a.getTaggedLogger(ctx, req.JobID, req.Data, req.EnableStats)
	if err != nil {
		return nil, err
	}
	logger.Logf("process call started - request: %v", req)
	defer func(start time.Time) {
		logger.Logf("process call finished - request: %v, response: %v, err %v, duration: %v", req, resp, retErr, time.Since(start))
	}(time.Now())
	atomic.AddInt64(&a.queueSize, 1)
	ctx, cancel := context.WithCancel(ctx)
	// Hash inputs
	tag := HashDatum(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.Salt, req.Data)
	tag15, err := HashDatum15(a.pipelineInfo, req.Data)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	var statsTag *pfs.Tag
	if req.EnableStats {
		statsTag = &pfs.Tag{tag + statsTagSuffix}
	}
	if foundTag15 && !foundTag {
		if _, err := a.pachClient.ObjectAPIClient.TagObject(auth.In2Out(ctx),
			&pfs.TagObjectRequest{
				Object: object,
				Tags:   []*pfs.Tag{&pfs.Tag{tag}},
			}); err != nil {
			return nil, err
		}
		if _, err := a.pachClient.ObjectAPIClient.DeleteTags(auth.In2Out(ctx),
			&pfs.DeleteTagsRequest{
				Tags: []string{tag15},
			}); err != nil {
			return nil, err
		}
	}
	if foundTag15 || foundTag {
		// We've already computed the output for these inputs. Return immediately
		logger.Logf("skipping input, as it's already been processed")
		return &ProcessResponse{
			Skipped: true,
		}, nil
	}
	stats := &pps.ProcessStats{}
	statsPath := path.Join("/", logger.template.DatumID)
	var statsTree hashtree.OpenHashTree
	if req.EnableStats {
		statsTree = hashtree.NewHashTree()
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
			if object != nil && req.EnableStats {
				if err := statsTree.PutFile(path.Join(statsPath, "logs"), []*pfs.Object{object}, size); err != nil && retErr == nil {
					retErr = err
					return
				}
			}
		}()
		defer func() {
			marshaler := &jsonpb.Marshaler{}
			statsString, err := marshaler.MarshalToString(stats)
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

	env := a.userCodeEnv(req.JobID, req.Data)
	// Download input data
	puller := filesync.NewPuller()
	dir, err := a.downloadData(logger, req.Data, puller, req.ParentOutput, stats, statsTree, path.Join(statsPath, "pfs"))
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
		return nil, err
	}
	// Create output directory (currently /pfs/out) and run user code
	if err := os.MkdirAll(filepath.Join(dir, "out"), 0666); err != nil {
		return nil, err
	}
	// unset the status when this function exits
	if response, err := func() (_ *ProcessResponse, retErr error) {
		a.runMu.Lock()
		defer a.runMu.Unlock()
		atomic.AddInt64(&a.queueSize, -1)
		func() {
			a.statusMu.Lock()
			defer a.statusMu.Unlock()
			a.jobID = req.JobID
			a.data = req.Data
			a.started = time.Now()
			a.cancel = cancel
			a.stats = stats
		}()
		if err := os.MkdirAll(client.PPSInputPrefix, 0666); err != nil {
			return nil, err
		}
		if err := syscall.Mount(dir, client.PPSInputPrefix, "", syscall.MS_BIND, ""); err != nil {
			return nil, err
		}
		defer func() {
			if err := syscall.Unmount(client.PPSInputPrefix, syscall.MNT_DETACH); err != nil && retErr == nil {
				retErr = err
			}
		}()
		err = a.runUserCode(ctx, logger, env, stats)
		if err != nil {
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
			return &ProcessResponse{
				Failed: true,
			}, nil
		}
		return nil, nil
	}(); err != nil {
		return nil, err
	} else if response != nil {
		return response, nil
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
	if err := a.uploadOutput(ctx, dir, tag, logger, req.Data, stats, statsTree, path.Join(statsPath, "pfs", "out")); err != nil {
		// If uploading failed because the user program outputed a special
		// file, then there's no point in retrying.  Thus we signal that
		// there's some problem with the user code so the job doesn't
		// infinitely retry to process this datum.
		if err == errSpecialFile {
			return &ProcessResponse{
				Failed: true,
			}, nil
		}
		return nil, err
	}
	return &ProcessResponse{Stats: stats}, nil
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
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(jobInfo.Pipeline.Name, pipelineInfo); err != nil {
			return err
		}
		if pipelineInfo.JobCounts == nil {
			pipelineInfo.JobCounts = make(map[int32]int32)
		}
		if pipelineInfo.JobCounts[int32(jobInfo.State)] != 0 {
			pipelineInfo.JobCounts[int32(jobInfo.State)]--
		}
		pipelineInfo.JobCounts[int32(state)]++
		pipelines.Put(pipelineInfo.Pipeline.Name, pipelineInfo)
	}
	jobInfo.State = state
	jobInfo.Reason = reason
	jobs := a.jobs.ReadWrite(stm)
	jobs.Put(jobInfo.Job.ID, jobInfo)
	return nil
}

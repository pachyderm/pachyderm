package worker

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

const (
	// The maximum number of concurrent download/upload operations
	concurrency = 100
	maxLogItems = 10
)

// Input is a generic input object that can either be a pipeline input or
// a job input.  It only defines the attributes that the worker cares about.
type Input struct {
	Name string
	Lazy bool
}

// APIServer implements the worker API
type APIServer struct {
	sync.Mutex
	pachClient *client.APIClient

	// Information needed to process input data and upload output
	transform *pps.Transform
	inputs    []*Input

	// Information attached to log lines
	logMsgTemplate pps.LogMessage

	// The k8s pod name of this worker
	workerName string
}

type taggedLogger struct {
	template  pps.LogMessage
	stderrLog log.Logger
	marshaler *jsonpb.Marshaler
	buffer    bytes.Buffer
}

func (a *APIServer) getTaggedLogger(req *ProcessRequest) *taggedLogger {
	result := &taggedLogger{
		template:  a.logMsgTemplate, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
	}
	result.stderrLog.SetOutput(os.Stderr)
	result.stderrLog.SetFlags(log.LstdFlags | log.Llongfile) // Log file/line

	// Add Job ID to log metadata
	result.template.JobID = req.JobID

	// Add inputs' details to log metadata, so we can find these logs later
	result.template.Data = make([]*pps.Datum, 0, len(req.Data))
	for i, d := range req.Data {
		result.template.Data = append(result.template.Data, new(pps.Datum))
		result.template.Data[i].Path = d.File.Path
		result.template.Data[i].Hash = d.Hash
	}
	return result
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
	fmt.Printf("%s\n", bytes)
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
			return 0, err
		}
		logger.Logf(message)
	}
}

func (logger *taggedLogger) userLogger() *taggedLogger {
	result := &taggedLogger{
		template:  logger.template, // Copy struct
		stderrLog: log.Logger{},
		marshaler: &jsonpb.Marshaler{},
	}
	result.template.User = true
	return result
}

// NewPipelineAPIServer creates an APIServer for a given pipeline
func NewPipelineAPIServer(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, workerName string) *APIServer {
	server := &APIServer{
		Mutex:      sync.Mutex{},
		pachClient: pachClient,
		transform:  pipelineInfo.Transform,
		inputs:     make([]*Input, 0, len(pipelineInfo.Inputs)),
		logMsgTemplate: pps.LogMessage{
			PipelineName: pipelineInfo.Pipeline.Name,
			PipelineID:   pipelineInfo.ID,
			WorkerID:     os.Getenv(client.PPSPodNameEnv),
		},
		workerName: workerName,
	}
	for _, input := range pipelineInfo.Inputs {
		server.inputs = append(server.inputs, &Input{
			Name: input.Name,
			Lazy: input.Lazy,
		})
	}
	return server
}

// NewJobAPIServer creates an APIServer for a given pipeline
func NewJobAPIServer(pachClient *client.APIClient, jobInfo *pps.JobInfo, workerName string) *APIServer {
	server := &APIServer{
		Mutex:          sync.Mutex{},
		pachClient:     pachClient,
		transform:      jobInfo.Transform,
		inputs:         make([]*Input, 0, len(jobInfo.Inputs)),
		logMsgTemplate: pps.LogMessage{},
		workerName:     workerName,
	}
	for _, input := range jobInfo.Inputs {
		server.inputs = append(server.inputs, &Input{
			Name: input.Name,
			Lazy: input.Lazy,
		})
	}
	return server
}

func (a *APIServer) downloadData(data []*pfs.FileInfo, puller *filesync.Puller) error {
	for i, datum := range data {
		input := a.inputs[i]
		if err := puller.Pull(a.pachClient, filepath.Join(client.PPSInputPrefix, input.Name, datum.File.Path), datum.File.Commit.Repo.Name, datum.File.Commit.ID, datum.File.Path, input.Lazy, concurrency); err != nil {
			return err
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context, logger *taggedLogger, environ []string) error {
	// Run user code
	transform := a.transform
	cmd := exec.Command(transform.Cmd[0], transform.Cmd[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(transform.Stdin, "\n") + "\n")
	cmd.Stdout = logger.userLogger()
	cmd.Stderr = logger.userLogger()
	cmd.Env = environ
	err := cmd.Run()
	// Log output from user cmd, line-by-line, whether or not cmd errored
	logger.Logf("running user code")
	// Return result
	if err == nil {
		return nil
	}
	// (if err is an acceptable return code, don't return err)
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			for _, returnCode := range transform.AcceptReturnCode {
				if int(returnCode) == status.ExitStatus() {
					return nil
				}
			}
		}
	}
	return err

}

func (a *APIServer) uploadOutput(ctx context.Context, tag string) error {
	// hashtree is not thread-safe--guard with 'lock'
	var lock sync.Mutex
	tree := hashtree.NewHashTree()

	// Upload all files in output directory
	var g errgroup.Group
	limiter := limit.New(concurrency)
	if err := filepath.Walk(client.PPSOutputPath, func(path string, info os.FileInfo, err error) error {
		g.Go(func() (retErr error) {
			limiter.Acquire()
			defer limiter.Release()
			if path == client.PPSOutputPath {
				return nil
			}

			relPath, err := filepath.Rel(client.PPSOutputPath, path)
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

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()

			object, size, err := a.pachClient.PutObject(f)
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			return tree.PutFile(relPath, []*pfs.Object{object}, size)
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

// cleanUpData removes everything under /pfs
//
// The reason we don't want to just os.RemoveAll(/pfs) is that we don't
// want to remove /pfs itself, since it's a symlink to the hostpath volume.
// We also don't want to remove /pfs and re-create the symlink, because for
// some reason that results in extremely pool performance.
//
// Most of the code is copied from os.RemoveAll().
func (a *APIServer) cleanUpData() error {
	path := filepath.Join(client.PPSHostPath, a.workerName)
	// Otherwise, is this a directory we need to recurse into?
	dir, serr := os.Lstat(path)
	if serr != nil {
		if serr, ok := serr.(*os.PathError); ok && (os.IsNotExist(serr.Err) || serr.Err == syscall.ENOTDIR) {
			return nil
		}
		return serr
	}
	if !dir.IsDir() {
		// Not a directory; return the error from Remove.
		return fmt.Errorf("%s is not a directory", path)
	}

	// Directory.
	fd, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Race. It was deleted between the Lstat and Open.
			// Return nil per RemoveAll's docs.
			return nil
		}
		return err
	}

	// Remove contents & return first error.
	err = nil
	for {
		names, err1 := fd.Readdirnames(100)
		for _, name := range names {
			err1 := os.RemoveAll(path + string(os.PathSeparator) + name)
			if err == nil {
				err = err1
			}
		}
		if err1 == io.EOF {
			break
		}
		// If Readdirnames returned an error, use it.
		if err == nil {
			err = err1
		}
		if len(names) == 0 {
			break
		}
	}

	// Close directory, because windows won't remove opened directory.
	fd.Close()
	return err
}

// HashDatum computes and returns the hash of a datum + pipeline.
func (a *APIServer) HashDatum(data []*pfs.FileInfo) (string, error) {
	hash := sha256.New()
	for i, fileInfo := range data {
		if _, err := hash.Write([]byte(a.inputs[i].Name)); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(fileInfo.File.Path)); err != nil {
			return "", err
		}
		if _, err := hash.Write(fileInfo.Hash); err != nil {
			return "", err
		}
	}
	bytes, err := proto.Marshal(a.transform)
	if err != nil {
		return "", err
	}
	if _, err := hash.Write(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Process processes a datum.
func (a *APIServer) Process(ctx context.Context, req *ProcessRequest) (resp *ProcessResponse, retErr error) {
	// We cannot run more than one user process at once; otherwise they'd be
	// writing to the same output directory. Acquire lock to make sure only one
	// user process runs at a time.
	a.Lock()
	defer a.Unlock()
	logger := a.getTaggedLogger(req)
	logger.Logf("Received request")

	// Hash inputs and check if output is in s3 already. Note: ppsserver sorts
	// inputs by input name for both jobs and pipelines, so this hash is stable
	// even if a.Inputs are reordered by the user
	tag, err := a.HashDatum(req.Data)
	if err != nil {
		return nil, err
	}
	if _, err := a.pachClient.InspectTag(ctx, &pfs.Tag{tag}); err == nil {
		// We've already computed the output for these inputs. Return immediately
		logger.Logf("skipping input, as it's already been processed")
		return &ProcessResponse{
			Tag: &pfs.Tag{tag},
		}, nil
	}

	// Download input data
	logger.Logf("input has not been processed, downloading data")
	puller := filesync.NewPuller()
	if err := a.downloadData(req.Data, puller); err != nil {
		return nil, err
	}
	defer func() {
		if err := puller.CleanUp(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	defer func() {
		if err := a.cleanUpData(); retErr == nil && err != nil {
			retErr = err
		}
	}()

	environ := a.userCodeEnviron(req)

	// Create output directory (currently /pfs/out) and run user code
	if err := os.MkdirAll(client.PPSOutputPath, 0666); err != nil {
		return nil, err
	}
	logger.Logf("beginning to process user input")
	err = a.runUserCode(ctx, logger, environ)
	logger.Logf("finished processing user input")
	if err != nil {
		return &ProcessResponse{
			Failed: true,
		}, nil
	}
	if err := a.uploadOutput(ctx, tag); err != nil {
		return nil, err
	}
	return &ProcessResponse{
		Tag: &pfs.Tag{tag},
	}, nil
}

func (a *APIServer) userCodeEnviron(req *ProcessRequest) []string {
	return append(os.Environ(), fmt.Sprintf("PACH_JOB_ID=%s", req.JobID))
}

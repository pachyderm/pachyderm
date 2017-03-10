package worker

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

// Input is a generic input object that can either be a pipeline input or
// a job input.  It only defines the attributes that the worker cares about.
type Input struct {
	Name string
	Lazy bool
}

// logInput contains metadata about job or pipeline's inputs, which this worker
// uses to annotate its log lines
type logInput struct {
	source string
	path   string
	hash   string
}

// logInfo is a struct that attaches some metadata to log lines. We json-
// encode this struct and then emit it to stdout, so that elasticsearch can
// consume the metadata and make it searchable
type logInfo struct {
	// Either (PipelineName and PipelineID) or (JobID) will be set
	PipelineName string
	PipelineID   string
	JobID        string

	// These will be updated in each call to process
	Inputs []logInput

	// The message being logged with each call
	Message string
}

// APIServer implements the worker API
type APIServer struct {
	sync.Mutex
	protorpclog.Logger
	pachClient *client.APIClient

	// Information needed to process input data and upload output
	transform *pps.Transform
	inputs    []*Input

	// Information attached to log lines
	logInfoTemplate logInfo
}

func (a *APIServer) taggedLogger(req *ProcessRequest) logInfo {
	result := a.logInfoTemplate // Copy struct, so we don't change the template

	// Add inputs detauls to log metadata, so we can find these logs later
	for i, d := range req.Data {
		result.Inputs[i].hash = string(d.Hash)
		result.Inputs[i].path = d.File.Path
	}
	return result
}

func (loginfo logInfo) Logf(formatString string, args ...interface{}) {
	loginfo.Message = fmt.Sprintf(formatString, args)
	bytes, err := json.Marshal(&loginfo)
	if err != nil {
		// log the struct anyway?
		log.Printf("could not marshal %v for logging: %s", loginfo, err)
		return
	}
	log.Println(bytes)
}

// NewPipelineAPIServer creates an APIServer for a given pipeline
func NewPipelineAPIServer(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) *APIServer {
	result := &APIServer{
		Mutex:      sync.Mutex{},
		Logger:     protorpclog.NewLogger(""),
		pachClient: pachClient,
		transform:  pipelineInfo.Transform,
		inputs:     make([]*Input, 0, len(pipelineInfo.Inputs)),
		logInfoTemplate: logInfo{
			PipelineName: pipelineInfo.Pipeline.Name,
			PipelineID:   pipelineInfo.ID,
			Inputs:       make([]logInput, 0, len(pipelineInfo.Inputs)),
		},
	}
	for _, input := range pipelineInfo.Inputs {
		result.inputs = append(result.inputs, &Input{
			Name: input.Name,
			Lazy: input.Lazy,
		})
		result.logInfoTemplate.Inputs = append(result.logInfoTemplate.Inputs, logInput{
			source: input.Name,
		})
	}
	return result
}

// NewJobAPIServer creates an APIServer for a given pipeline
func NewJobAPIServer(pachClient *client.APIClient, jobInfo *pps.JobInfo) *APIServer {
	result := &APIServer{
		Mutex:      sync.Mutex{},
		Logger:     protorpclog.NewLogger(""),
		pachClient: pachClient,
		transform:  jobInfo.Transform,
		inputs:     make([]*Input, 0, len(jobInfo.Inputs)),
		logInfoTemplate: logInfo{
			JobID:  jobInfo.Job.ID,
			Inputs: make([]logInput, 0, len(jobInfo.Inputs)),
		},
	}
	for _, input := range jobInfo.Inputs {
		result.inputs = append(result.inputs, &Input{
			Name: input.Name,
			Lazy: input.Lazy,
		})
		result.logInfoTemplate.Inputs = append(result.logInfoTemplate.Inputs, logInput{
			source: input.Name,
		})
	}
	return result
}

func (a *APIServer) downloadData(ctx context.Context, data []*pfs.FileInfo) error {
	for i, datum := range data {
		input := a.inputs[i]
		if err := filesync.Pull(ctx, a.pachClient, filepath.Join(client.PPSInputPrefix, input.Name), datum, input.Lazy); err != nil {
			return err
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context, taggedLogger *logInfo) (string, error) {
	// Run user code
	transform := a.transform
	cmd := exec.Command(transform.Cmd[0], transform.Cmd[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(transform.Stdin, "\n") + "\n")
	var userlog bytes.Buffer
	cmd.Stdout = &userlog
	cmd.Stderr = &userlog
	err := cmd.Run()

	// Log output from user cmd, line-by-line, whether or not cmd errored
	taggedLogger.Logf("running user code")
	logscanner := bufio.NewScanner(&userlog)
	for logscanner.Scan() {
		taggedLogger.Logf(logscanner.Text())
	}

	// Return result
	if err == nil {
		return userlog.String(), nil
	}
	// (if err is an acceptable return code, don't return err)
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			for _, returnCode := range transform.AcceptReturnCode {
				if int(returnCode) == status.ExitStatus() {
					return userlog.String(), nil
				}
			}
		}
	}
	return userlog.String(), err

}

func (a *APIServer) uploadOutput(ctx context.Context, tag string) error {
	// hashtree is not thread-safe--guard with 'lock'
	var lock sync.Mutex
	tree := hashtree.NewHashTree()

	// Upload all files in output directory
	var g errgroup.Group
	if err := filepath.Walk(client.PPSOutputPath, func(path string, info os.FileInfo, err error) error {
		g.Go(func() (retErr error) {
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

// HashDatum computes and returns the hash of a datum + pipeline.
func (a *APIServer) HashDatum(data []*pfs.FileInfo) (string, error) {
	hash := sha256.New()
	sort.Slice(data, func(i, j int) bool {
		return data[i].File.Path < data[j].File.Path
	})
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
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())

	// We cannot run more than one user process at once; otherwise they'd be
	// writing to the same output directory. Acquire lock to make sure only one
	// user process runs at a time.
	a.Lock()
	defer a.Unlock()
	taggedLogger := a.taggedLogger(req)
	taggedLogger.Logf("Recieved request")

	// Hash inputs and check if output is in s3 already. Note: ppsserver sorts
	// inputs by input name for both jobs and pipelines, so this hash is stable
	// even if a.Inputs are reordered by the user
	tag, err := a.HashDatum(req.Data)
	if err != nil {
		return nil, err
	}
	if _, err := a.pachClient.InspectTag(ctx, &pfs.Tag{tag}); err == nil {
		// We've already computed the output for these inputs. Return immediately
		taggedLogger.Logf("skipping input, as it's already been processed")
		return &ProcessResponse{
			Tag: &pfs.Tag{tag},
		}, nil
	}

	// Download input data
	taggedLogger.Logf("input has not been processed, downloading data")
	if err := a.downloadData(ctx, req.Data); err != nil {
		return nil, err
	}
	defer func() {
		if err := os.RemoveAll(client.PPSInputPrefix); retErr == nil && err != nil {
			retErr = err
			return
		}
	}()

	// Create output directory (currently /pfs/out) and run user code
	if err := os.MkdirAll(client.PPSOutputPath, 0666); err != nil {
		return nil, err
	}
	userlog, err := a.runUserCode(ctx, &taggedLogger)
	if err != nil {
		return &ProcessResponse{
			Log: userlog,
		}, nil
	}
	if err := a.uploadOutput(ctx, tag); err != nil {
		return nil, err
	}
	return &ProcessResponse{
		Tag: &pfs.Tag{tag},
	}, nil
}

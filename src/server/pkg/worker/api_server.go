package worker

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

const (
	// The maximum number of concurrent download/upload operations
	concurrency = 10
	maxLogItems = 10
)

var (
	errSpecialFile = errors.New("cannot upload special file")
)

// APIServer implements the worker API
type APIServer struct {
	pachClient *client.APIClient

	// Information needed to process input data and upload output
	pipelineInfo *pps.PipelineInfo

	// Information attached to log lines
	logMsgTemplate pps.LogMessage

	statusMu sync.Mutex
	// The currently running job ID
	jobID string
	// The currently running data
	data []*Input
	// The time we started the currently running
	started time.Time
	// Func to cancel the currently running datum
	cancel func()
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
	for _, d := range req.Data {
		result.template.Data = append(result.template.Data, &pps.Datum{
			Path: d.FileInfo.File.Path,
			Hash: d.FileInfo.Hash,
		})
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
		pachClient:   pachClient,
		pipelineInfo: pipelineInfo,
		logMsgTemplate: pps.LogMessage{
			PipelineName: pipelineInfo.Pipeline.Name,
			PipelineID:   pipelineInfo.ID,
			WorkerID:     os.Getenv(client.PPSPodNameEnv),
		},
		workerName: workerName,
	}
	return server
}

func (a *APIServer) downloadData(logger *taggedLogger, inputs []*Input, puller *filesync.Puller, parentTag *pfs.Tag) error {
	for _, input := range inputs {
		file := input.FileInfo.File
		root := filepath.Join(client.PPSInputPrefix, input.Name, file.Path)
		if a.pipelineInfo.Incremental && input.ParentCommit != nil {
			if err := puller.PullDiff(a.pachClient, root,
				file.Commit.Repo.Name, file.Commit.ID, file.Path,
				input.ParentCommit.Repo.Name, input.ParentCommit.ID, file.Path,
				true, input.Lazy, concurrency); err != nil {
				return err
			}
		} else {
			if err := puller.Pull(a.pachClient, root, file.Commit.Repo.Name, file.Commit.ID, file.Path, input.Lazy, concurrency); err != nil {
				return err
			}
		}
	}
	if parentTag != nil {
		var buffer bytes.Buffer
		if err := a.pachClient.GetTag(parentTag.Name, &buffer); err != nil {
			logger.Logf("error getting parent for datum %v: %v", inputs, err)
		}
		tree, err := hashtree.Deserialize(buffer.Bytes())
		if err != nil {
			return fmt.Errorf("failed to deserialize parent hashtree: %v", err)
		}
		if err := puller.PullTree(a.pachClient, client.PPSOutputPath, tree, false, concurrency); err != nil {
			return fmt.Errorf("error pulling output tree: %+v", err)
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context, logger *taggedLogger, environ []string) error {
	// Run user code
	cmd := exec.CommandContext(ctx, a.pipelineInfo.Transform.Cmd[0], a.pipelineInfo.Transform.Cmd[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(a.pipelineInfo.Transform.Stdin, "\n") + "\n")
	cmd.Stdout = logger.userLogger()
	cmd.Stderr = logger.userLogger()
	logger.Logf("running user code")
	cmd.Env = environ
	err := cmd.Run()
	if err != nil {
		logger.Logf("user code finished, err: %+v", err)
	} else {
		logger.Logf("user code finished")
	}

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

func (a *APIServer) uploadOutput(ctx context.Context, tag string, logger *taggedLogger, inputs []*Input) error {
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
				realPath, err := os.Readlink(path)
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
					if input != nil {
						return filepath.Walk(realPath, func(path string, info os.FileInfo, err error) error {
							rel, err := filepath.Rel(realPath, path)
							if err != nil {
								return err
							}
							subRelPath := filepath.Join(relPath, rel)
							// The path of the input file
							pfsPath, err := filepath.Rel(filepath.Join(client.PPSInputPrefix, input.Name), path)
							if err != nil {
								return err
							}

							if info.IsDir() {
								lock.Lock()
								defer lock.Unlock()
								tree.PutDir(subRelPath)
								return nil
							}

							fileInfo, err := a.pachClient.PfsAPIClient.InspectFile(ctx, &pfs.InspectFileRequest{
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
							return tree.PutFile(subRelPath, fileInfo.Objects, int64(fileInfo.SizeBytes))
						})
					}
				}
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

			putObjClient, err := a.pachClient.ObjectAPIClient.PutObject(ctx)
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

// cleanUpData removes everything under /pfs
//
// The reason we don't want to just os.RemoveAll(/pfs) is that we don't
// want to remove /pfs itself, since it's a emptyDir volume.
//
// Most of the code is copied from os.RemoveAll().
func (a *APIServer) cleanUpData() error {
	path := client.PPSInputPrefix
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

// HashDatum computes and returns the hash of datum + pipeline, with a
// pipeline-specific prefix.
func HashDatum(pipelineInfo *pps.PipelineInfo, data []*Input) (string, error) {
	hash := sha256.New()
	for _, datum := range data {
		hash.Write([]byte(datum.Name))
		hash.Write([]byte(datum.FileInfo.File.Path))
		hash.Write(datum.FileInfo.Hash)
	}

	bytes, err := proto.Marshal(pipelineInfo.Transform)
	if err != nil {
		return "", err
	}
	hash.Write(bytes)
	hash.Write([]byte(pipelineInfo.Pipeline.Name))
	hash.Write([]byte(pipelineInfo.ID))
	hash.Write([]byte(strconv.Itoa(int(pipelineInfo.Version))))

	return client.HashPipelineID(pipelineInfo.ID) + hex.EncodeToString(hash.Sum(nil)), nil
}

// Process processes a datum.
func (a *APIServer) Process(ctx context.Context, req *ProcessRequest) (resp *ProcessResponse, retErr error) {
	// We cannot run more than one user process at once; otherwise they'd be
	// writing to the same output directory. Acquire lock to make sure only one
	// user process runs at a time.
	// set the status for the datum
	ctx, cancel := context.WithCancel(ctx)
	if err := func() error {
		a.statusMu.Lock()
		defer a.statusMu.Unlock()
		if a.jobID != "" {
			// we error in this case so that callers have a chance to find a
			// non-busy worker
			return fmt.Errorf("worker busy")
		}
		a.jobID = req.JobID
		a.data = req.Data
		a.started = time.Now()
		a.cancel = cancel
		return nil
	}(); err != nil {
		return nil, err
	}
	// unset the status when this function exits
	defer func() {
		a.statusMu.Lock()
		defer a.statusMu.Unlock()
		a.jobID = ""
		a.data = nil
		a.started = time.Time{}
		a.cancel = nil
	}()
	logger := a.getTaggedLogger(req)
	logger.Logf("Received request")

	// Hash inputs and check if output is in s3 already. Note: ppsserver sorts
	// inputs by input name for both jobs and pipelines, so this hash is stable
	// even if a.Inputs are reordered by the user
	tag, err := HashDatum(a.pipelineInfo, req.Data)
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
	err = a.downloadData(logger, req.Data, puller, req.ParentOutput)
	// We run these cleanup functions no matter what, so that if
	// downloadData partially succeeded, we still clean up the resources.
	defer func() {
		if err := a.cleanUpData(); retErr == nil && err != nil {
			retErr = err
		}
	}()
	// It's important that we run puller.CleanUp before a.cleanUpData,
	// because otherwise puller.Cleanup might try tp open pipes that have
	// been deleted.
	defer func() {
		if err := puller.CleanUp(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		return nil, err
	}

	environ := a.userCodeEnviron(req)

	// Create output directory (currently /pfs/out) and run user code
	if err := os.MkdirAll(client.PPSOutputPath, 0666); err != nil {
		return nil, err
	}
	logger.Logf("beginning to process user input")
	err = a.runUserCode(ctx, logger, environ)
	logger.Logf("finished processing user input")
	if err != nil {
		logger.Logf("failed to process datum with error: %+v", err)
		return &ProcessResponse{
			Failed: true,
		}, nil
	}
	// CleanUp is idempotent so we can call it however many times we want.
	// The reason we are calling it here is that the puller could've
	// encountered an error as it was lazily loading files, in which case
	// the output might be invalid since as far as the user's code is
	// concerned, they might've just seen an empty or partially completed
	// file.
	if err := puller.CleanUp(); err != nil {
		logger.Logf("puller encountered an error while cleaning up: %+v", err)
		return nil, err
	}
	if err := a.uploadOutput(ctx, tag, logger, req.Data); err != nil {
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
	return &ProcessResponse{
		Tag: &pfs.Tag{tag},
	}, nil
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
		JobID:    a.jobID,
		WorkerID: a.workerName,
		Started:  started,
		Data:     a.datum(),
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

func (a *APIServer) datum() []*pps.Datum {
	var result []*pps.Datum
	for _, datum := range a.data {
		result = append(result, &pps.Datum{
			Path: datum.FileInfo.File.Path,
			Hash: datum.FileInfo.Hash,
		})
	}
	return result
}

func (a *APIServer) userCodeEnviron(req *ProcessRequest) []string {
	return append(os.Environ(), fmt.Sprintf("PACH_JOB_ID=%s", req.JobID))
}

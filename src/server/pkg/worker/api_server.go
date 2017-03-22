package worker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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
	"github.com/pachyderm/pachyderm/src/client/limit"
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

// Options are the options used to initialize a worker.
type Options struct {
	Transform  *pps.Transform
	Inputs     []*Input
	WorkerName string
}

// APIServer implements the worker API
type APIServer struct {
	sync.Mutex
	protorpclog.Logger
	pachClient *client.APIClient
	options    *Options
}

const (
	// the number of files that can be downloaded/uploaded at the same time
	parallelism = 100
)

// NewAPIServer creates an APIServer for a given pipeline
func NewAPIServer(pachClient *client.APIClient, options *Options) *APIServer {
	return &APIServer{
		Mutex:      sync.Mutex{},
		Logger:     protorpclog.NewLogger(""),
		pachClient: pachClient,
		options:    options,
	}
}

func (a *APIServer) downloadData(ctx context.Context, data []*pfs.FileInfo) error {
	for i, datum := range data {
		input := a.options.Inputs[i]
		if err := filesync.Pull(ctx, a.pachClient, filepath.Join(client.PPSInputPrefix, input.Name), datum, input.Lazy, parallelism); err != nil {
			return err
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context) (string, error) {
	transform := a.options.Transform
	cmd := exec.Command(transform.Cmd[0], transform.Cmd[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(transform.Stdin, "\n") + "\n")
	var log bytes.Buffer
	cmd.Stdout = &log
	cmd.Stderr = &log
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						return log.String(), nil
					}
				}
			}
		}
		return log.String(), err
	}
	return log.String(), nil
}

func (a *APIServer) uploadOutput(ctx context.Context, tag string) error {
	// hashtree is not thread-safe--guard with 'lock'
	var lock sync.Mutex
	tree := hashtree.NewHashTree()

	// Upload all files in output directory
	var g errgroup.Group
	limiter := limit.New(parallelism)
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
	path := filepath.Join(client.PPSHostPath, a.options.WorkerName)
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
func HashDatum(data []*pfs.FileInfo, options *Options) (string, error) {
	hash := sha256.New()
	sort.Slice(data, func(i, j int) bool {
		return data[i].File.Path < data[j].File.Path
	})
	for i, fileInfo := range data {
		if _, err := hash.Write([]byte(options.Inputs[i].Name)); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(fileInfo.File.Path)); err != nil {
			return "", err
		}
		if _, err := hash.Write(fileInfo.Hash); err != nil {
			return "", err
		}
	}
	bytes, err := proto.Marshal(options.Transform)
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

	// ppsserver sorts inputs by input name, so this is stable even if
	// a.options.Inputs are reordered by the user
	tag, err := HashDatum(req.Data, a.options)
	if err != nil {
		return nil, err
	}
	if _, err := a.pachClient.InspectTag(ctx, &pfs.Tag{tag}); err == nil {
		// We've already computed the output for these inputs. Return immediately
		return &ProcessResponse{
			Tag: &pfs.Tag{tag},
		}, nil
	}

	if err := a.downloadData(ctx, req.Data); err != nil {
		return nil, err
	}
	// Create output directory (currently /pfs/out)
	if err := os.MkdirAll(client.PPSOutputPath, 0666); err != nil {
		return nil, err
	}
	defer func() {
		if err := a.cleanUpData(); retErr == nil && err != nil {
			retErr = err
			return
		}
	}()
	if log, err := a.runUserCode(ctx); err != nil {
		return &ProcessResponse{
			Log: log,
		}, nil
	}
	if err := a.uploadOutput(ctx, tag); err != nil {
		return nil, err
	}
	return &ProcessResponse{
		Tag: &pfs.Tag{tag},
	}, nil
}

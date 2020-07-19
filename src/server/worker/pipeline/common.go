package pipeline

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	//"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
	"github.com/rjeczalik/notify"
)

// TarFile holder structure
type TarFile struct {
	hdr  *tar.Header
	body []byte
}

var fileList []TarFile
var curr TarFile
var fileCnt int

// SkipTarPadding return number of bytes skipped from a tar stream and an EOF
func SkipTarPadding(r *bytes.Reader) (int64, error) {
	var n int64 = 0
	for {
		c, err := r.ReadByte()
		if err != nil {
			// Read byte should succeed unless it reaches EOF
			return n, err
		}
		if c != 0 {
			// Found a non-zero byte. This is start of a new stream
			// Unread the byte
			r.UnreadByte()
			return n, nil
		}
		n++
	}
	// unreachable
	return n, nil
}

// TarReader is the heart of tar state machine. It return the list of tar files received on the pipe
// It reads the buffer from the pipe, processes all the byte
// if there is an incomplete buffer, it goes around to read from the pipe
// TDOD: only read the header from the ReadAll then read the size from the hdr
//       That's the only way to make sure we don't buffer too much
func TarReader(buff bytes.Buffer, logger logs.TaggedLogger) {
	r := bytes.NewReader(buff.Bytes())
	for r.Len() > 0 {
		tr := tar.NewReader(r)
		if curr.hdr == nil {
			// We are reading this file for the first time
			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					curr.hdr = hdr
				}
				logger.Logf("Error in reading tar header", err)
				return
			}
			curr.hdr = hdr
		}
		var bBytes int = int(curr.hdr.Size) - len(curr.body)
		var body []byte = make([]byte, bBytes)
		_, err := tr.Read(body)
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				logger.Logf("Error in reading tar file", err)
				return
			}
		}
		logger.Logf("Read body", len(body))
		curr.body = append(curr.body, body...)
		if err == io.EOF {
			// Read the entire body
			fileList = append(fileList, curr)
			logger.Logf("Received file %s of len %d\n", curr.hdr.Name, curr.hdr.Size)
			fileCnt++
			curr.hdr = nil
			curr.body = nil
		}

		// skip all the padding
		SkipTarPadding(r)
	}

	logger.Logf("Received %d tar files\n", fileCnt)
}

// RunUserCode will run the pipeline's user code until canceled by the context
// - used for services and spouts. Unlike how the transform worker runs user
// code, this does not set environment variables or collect stats.
func RunUserCode(
	driver driver.Driver,
	logger logs.TaggedLogger,
	outputCommit *pfs.Commit,
	inputs []*common.Input,
) error {
	return backoff.RetryUntilCancel(driver.PachClient().Ctx(), func() error {
		// TODO: what about the user error handling code?
		env := driver.UserCodeEnv(logger.JobID(), outputCommit, inputs)
		return driver.RunUserCode(logger, env, &pps.ProcessStats{}, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in RunUserCode: %+v, retrying in: %+v", err, d)
		return nil
	})
}

func processFiles(ctx context.Context,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger) (retErr1 error) {
	repo := pipelineInfo.Pipeline.Name
	var commit *pfs.Commit
	var err error
	for _, f := range fileList {
		if commit == nil {
			// start commit
			commit, err = pachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
				Parent:     client.NewCommit(repo, ""),
				Branch:     pipelineInfo.OutputBranch,
				Provenance: []*pfs.CommitProvenance{client.NewCommitProvenance(ppsconsts.SpecRepo, repo, pipelineInfo.SpecCommit.ID)},
			})
			if err != nil {
				return err
			}

			// finish the commit even if there was an issue
			defer func() {
				if err := pachClient.FinishCommit(repo, commit.ID); err != nil {
					// this lets us pass the error through if FinishCommit fails
					retErr1 = err
				}
			}()
		}
		r := bytes.NewReader(f.body)
		// put files into pachyderm
		if pipelineInfo.Spout.Marker != "" && strings.HasPrefix(path.Clean(f.hdr.Name), pipelineInfo.Spout.Marker) {
			// we'll check that this is the latest version of the spout, and then commit to it
			// we need to do this atomically because we otherwise might hit a subtle race condition

			// check to see if this spout is the latest version of this spout by seeing if its spec commit has any children
			spec, err := pachClient.InspectCommit(ppsconsts.SpecRepo, pipelineInfo.SpecCommit.ID)
			if err != nil && !errutil.IsNotFoundError(err) {
				return err
			}
			if spec != nil && len(spec.ChildCommits) != 0 {
				return errors.New("outdated spout, now shutting down")
			}
			_, err = pachClient.PutFileOverwrite(repo, ppsconsts.SpoutMarkerBranch, f.hdr.Name, r, 0)
			if err != nil {
				return err
			}

		} else if pipelineInfo.Spout.Overwrite {
			_, err = pachClient.PutFileOverwrite(repo, commit.ID, f.hdr.Name, r, 0)
			if err != nil {
				return err
			}
		} else {
			_, err = pachClient.PutFile(repo, commit.ID, f.hdr.Name, r)
			if err != nil {
				return err
			}
		}
	}
	return retErr1
}

func openPipeReadOnly(pName string, logger logs.TaggedLogger) (*os.File, error) {
	p, err := os.Open(pName)
	if os.IsNotExist(err) {
		logger.Logf("Named pipe '%s' does not exist", pName)
	} else if os.IsPermission(err) {
		logger.Logf("Insufficient permissions to read named pipe '%s': %s", pName, err)
	} else if err != nil {
		logger.Logf("Error while opening named pipe '%s': %s", pName, err)
	}
	return p, err
}

func processFilesReceived(
	ctx context.Context,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger) error {
	if fileCnt > 0 {
		err := processFiles(ctx, pachClient, pipelineInfo, logger)
		fileList = nil
		fileCnt = 0
		if err != nil {
			return err
		}
	}
	return nil
}

// ReceiveSpout is used by both services and spouts if a spout is defined on the
// pipeline. ctx is separate from pachClient because services may call this, and
// they use a cancel function that affects the context but not the pachClient
// (so metadata updates can still be made while unwinding).
func ReceiveSpout(
	ctx context.Context,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
) error {
	// open a read connection to the /pfs/out named pipe
	out, err := openPipeReadOnly("/pfs/out", logger)
	if err != nil {
		return err
	}
	// and close it when we're done
	defer out.Close()

	c := make(chan notify.EventInfo, 5)
	notify.Watch("/pfs/out", c, notify.Write|notify.Remove)
	for {
		e := <-c
		switch e.Event() {
		case notify.Write:
			var buff bytes.Buffer
			n, err := io.Copy(&buff, out)
			if err != nil {
				logger.Logf("Received an EOF", err)
				break
			}
			if n == 0 {
				continue
			}
			logger.Logf("Got data: ", n, buff.Len())
			TarReader(buff, logger)
			logger.Logf("Total files", fileCnt)
			processFilesReceived(ctx, pachClient, pipelineInfo, logger)
		case notify.Remove:
			logger.Logf("Fatal: pipe /pfs/out removed")
			return err
		}
	}
	return nil
}

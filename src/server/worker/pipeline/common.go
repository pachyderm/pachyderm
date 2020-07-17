package pipeline

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
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
)

// TarFile holder structure
type TarFile struct {
	hdr  *tar.Header
	body []byte
}

/*
func openAndWait() error {
	// at the end of file, we open the pipe again, since this blocks until something is written to the pipe
	openAndWait, err := os.Open("/pfs/out")
	if err != nil {
		return err
	}
	// and then we immediately close this reader of the pipe, so that the main reader can continue its standard behavior
	err = openAndWait.Close()
	if err != nil {
		return err
	}
	return nil
}
*/

// TarState machine
type TarState int

// NewTar, Partial, TarEOF, TarERR
const (
	NewTar TarState = iota
	Partial
	TarEOF
	TarERR
)

func (t TarState) String() string {
	return [...]string{"NewTar", "Partial", "TarEOF", "TarERR"}[t]
}

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

// TarBody return the bytes read for a tar stream body. It should only get an ERO or unexpected EOF
func TarBody(tr io.Reader, min int64) ([]byte, error) {
	buf := make([]byte, min)
	_, err := io.ReadAtLeast(tr, buf, int(min))
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return buf, err
	}
	return buf, err
}

// NewTarStream returns a tar state, number of byte remaining to read in a stream,
// number of bytes remain in buffer and error
// Depending on what we see in the steam, the next state is set. Ideally, we just handle that in the
// TarReader function
func NewTarStream(r *bytes.Reader) (TarState, int64, int64, *TarFile, error) {
	var bytesToRead int64 = 0
	var blen int = r.Len()
	var state TarState = TarEOF
	var f TarFile
	var err error

	if _, err = SkipTarPadding(r); err == io.EOF {
		return state, bytesToRead, int64(blen - r.Len()), nil, err
	}
	tr := tar.NewReader(r)
	f.hdr, _ = tr.Next()
	f.body, err = TarBody(tr, f.hdr.Size)
	bytesToRead = f.hdr.Size - int64(len(f.body))
	if bytesToRead != 0 {
		// More bytes to read for the current tar stream. return Partial state
		state = Partial
	}
	if err == io.ErrUnexpectedEOF {
		// Got an unexpected error -- could be due to a short read
		state = TarERR
	}
	//fmt.Printf("reading [%d]: %s\n", hdr.Size, hdr.Name)
	return state, bytesToRead, int64(blen - r.Len()), &f, err
}

// TarStreamRead returns a tar state, number of byte read, number of bytes remain in buffer and error
// Depending on what we see in the steam, the next state is set. Ideally, we just handle that in the
// TarReader function
func TarStreamRead(r *bytes.Reader, f *TarFile, min int64) (TarState, int64, int64, error) {
	var state TarState = TarEOF
	var blen int = r.Len()
	b, err := TarBody(r, min)
	if int64(len(b)) != min {
		// Did not read the required min. Ask for more. Return Partial state
		state = Partial
	}
	if err == io.ErrUnexpectedEOF {
		// Got an unexpected error -- could be due to a short read
		state = TarERR
	}
	f.body = append(f.body, b...)
	return state, (min - int64(len(b))), int64(blen - r.Len()), err
}

// TarReader is the heart of tar state machine. It return the list of tar files received on the pipe
// It reads the buffer from the pipe, processes all the byte
// if there is an incomplete buffer, it goes around to read from the pipe
// TDOD: only read the header from the ReadAll then read the size from the hdr
//       That's the only way to make sure we don't buffer too much
func TarReader(out *os.File, logger logs.TaggedLogger) ([]TarFile, error) {
	// Tar state machine
	//		NewTar -> start of a new Tar Stream
	//		Partial -> existing tar stream has not been full received
	//		TarEOF -> a Tar stream is completely read
	//		TarERR -> Error state. Special handling for Unexpected EOF
	var sstate TarState = NewTar
	var bytesToRead int64 = 0
	var done bool = false
	var filesReceived []TarFile
	var filesCnt int = 0
	for !done {
		var b []byte
		// Read from the pipe
		b, err := ioutil.ReadAll(out)
		if err != nil {
			if err != io.EOF {
				return filesReceived, err
			}
			continue
		}
		if len(b) == 0 {
			// Need it for ReadAll
			continue
		}

		r := bytes.NewReader(b)
		var bytesRemain int64 = int64(len(b))
		var f *TarFile
		for bytesRemain != 0 {
			switch sstate {
			case NewTar:
				sstate, bytesToRead, bytesRemain, f, err = NewTarStream(r)
			case Partial:
				sstate, bytesToRead, bytesRemain, err = TarStreamRead(r, f, bytesToRead)
			case TarEOF:
				sstate = NewTar
				if f != nil {
					filesReceived[filesCnt] = *f
					filesCnt++
				}
				if bytesRemain == 0 {
					done = true
				}
			case TarERR:
				switch err {
				case io.ErrUnexpectedEOF:
					sstate = NewTar
					if bytesToRead != 0 {
						sstate = Partial
					}
				default:
					return filesReceived, err
				}

			default:
			}
		}
	}
	logger.Logf("Received %d tar files\n", filesCnt)
	return filesReceived, io.EOF
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

// ReceiveSpout is used by both services and spouts if a spout is defined on the
// pipeline. ctx is separate from pachClient because services may call this, and
// they use a cancel function that affects the context but not the pachClient
// (so metadata updates can still be made while unwinding).
func ReceiveSpout(
	ctx context.Context,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
) (retErr1 error) {
	return backoff.RetryUntilCancel(ctx, func() error {
		repo := pipelineInfo.Pipeline.Name
		// open a read connection to the /pfs/out named pipe
		out, err := os.Open("/pfs/out")
		if err != nil {
			return err
		}
		// and close it when we're done
		defer func() {
			if err := out.Close(); err != nil && retErr1 == nil {
				// this lets us pass the error through if Close fails
				retErr1 = err
			}
		}()
		for {
			// this extra closure is so that we can scope the defer for FinishCommit
			files, err := TarReader(out, logger)
			if err == io.EOF {
				break
			}
			var commit *pfs.Commit
			for _, f := range files {
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
		}
		return err
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in receiveSpout: %+v, retrying in: %+v", err, d)
		return nil
	})
}

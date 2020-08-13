package logs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

const (
	logBuffer = 25
)

// TaggedLogger is an interface providing logging functionality for use by the
// worker, worker master, and user code processes
type TaggedLogger interface {
	// The TaggedLogger is usable as a io.Writer so that it can be set as the
	// user code's subprocess stdout/stderr and still provide the same guarantees
	io.Writer

	// Logf logs to stdout and object storage (if stats are enabled), including
	// metadata about the current pipeline and job
	Logf(formatString string, args ...interface{})
	// Errf logs only to stderr
	Errf(formatString string, args ...interface{})

	// LogStep will log before and after the given callback function runs, using
	// the name provided
	LogStep(name string, cb func() error) error

	// These helpers will clone the current logger and construct a new logger that
	// includes the given metadata in log messages.
	WithJob(jobID string) TaggedLogger
	WithData(data []*common.Input) TaggedLogger
	WithUserCode() TaggedLogger

	JobID() string

	// Close will flush any writes to object storage and return information about
	// where log statements were stored in object storage.
	Close() (*pfs.Object, int64, error)
}

type taggedLogger struct {
	template  pps.LogMessage
	stderrLog *log.Logger
	marshaler *jsonpb.Marshaler

	// Used for mirroring log statements to object storage
	putObjClient pfs.ObjectAPI_PutObjectClient
	objSize      int64
	msgCh        chan string
	buffer       bytes.Buffer
	eg           errgroup.Group
}

func newLogger(pipelineInfo *pps.PipelineInfo) *taggedLogger {
	name := ""
	if pipelineInfo != nil {
		name = pipelineInfo.Pipeline.Name
	}

	return &taggedLogger{
		template: pps.LogMessage{
			PipelineName: name,
			WorkerID:     os.Getenv(client.PPSPodNameEnv),
		},
		stderrLog: log.New(os.Stderr, "", log.LstdFlags|log.Llongfile),
		marshaler: &jsonpb.Marshaler{},
		msgCh:     make(chan string, logBuffer),
	}
}

// NewLogger constructs a TaggedLogger for the given pipeline, optionally
// including mirroring log statements to object storage.
//
// If a pachClient is passed in and stats are enabled on the pipeline, log
// statements will be chunked and written to the Object API on the given
// client.
//
// TODO: the whole object api interface here is bad, there are a few shortcomings:
//  - it's only used under the worker function when stats are enabled
//  - 'Close' is used to end the object, but it must be explicitly called
//  - the 'eg', 'putObjClient', 'msgCh', 'buffer', and 'objSize' don't play well
//      with cloned loggers.
// Abstract this into a separate object with a more explicit lifetime?
func NewLogger(pipelineInfo *pps.PipelineInfo, pachClient *client.APIClient) (TaggedLogger, error) {
	result := newLogger(pipelineInfo)

	if pachClient != nil && pipelineInfo.EnableStats {
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
					}); err != nil && !errors.Is(err, io.EOF) {
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

// NewStatlessLogger constructs a TaggedLogger for the given pipeline.  This is
// typically used outside of the worker/master main path.
func NewStatlessLogger(pipelineInfo *pps.PipelineInfo) TaggedLogger {
	return newLogger(pipelineInfo)
}

// NewMasterLogger constructs a TaggedLogger for the given pipeline that
// includes the 'Master' flag.  This is typically used for all logging from the
// worker master goroutine. The 'User' flag is set to false to maintain the
// invariant that 'User' and 'Master' are mutually exclusive, which is done to
// make it easier to query for specific logs.
func (logger *taggedLogger) WithUserCode() TaggedLogger {
func NewMasterLogger(pipelineInfo *pps.PipelineInfo) TaggedLogger {
	result := newLogger(pipelineInfo) // master loggers don't log stats
	result.template.Master = true
	result.template.User = false
	return result
}

// WithJob clones the current logger and returns a new one that will include
// the given job ID in log statement metadata.
func (logger *taggedLogger) WithJob(jobID string) TaggedLogger {
	result := logger.clone()
	result.template.JobID = jobID
	return result
}

// WithJob clones the current logger and returns a new one that will include
// the given data inputs in log statement metadata.
func (logger *taggedLogger) WithData(data []*common.Input) TaggedLogger {
	result := logger.clone()

	// Add inputs' details to log metadata, so we can find these logs later
	result.template.Data = []*pps.InputFile{}
	for _, d := range data {
		result.template.Data = append(result.template.Data, &pps.InputFile{
			Path: d.FileInfo.File.Path,
			Hash: d.FileInfo.Hash,
		})
	}

	// This is the same ID used in the stats tree for the datum
	result.template.DatumID = common.DatumID(data)
	return result
}

// WithUserCode clones the current logger and returns a new one that will
// include the 'User' flag in log statement metadata. The 'Master' flag is set
// to false to maintain the invariant that 'User' and 'Master' are mutually
// exclusive, which is done to make it easier to query for specific logs.
func (logger *taggedLogger) WithUserCode() TaggedLogger {
	result := logger.clone()
	result.template.User = true
	result.template.Master = false
	return result
}

// JobID returns the current job that the logger is configured with.
func (logger *taggedLogger) JobID() string {
	return logger.template.JobID
}

func (logger *taggedLogger) clone() *taggedLogger {
	return &taggedLogger{
		template:     logger.template,  // Copy struct
		stderrLog:    logger.stderrLog, // logger should be goroutine-safe
		marshaler:    &jsonpb.Marshaler{},
		putObjClient: logger.putObjClient,
		msgCh:        logger.msgCh,
	}
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
		logger.Errf("could not generate logging timestamp: %s\n", err)
		return
	}
	msg, err := logger.marshaler.MarshalToString(&logger.template)
	if err != nil {
		logger.Errf("could not marshal %v for logging: %s\n", &logger.template, err)
		return
	}
	fmt.Println(msg)
	if logger.putObjClient != nil {
		logger.msgCh <- msg + "\n"
	}
}

// LogStep will log before and after the given callback function runs, using
// the name provided
func (logger *taggedLogger) LogStep(name string, cb func() error) (retErr error) {
	logger.Logf("started %v", name)
	defer func() {
		if retErr != nil {
			logger.Logf("errored %v: %v", name, retErr)
		} else {
			logger.Logf("finished %v", name)
		}
	}()
	return cb()
}

// Errf writes the given line to the stderr of the worker process.  This does
// not go to a persistent log.
func (logger *taggedLogger) Errf(formatString string, args ...interface{}) {
	logger.stderrLog.Printf(formatString, args...)
}

// This is provided so that taggedLogger can be used as a io.Writer for stdout
// and stderr when running user code.
func (logger *taggedLogger) Write(p []byte) (_ int, retErr error) {
	// never errors
	logger.buffer.Write(p)
	r := bufio.NewReader(&logger.buffer)
	for {
		message, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.buffer.Write([]byte(message))
				return len(p), nil
			}
			// this shouldn't technically be possible to hit - io.EOF should be
			// the only error bufio.Reader can return when using a buffer.
			return 0, errors.Wrap(err, "ReadString")
		}
		// We don't want to make this call as:
		// logger.Logf(message)
		// because if the message has format characters like %s in it those
		// will result in errors being logged.
		logger.Logf("%s", strings.TrimSuffix(message, "\n"))
	}
}

// Close flushes and closes the object storage client used to mirror log
// statements to object storage.  Returns a pointer to the generated pfs.Object
// as well as the total size of all written messages.
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

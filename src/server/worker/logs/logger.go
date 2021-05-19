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

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
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
	WithPipelineJob(pipelineJobID string) TaggedLogger
	WithData(data []*common.Input) TaggedLogger
	WithUserCode() TaggedLogger

	PipelineJobID() string
}

type taggedLogger struct {
	template  pps.LogMessage
	stderrLog *log.Logger
	marshaler *jsonpb.Marshaler

	buffer bytes.Buffer
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
func NewMasterLogger(pipelineInfo *pps.PipelineInfo) TaggedLogger {
	result := newLogger(pipelineInfo) // master loggers don't log stats
	result.template.Master = true
	result.template.User = false
	return result
}

// WithPipelineJob clones the current logger and returns a new one that will include
// the given job ID in log statement metadata.
func (logger *taggedLogger) WithPipelineJob(pipelineJobID string) TaggedLogger {
	result := logger.clone()
	result.template.PipelineJobID = pipelineJobID
	return result
}

// WithData clones the current logger and returns a new one that will include
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

// PipelineJobID returns the current job that the logger is configured with.
func (logger *taggedLogger) PipelineJobID() string {
	return logger.template.PipelineJobID
}

func (logger *taggedLogger) clone() *taggedLogger {
	return &taggedLogger{
		template:  logger.template,  // Copy struct
		stderrLog: logger.stderrLog, // logger should be goroutine-safe
		marshaler: &jsonpb.Marshaler{},
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

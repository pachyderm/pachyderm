package logs

import (
	"context"
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"go.uber.org/zap"
)

// TaggedLogger is an interface providing logging functionality for use by the
// worker, worker master, and user code processes
type TaggedLogger interface {
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
	Writer(string) io.WriteCloser
}

// taggedLogger is a port of the legacy worker logger to one compatible with the rest of Pachyderm.
type taggedLogger struct {
	// Context ONLY FOR LOGGING.  This only exists because the worker predates context.Context,
	// and we don't have enough tests for a safe refactor.
	ctx   context.Context
	jobID string // Yup, we also use the logger to pass around state.
}

// New adapts context-based logging to the worker.
//
// NOTE(jonathan): This is kind of a workaround to switch over to new logging without any risk of
// breaking the worker.  The logging flow and context flow in the worker are not compatible with
// each other; I tried to fix this, but each little tweak made a different test break in a different
// way.  To properly fix the context situation in the worker to enable clean logging, we'll have to
// get a lot more test coverage.
func New(ctx context.Context) *taggedLogger {
	// It is expected that this context come from the one created in NewWorker; it already
	// contains projectId, pipelineName, and workerId.
	return &taggedLogger{
		ctx: ctx,
	}
}

// NewMasterLogger constructs a TaggedLogger for the given pipeline that includes the 'Master' flag.
// The root context should come from worker.NewWorker, with projectId, pipelineName, and workerId
// already set.
func NewMasterLogger(ctx context.Context) TaggedLogger {
	return &taggedLogger{
		ctx: pctx.Child(ctx, "", pctx.WithFields(pps.MasterField(true))),
	}
}

// WithJob clones the current logger and returns a new one that will include
// the given job ID in log statement metadata.
//
// This is also where we turn off log rate limiting, though WithData actually seems like a slightly
// more aggressive but still safe choice.  If there is a lot of log spam with jobId but not datumId,
// move it down there.  We have tests that test this!
func (logger *taggedLogger) WithJob(jobID string) TaggedLogger {
	return &taggedLogger{
		jobID: jobID,
		ctx:   pctx.Child(logger.ctx, "", pctx.WithFields(pps.JobIDField(jobID)), pctx.WithoutRatelimit()),
	}
}

// WithData clones the current logger and returns a new one that will include
// the given data inputs in log statement metadata.
func (logger *taggedLogger) WithData(data []*common.Input) TaggedLogger {
	var inputFiles []*pps.InputFile
	for _, d := range data {
		inputFiles = append(inputFiles, &pps.InputFile{
			Path: d.FileInfo.File.Path,
			Hash: d.FileInfo.Hash,
		})
	}
	return &taggedLogger{
		jobID: logger.jobID,
		ctx:   pctx.Child(logger.ctx, "", pctx.WithFields(pps.DataField(inputFiles), pps.DatumIDField(common.DatumID(data)))),
	}
}

// WithUserCode clones the current logger and returns a new one that will
// include the 'User' flag in log statement metadata.
func (logger *taggedLogger) WithUserCode() TaggedLogger {
	return &taggedLogger{
		jobID: logger.jobID,
		ctx:   pctx.Child(logger.ctx, "", pctx.WithFields(pps.UserField(true))),
	}
}

// JobID returns the current job that the logger is configured with.
func (logger *taggedLogger) JobID() string {
	return logger.jobID
}

// Logf logs the line Sprintf(formatString, args...), with any data added by previous WithXXX calls.
func (logger *taggedLogger) Logf(formatString string, args ...any) {
	log.Info(logger.ctx, fmt.Sprintf(formatString, args...))
}

// LogStep will log before and after the given callback function runs, using
// the name provided.
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

// Errf writes the given line as an error.
func (logger *taggedLogger) Errf(formatString string, args ...any) {
	log.Error(logger.ctx, fmt.Sprintf(formatString, args...))
}

// Writer returns an io.WriteCloser that logs each line to the logger.  The logs are NOT
// rate-limited, even if the parent logger is.
func (logger *taggedLogger) Writer(stream string) io.WriteCloser {
	return log.WriterAt(pctx.Child(logger.ctx, "", pctx.WithFields(zap.String("stream", stream)), pctx.WithoutRatelimit()), log.InfoLevel)
}

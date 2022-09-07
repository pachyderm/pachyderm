package miscutil

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	log "github.com/sirupsen/logrus"
)

// LogStep logs how long it takes to perform an operation.  If ctx comes from a
// gRPC method intercepted by a logging.LoggingInterceptor, the method name will
// be included in the structured log.
//
// TODO: refactor into a common logging utility.
func LogStep(ctx context.Context, logger *log.Entry, name string, cb func() error) (retErr error) {
	start := time.Now()
	methodName, ok := logging.MethodNameFromContext(ctx)
	if ok {
		logger = logger.WithField("method", methodName)
	}
	logger.Infof("started %v", name)
	defer func() {
		duration := time.Since(start)
		if retErr != nil {
			logger.WithFields(log.Fields{
				"duration": duration,
				"error":    retErr,
			}).Errorf("errored %v", name)
		} else {
			logger.WithField("duration", duration).Infof("finished %v", name)
		}
	}()
	return cb()
}

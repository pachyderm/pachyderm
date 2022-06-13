package miscutil

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// TODO: refactor into a common logging utility.
func LogStep(name string, cb func() error) (retErr error) {
	start := time.Now()
	log.Infof("started %v", name)
	defer func() {
		duration := time.Since(start)
		if retErr != nil {
			log.WithFields(log.Fields{
				"duration": duration,
				"error":    retErr,
			}).Errorf("errored %v", name)
		} else {
			log.WithField("duration", duration).Infof("finished %v", name)
		}
	}()
	return cb()
}

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
			log.Errorf("errored %v: %v (duration: %v)", name, retErr, duration)
		} else {
			log.Infof("finished %v (duration: %v)", name, duration)
		}
	}()
	return cb()
}

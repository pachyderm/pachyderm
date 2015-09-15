package dockervolume

import "go.pedge.io/protolog"

var (
	loggerInstance = &logger{}
)

type logger struct{}

func (l *logger) LogCall(call *Call) {
	if call.Error != "" {
		protolog.Error(call)
	} else {
		protolog.Info(call)
	}
}

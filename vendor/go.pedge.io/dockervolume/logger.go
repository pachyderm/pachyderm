package dockervolume

import "go.pedge.io/protolog"

var (
	loggerInstance = &logger{}
)

type logger struct{}

func (l *logger) LogMethodInvocation(methodInvocation *MethodInvocation) {
	if methodInvocation.Error != "" {
		protolog.Error(methodInvocation)
	} else {
		protolog.Info(methodInvocation)
	}
}

package dockervolume

import "go.pedge.io/protolog"

var (
	loggerInstance = &logger{}
)

type logger struct{}

func (l *logger) LogMethodInvocation(methodInvocation *MethodInvocation) {
	protolog.Info(methodInvocation)
}

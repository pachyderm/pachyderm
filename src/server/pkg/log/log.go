package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Log(request interface{}, response interface{}, err error, duration time.Duration)
}

type logger struct {
	*logrus.Entry
}

func NewLogger(service string) Logger {
	l := logrus.New()
	logrus.SetFormatter(new(PrettyFormatter))
	return &logger{
		l.WithFields(logrus.Fields{"service": service}),
	}
}

func (l *logger) Log(request interface{}, response interface{}, err error, duration time.Duration) {
	depth := 2
	pc := make([]uintptr, 2+depth)
	runtime.Callers(2+depth, pc)
	split := strings.Split(runtime.FuncForPC(pc[0]).Name(), ".")
	method := split[len(split)-1]

	entry := l.WithFields(
		logrus.Fields{
			"method":   method,
			"request":  request,
			"response": response,
			"error":    err,
			"duration": duration,
		},
	)

	if err != nil {
		entry.Error(err)
	} else {
		entry.Info()
	}
}

type PrettyFormatter struct {
}

func (f *PrettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	serialized := []byte(
		fmt.Sprintf(
			"%v %v %v.%v",
			entry.Time,
			strings.ToUpper(entry.Level.String()),
			entry.Data["service"],
			entry.Data["method"],
		),
	)
	delete(entry.Data, "service")
	delete(entry.Data, "method")
	data, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}

	serialized = append(serialized, []byte(fmt.Sprintf("%v %v\n", data, entry.Message))...)
	return serialized, nil
}

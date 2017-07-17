package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is a helper for emitting our grpc API logs
type Logger interface {
	Log(request interface{}, response interface{}, err error, duration time.Duration)
}

type logger struct {
	*logrus.Entry
}

// NewLogger creates a new logger
func NewLogger(service string) Logger {
	l := logrus.New()
	l.Formatter = new(prettyFormatter)
	return &logger{
		l.WithFields(logrus.Fields{"service": service}),
	}
}

func (l *logger) Log(request interface{}, response interface{}, err error, duration time.Duration) {
	depth := 1
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

type prettyFormatter struct {
}

func (f *prettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	serialized := []byte(
		fmt.Sprintf(
			"%v %v ",
			entry.Time.Format(logrus.DefaultTimestampFormat),
			strings.ToUpper(entry.Level.String()),
		),
	)
	if entry.Data["service"] != nil {
		serialized = append(serialized, []byte(fmt.Sprintf("%v.%v ", entry.Data["service"], entry.Data["method"]))...)
	}
	if len(entry.Data) > 2 {
		delete(entry.Data, "service")
		delete(entry.Data, "method")
		if entry.Data["duration"] != nil {
			entry.Data["duration"] = entry.Data["duration"].(time.Duration).Seconds()
		}
		data, err := json.Marshal(entry.Data)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
		}
		serialized = append(serialized, []byte(string(data))...)
		serialized = append(serialized, ' ')
	}

	serialized = append(serialized, []byte(entry.Message)...)
	serialized = append(serialized, '\n')
	return serialized, nil
}

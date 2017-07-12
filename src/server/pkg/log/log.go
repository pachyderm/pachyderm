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
	return &logger{
		l.WithFields(logrus.Fields{"service": service}),
	}
}

func (l *logger) Log(request interface{}, response interface{}, err error, duration time.Duration) {
	logrus.SetFormatter(new(PrettyFormatter))
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
	fmt.Printf("OMG CALLING PRETTY FORMATTER\n")
	serialized := []byte(
		fmt.Sprintf(
			"%v %v",
			entry.Time,
			strings.ToUpper(entry.Level.String()),
		),
	)
	if entry.Data["service"] != nil {
		serialized = append(serialized, []byte(fmt.Sprintf(" %v.%v", entry.Data["service"], entry.Data["method"]))...)
	}
	delete(entry.Data, "service")
	delete(entry.Data, "method")
	data, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}

	serialized = append(serialized, []byte(fmt.Sprintf("%v %v\n", string(data), entry.Message))...)
	fmt.Printf("should be printing:\n%v\n", string(serialized))
	return serialized, nil
}

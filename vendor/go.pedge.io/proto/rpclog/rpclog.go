package protorpclog
import (
	"time"

	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"

	"github.com/golang/protobuf/proto"
)

// Debug logs an RPC call at the debug level.
func Debug(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolog.Debug(event(serviceName, methodName, request, response, err, duration))
}

// Info logs an RPC call at the info level.
func Info(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolog.Info(event(serviceName, methodName, request, response, err, duration))
}

// Warn logs an RPC call at the warn level.
func Warn(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolog.Warn(event(serviceName, methodName, request, response, err, duration))
}

// Error logs an RPC call at the error level.
func Error(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolog.Error(event(serviceName, methodName, request, response, err, duration))
}

func event(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) *Call {
	call := &Call{
		Service:  serviceName,
		Method:   methodName,
		Duration: prototime.DurationToProto(duration),
	}
	if request != nil {
		call.Request = request.String()
	}
	if response != nil {
		call.Response = response.String()
	}
	if err != nil {
		call.Error = err.Error()
	}
	return call
}

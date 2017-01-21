package protorpclog //import "go.pedge.io/proto/rpclog"
import (
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"

	"github.com/golang/protobuf/proto"
)

var (
	// LoggingUnaryServerInterceptor is a grpc.UnaryServerInterceptor that logs every request and response.
	LoggingUnaryServerInterceptor = newLoggingUnaryServerInterceptor()
)

// Logger is a logger intended to be used as such:
//
// type apiServer struct {
//   protorpclog.Logger
//   ...
// }
//
// func (a *apiServer) Foo(ctx Context.Context, request *FooRequest) (response *FooResponse, err error) {
//   defer func(start time.Now) { a.Log(request, response, err, time.Since(start)) }(time.Now())
//   ...
// }
type Logger interface {
	Log(request proto.Message, response proto.Message, err error, duration time.Duration)
}

// NewLogger returns a new Logger.
func NewLogger(serviceName string) Logger {
	return newLogger(serviceName)
}

type logger struct {
	serviceName string
}

func newLogger(serviceName string) *logger {
	return &logger{serviceName}
}

func (l *logger) Log(request proto.Message, response proto.Message, err error, duration time.Duration) {
	Log(l.serviceName, getMethodName(2), request, response, err, duration)
}

// Log logs an RPC call at the info level if no error, or at the error level if error.
func Log(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	if err != nil {
		Error(serviceName, methodName, request, response, err, duration)
	} else {
		Info(serviceName, methodName, request, response, err, duration)
	}
}

// Debug logs an RPC call at the debug level.
func Debug(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolion.Debug(event(serviceName, methodName, request, response, err, duration))
}

// Info logs an RPC call at the info level.
func Info(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolion.Info(event(serviceName, methodName, request, response, err, duration))
}

// Warn logs an RPC call at the warn level.
func Warn(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolion.Warn(event(serviceName, methodName, request, response, err, duration))
}

// Error logs an RPC call at the error level.
func Error(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) {
	protolion.Error(event(serviceName, methodName, request, response, err, duration))
}

func newLoggingUnaryServerInterceptor() func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
	return func(
		ctx context.Context,
		request interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var protoRequest proto.Message
		var protoResponse proto.Message
		var ok bool
		if request != nil {
			protoRequest, ok = request.(proto.Message)
			if !ok {
				return handler(ctx, request)
			}
		}
		start := time.Now()
		response, err := handler(ctx, request)
		duration := time.Since(start)
		if response != nil {
			protoResponse, ok = response.(proto.Message)
			if !ok {
				return response, err
			}
		}
		split := strings.Split(info.FullMethod, "/")
		if len(split) != 3 {
			return response, err
		}
		Log(
			split[1],
			split[2],
			protoRequest,
			protoResponse,
			err,
			duration,
		)
		return response, err
	}
}

func event(serviceName string, methodName string, request proto.Message, response proto.Message, err error, duration time.Duration) *Call {
	call := &Call{
		Service:  serviceName,
		Method:   methodName,
		Duration: google_protobuf.DurationToProto(duration),
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

func getMethodName(depth int) string {
	pc := make([]uintptr, 2+depth)
	runtime.Callers(2+depth, pc)
	split := strings.Split(runtime.FuncForPC(pc[0]).Name(), ".")
	return split[len(split)-1]
}

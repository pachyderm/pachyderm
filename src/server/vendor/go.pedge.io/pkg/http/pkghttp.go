/*
Package pkghttp defines common functionality for http.
*/
package pkghttp // import "go.pedge.io/pkg/http"

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"go.pedge.io/env"
	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/pb/go/pb/money"
	"go.pedge.io/pkg/tmpl"
)

var (
	// ErrRequireHandler is the error returned if handler is not set.
	ErrRequireHandler = errors.New("pkghttp: handler must be set")
)

// HandlerEnv is the environment for a handler.
type HandlerEnv struct {
	// The port to serve on.
	Port uint16 `env:"HTTP_PORT,default=8080"`
	// HealthCheckPath is the path for health checking.
	// This path will always return 200 for a GET.
	// Default value is /health.
	HealthCheckPath string `env:"HTTP_HEALTH_CHECK_PATH,default=/health"`
	// The time in seconds to shutdown after a SIGINT or SIGTERM.
	// Default value is 10.
	ShutdownTimeoutSec uint64 `env:"HTTP_SHUTDOWN_TIMEOUT_SEC,default=10"`
}

// GetHandlerEnv gets the HandlerEnv from the environment.
func GetHandlerEnv() (HandlerEnv, error) {
	handlerEnv := HandlerEnv{}
	if err := env.Populate(&handlerEnv); err != nil {
		return HandlerEnv{}, err
	}
	return handlerEnv, nil
}

// NewWrapperHandler returns a new wrapper handler.
func NewWrapperHandler(delegate http.Handler, handlerEnv HandlerEnv) http.Handler {
	return newWrapperHandler(delegate, handlerEnv)
}

// ListenAndServe is the equivalent to http's method.
//
// Intercepts requests and responses, handles SIGINT and SIGTERM.
// When this returns, any errors will have been logged.
// If the server starts, this will block until the server stops.
//
// Uses a wrapper handler.
func ListenAndServe(handler http.Handler, handlerEnv HandlerEnv) error {
	if handler == nil {
		return handleErrorBeforeStart(ErrRequireHandler)
	}
	if handlerEnv.Port == 0 {
		handlerEnv.Port = 8080
	}
	if handlerEnv.HealthCheckPath == "" {
		handlerEnv.HealthCheckPath = "/health"
	}
	if handlerEnv.ShutdownTimeoutSec == 0 {
		handlerEnv.ShutdownTimeoutSec = 10
	}
	server := &graceful.Server{
		Timeout: time.Duration(handlerEnv.ShutdownTimeoutSec) * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", handlerEnv.Port),
			Handler: NewWrapperHandler(
				handler,
				handlerEnv,
			),
		},
	}
	protolion.Info(
		&ServerStarting{
			Port: uint32(handlerEnv.Port),
		},
	)
	start := time.Now()
	err := server.ListenAndServe()
	serverFinished := &ServerFinished{
		Duration: google_protobuf.DurationToProto(time.Since(start)),
	}
	if err != nil {
		serverFinished.Error = err.Error()
		protolion.Error(serverFinished)
		return err
	}
	protolion.Info(serverFinished)
	return nil
}

// GetAndListenAndServe is GetHandlerEnv then ListenAndServe.
func GetAndListenAndServe(handler http.Handler) error {
	handlerEnv, err := GetHandlerEnv()
	if err != nil {
		return err
	}
	return ListenAndServe(handler, handlerEnv)
}

// Templater is a pkgtmpl.Templater for http responses.
type Templater interface {
	WithFuncs(funcMap template.FuncMap) Templater
	Execute(responseWriter http.ResponseWriter, name string, data interface{})
}

// NewTemplater returns a new Templater.
func NewTemplater(baseDirPath string) Templater {
	return newTemplater(pkgtmpl.NewTemplater(baseDirPath))
}

// Error does http.Error on the error if not nil, and returns true if not nil.
func Error(responseWriter http.ResponseWriter, statusCode int, err error) bool {
	if err != nil {
		http.Error(responseWriter, err.Error(), statusCode)
		return true
	}
	return false
}

// ErrorInternal does Error 500.
func ErrorInternal(responseWriter http.ResponseWriter, err error) bool {
	return Error(responseWriter, http.StatusInternalServerError, err)
}

// FileHandlerFunc returns a handler function for a file path.
func FileHandlerFunc(filePath string) func(http.ResponseWriter, *http.Request) {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		http.ServeFile(responseWriter, request, filePath)
	}
}

// QueryGet gets the string by key from the request query, if it exists.
// Otherwise, returns "".
func QueryGet(request *http.Request, key string) string {
	return request.URL.Query().Get(key)
}

// QueryGetUint32 gets the uint32 by key from the request query, if it exists.
// Otherwise, returns 0.
// error returned if there is a parsing error.
func QueryGetUint32(request *http.Request, key string) (uint32, error) {
	valueString := QueryGet(request, key)
	if valueString == "" {
		return 0, nil
	}
	valueString = strings.Replace(valueString, ",", "", -1)
	value, err := strconv.ParseUint(valueString, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(value), nil
}

// QueryGetFloat64 gets the float64 by key from the request query, if it exists.
// Otherwise, returns 0.0.
// error returned if there is a parsing error.
func QueryGetFloat64(request *http.Request, key string) (float64, error) {
	valueString := QueryGet(request, key)
	if valueString == "" {
		return 0.0, nil
	}
	valueString = strings.Replace(valueString, ",", "", -1)
	return strconv.ParseFloat(valueString, 64)
}

// QueryGetMoney gets the money by key from the request query, if it exists.
// Otherwise, returns nil.
// Money is expected to be in float notation, ie $123.45 is 123.45.
// error returned if there is a parsing error.
func QueryGetMoney(request *http.Request, key string) (*pbmoney.Money, error) {
	valueDollars, err := QueryGetFloat64(request, key)
	if err != nil {
		return nil, err
	}
	if valueDollars == 0.0 {
		return nil, nil
	}
	return pbmoney.NewMoneyFloatUSD(valueDollars), nil
}

func handleErrorBeforeStart(err error) error {
	protolion.Error(
		&ServerCouldNotStart{
			Error: err.Error(),
		},
	)
	return err
}

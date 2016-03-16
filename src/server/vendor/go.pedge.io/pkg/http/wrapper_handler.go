package pkghttp

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
)

type wrapperHandler struct {
	http.Handler
	healthCheckPath string
}

func newWrapperHandler(handler http.Handler, handlerEnv HandlerEnv) *wrapperHandler {
	wrapperHandler := &wrapperHandler{handler, handlerEnv.HealthCheckPath}
	if wrapperHandler.healthCheckPath == "" {
		wrapperHandler.healthCheckPath = "/health"
	}
	return wrapperHandler
}

func (h *wrapperHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	start := time.Now()
	wrapperResponseWriter := newWrapperResponseWriter(responseWriter)
	defer func() {
		call := &Call{
			Method:         request.Method,
			RequestHeader:  valuesMap(request.Header),
			RequestForm:    valuesMap(request.Form),
			ResponseHeader: valuesMap(wrapperResponseWriter.Header()),
			StatusCode:     uint32(statusCode(wrapperResponseWriter.statusCode)),
			Error:          errorString(wrapperResponseWriter.writeError),
		}
		if request.URL != nil {
			call.Path = request.URL.Path
			call.Query = valuesMap(request.URL.Query())
		}
		if recoverErr := recover(); recoverErr != nil {
			// TODO(pedge): should we write anything at all?
			responseWriter.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, 8192)
			stack = stack[:runtime.Stack(stack, false)]
			call.Error = fmt.Sprintf("panic: %v\n%s", recoverErr, string(stack))
		}
		call.Duration = google_protobuf.DurationToProto(time.Since(start))
		protolion.Info(call)
	}()
	if request.URL != nil && request.URL.Path == h.healthCheckPath {
		wrapperResponseWriter.WriteHeader(http.StatusOK)
		return
	}
	h.Handler.ServeHTTP(wrapperResponseWriter, request)
}

func valuesMap(values map[string][]string) map[string]string {
	if values == nil {
		return nil
	}
	m := make(map[string]string)
	for key, value := range values {
		m[key] = strings.Join(value, " ")
	}
	return m
}

func statusCode(code int) int {
	if code == 0 {
		return http.StatusOK
	}
	return code
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

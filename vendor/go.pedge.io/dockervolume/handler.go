package dockervolume

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	contentType = "application/vnd.docker.plugins.v1.1+json"
)

var (
	activateResponse = []byte("{\"Implements\": [\"VolumeDriver\"]}\n")
)

type handler struct {
	*http.ServeMux
}

func newHandler(volumeDriver VolumeDriver, opts HandlerOptions) *handler {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc(
		"/Plugin.Activate",
		func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Content-Type", contentType)
			_, _ = responseWriter.Write(activateResponse)
		},
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Create",
		newGenericHandlerFunc(
			Method_METHOD_CREATE,
			func(name string, opts map[string]string) (string, error) {
				return "", volumeDriver.Create(name, opts)
			},
			opts,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Remove",
		newGenericHandlerFunc(
			Method_METHOD_REMOVE,
			func(name string, opts map[string]string) (string, error) {
				return "", volumeDriver.Remove(name)
			},
			opts,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Mount",
		newGenericHandlerFunc(
			Method_METHOD_MOUNT,
			func(name string, opts map[string]string) (string, error) {
				return volumeDriver.Mount(name)
			},
			opts,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Path",
		newGenericHandlerFunc(
			Method_METHOD_PATH,
			func(name string, opts map[string]string) (string, error) {
				return volumeDriver.Path(name)
			},
			opts,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Unmount",
		newGenericHandlerFunc(
			Method_METHOD_UNMOUNT,
			func(name string, opts map[string]string) (string, error) {
				return "", volumeDriver.Unmount(name)
			},
			opts,
		),
	)
	return &handler{
		serveMux,
	}
}

func newGenericHandlerFunc(
	method Method,
	f func(string, map[string]string) (string, error),
	opts HandlerOptions,
) func(http.ResponseWriter, *http.Request) {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		m := make(map[string]interface{})
		if err := json.NewDecoder(request.Body).Decode(&m); err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			return
		}
		responseWriter.Header().Set("Content-Type", contentType)
		_ = json.NewEncoder(responseWriter).Encode(wrap(getLogger(opts), method, f, m))
	}
}

func wrap(
	logger Logger,
	method Method,
	f func(string, map[string]string) (string, error),
	request map[string]interface{},
) map[string]interface{} {
	name, opts, err := extractParameters(request)
	call := &Call{
		Method: method,
		Name:   name,
		Opts:   opts,
	}
	if err != nil {
		return handleResponse(logger, call, "", err)
	}
	mountpoint, err := f(name, opts)
	return handleResponse(logger, call, mountpoint, err)
}

func extractParameters(request map[string]interface{}) (string, map[string]string, error) {
	if err := checkRequiredParameters(request, "Name"); err != nil {
		return "", nil, err
	}
	name := request["Name"].(string)
	var opts map[string]string
	if _, ok := request["Opts"]; ok {
		opts = mapStringInterfaceToMapStringString(request["Opts"].(map[string]interface{}))
	}
	return name, opts, nil
}

func handleResponse(logger Logger, call *Call, mountpoint string, err error) map[string]interface{} {
	response := make(map[string]interface{})
	if mountpoint != "" {
		response["Mountpoint"] = mountpoint
		call.Mountpoint = mountpoint
	}
	if err != nil {
		errString := err.Error()
		response["Err"] = errString
		call.Error = errString
	}
	logger.LogCall(call)
	return response
}

func checkRequiredParameters(request map[string]interface{}, parameters ...string) error {
	for _, parameter := range parameters {
		if _, ok := request[parameter]; !ok {
			return fmt.Errorf("required parameter %s not set", parameter)
		}
	}
	return nil
}

func getLogger(opts HandlerOptions) Logger {
	if opts.Logger != nil {
		return opts.Logger
	}
	return loggerInstance
}

func mapStringInterfaceToMapStringString(m map[string]interface{}) map[string]string {
	n := make(map[string]string)
	for key, value := range m {
		n[key] = fmt.Sprintf("%v", value)
	}
	return n
}

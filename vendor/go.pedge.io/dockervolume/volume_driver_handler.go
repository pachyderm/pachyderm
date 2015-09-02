package dockervolume

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	contentType = "application/vnd.docker.plugins.v1+json"
)

var (
	activateResponse = []byte("{\"Implements\": [\"VolumeDriver\"]}\n")
)

func newVolumeDriverHandler(volumeDriver VolumeDriver) *http.ServeMux {
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
			volumeDriverCreate,
			volumeDriver,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Remove",
		newGenericHandlerFunc(
			volumeDriverRemove,
			volumeDriver,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Mount",
		newGenericHandlerFunc(
			volumeDriverMount,
			volumeDriver,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Path",
		newGenericHandlerFunc(
			volumeDriverPath,
			volumeDriver,
		),
	)
	serveMux.HandleFunc(
		"/VolumeDriver.Unmount",
		newGenericHandlerFunc(
			volumeDriverUnmount,
			volumeDriver,
		),
	)
	return serveMux
}

func newGenericHandlerFunc(
	f func(VolumeDriver, map[string]interface{}) (map[string]interface{}, error),
	volumeDriver VolumeDriver,
) func(http.ResponseWriter, *http.Request) {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		m := make(map[string]interface{})
		if err := json.NewDecoder(request.Body).Decode(&m); err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			return
		}
		n, err := f(volumeDriver, m)
		if n == nil {
			n = make(map[string]interface{})
		}
		if err != nil {
			n["Err"] = err.Error()
		}
		responseWriter.Header().Set("Content-Type", contentType)
		_ = json.NewEncoder(responseWriter).Encode(n)
	}
}

func volumeDriverCreate(volumeDriver VolumeDriver, request map[string]interface{}) (map[string]interface{}, error) {
	if err := checkRequiredParameters(request, "Name"); err != nil {
		return nil, err
	}
	name := request["Name"].(string)
	var opts map[string]string
	if _, ok := request["Opts"]; ok {
		opts = make(map[string]string)
		for key, value := range request["Opts"].(map[string]interface{}) {
			opts[key] = fmt.Sprintf("%v", value)
		}
	}
	return nil, volumeDriver.Create(name, opts)
}

func volumeDriverRemove(volumeDriver VolumeDriver, request map[string]interface{}) (map[string]interface{}, error) {
	if err := checkRequiredParameters(request, "Name"); err != nil {
		return nil, err
	}
	return nil, volumeDriver.Remove(request["Name"].(string))
}

func volumeDriverMount(volumeDriver VolumeDriver, request map[string]interface{}) (map[string]interface{}, error) {
	if err := checkRequiredParameters(request, "Name"); err != nil {
		return nil, err
	}
	mountpoint, err := volumeDriver.Mount(request["Name"].(string))
	var n map[string]interface{}
	if mountpoint != "" {
		n = make(map[string]interface{})
		n["Mountpoint"] = mountpoint
	}
	return n, err
}

func volumeDriverPath(volumeDriver VolumeDriver, request map[string]interface{}) (map[string]interface{}, error) {
	if err := checkRequiredParameters(request, "Name"); err != nil {
		return nil, err
	}
	mountpoint, err := volumeDriver.Path(request["Name"].(string))
	var n map[string]interface{}
	if mountpoint != "" {
		n = make(map[string]interface{})
		n["Mountpoint"] = mountpoint
	}
	return n, err
}

func volumeDriverUnmount(volumeDriver VolumeDriver, request map[string]interface{}) (map[string]interface{}, error) {
	if err := checkRequiredParameters(request, "Name"); err != nil {
		return nil, err
	}
	return nil, volumeDriver.Unmount(request["Name"].(string))
}

func checkRequiredParameters(request map[string]interface{}, parameters ...string) error {
	for _, parameter := range parameters {
		if _, ok := request[parameter]; !ok {
			return fmt.Errorf("required parameter %s not set", parameter)
		}
	}
	return nil
}

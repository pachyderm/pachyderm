package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-yaml2json"
)

func ParsePipeline(dirPath string) (*pps.Pipeline, error) {
	configFilePath := filepath.Join(dirPath, "pps.yml")
	if err := checkFileExists(configFilePath); err != nil {
		return nil, err
	}
	data, err := getJSONFromYAMLFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	fmt.Println(m)
	return nil, nil
}

func getJSONFromYAMLFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return yaml2json.Transform(data, yaml2json.TransformOptions{})
}

func checkFileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

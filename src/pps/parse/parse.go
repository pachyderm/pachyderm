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

var (
	versionToParseFunc = map[string]func(string, *config) (*pps.Pipeline, error){
		"v1": parsePipelineV1,
	}
)

type config struct {
	Version string
	Root    string
	Ignore  []string
}

func ParsePipeline(dirPath string) (*pps.Pipeline, error) {
	config, err := parseConfig(dirPath)
	if err != nil {
		return nil, err
	}
	if config.Version == "" {
		return nil, fmt.Errorf("no version specified in pps.yml")
	}
	parseFunc, ok := versionToParseFunc[config.Version]
	if !ok {
		return nil, fmt.Errorf("unknown pps specification version: %s", config.Version)
	}
	return parseFunc(dirPath, config)
}

func parseConfig(dirPath string) (*config, error) {
	configFilePath := filepath.Join(dirPath, "pps.yml")
	if err := checkFileExists(configFilePath); err != nil {
		return nil, err
	}
	data, err := getJSONFromYAMLFile(configFilePath)
	if err != nil {
		return nil, err
	}
	config := &config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func parsePipelineV1(dirPath string, config *config) (*pps.Pipeline, error) {
	fmt.Println(config)
	return &pps.Pipeline{}, nil
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

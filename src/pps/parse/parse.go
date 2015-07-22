package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-yaml2json"
)

var (
	versionToParseFunc = map[string]func(string, *config) (*pps.Pipeline, error){
		"v1": parsePipeline,
	}
)

type config struct {
	Version string
	Include []string
	Exclude []string
}

type ppsMeta struct {
	Kind string
	Name string
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

func parsePipeline(dirPath string, config *config) (*pps.Pipeline, error) {
	dirPath = filepath.Clean(dirPath)
	filePaths, err := getAllFilePaths(dirPath, config.Include, config.Exclude)
	if err != nil {
		return nil, err
	}
	pipeline := &pps.Pipeline{
		NameToElement: make(map[string]*pps.Element),
	}
	for _, filePath := range filePaths {
		element, err := getElementForPipelineFile(dirPath, filePath)
		if err != nil {
			return nil, err
		}
		if _, ok := pipeline.NameToElement[element.Name]; ok {
			return nil, fmt.Errorf("duplicate element: %s", element.Name)
		}
		pipeline.NameToElement[element.Name] = element
	}
	return pipeline, nil
}

func getAllFilePaths(dirPath string, includes []string, excludes []string) ([]string, error) {
	var filePaths []string
	if err := filepath.Walk(
		dirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				filePaths = append(filePaths, path)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}
	relFilePaths := make([]string, len(filePaths))
	for i, filePath := range filePaths {
		relFilePath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return nil, err
		}
		relFilePaths[i] = relFilePath
	}
	var filteredRelFilePaths []string
	for _, relFilePath := range relFilePaths {
		isPipelineFile, err := isPipelineFile(relFilePath, includes, excludes)
		if err != nil {
			return nil, err
		}
		if isPipelineFile {
			filteredRelFilePaths = append(filteredRelFilePaths, relFilePath)
		}
	}
	return filteredRelFilePaths, nil
}

func isPipelineFile(filePath string, includes []string, excludes []string) (bool, error) {
	isPipelineFileIncluded, err := isPipelineFileIncluded(filePath, includes)
	if err != nil {
		return false, err
	}
	isPipelineFileExcluded, err := isPipelineFileExcluded(filePath, excludes)
	if err != nil {
		return false, err
	}
	return isPipelineFileIncluded && !isPipelineFileExcluded, nil
}

func isPipelineFileIncluded(filePath string, includes []string) (bool, error) {
	if !strings.HasSuffix(filePath, ".yml") {
		return false, nil
	}
	if filePath == "pps.yml" {
		return false, nil
	}
	if len(includes) == 0 {
		return true, nil
	}
	for _, include := range includes {
		matched, err := matches(include, filePath)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func isPipelineFileExcluded(filePath string, excludes []string) (bool, error) {
	if !strings.HasSuffix(filePath, ".yml") {
		return true, nil
	}
	for _, exclude := range excludes {
		matched, err := matches(exclude, filePath)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func matches(match string, filePath string) (bool, error) {
	if strings.HasPrefix(filePath, match) {
		return true, nil
	}
	matched, err := filepath.Match(match, filePath)
	if err != nil {
		return false, err
	}
	return matched, nil
}

func getElementForPipelineFile(dirPath string, relFilePath string) (*pps.Element, error) {
	filePath := filepath.Join(dirPath, relFilePath)
	data, err := getJSONFromYAMLFile(filePath)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	ppsMetaObj, ok := m["pps"]
	if !ok {
		return nil, fmt.Errorf("no pps section for %s", relFilePath)
	}
	ppsMeta := ppsMetaObj.(map[string]interface{})
	if ppsMeta["kind"] == "" {
		return nil, fmt.Errorf("no kind specified for %s", relFilePath)
	}
	if ppsMeta["name"] == "" {
		return nil, fmt.Errorf("no name specified for %s", relFilePath)
	}
	element := &pps.Element{
		Name: ppsMeta["name"].(string),
		Path: relFilePath,
	}
	switch ppsMeta["kind"] {
	case "node":
		node := &pps.Node{}
		if err := json.Unmarshal(data, node); err != nil {
			return nil, err
		}
		element.Node = node
		return element, nil
	case "docker_service":
		dockerService := &pps.DockerService{}
		if err := json.Unmarshal(data, dockerService); err != nil {
			return nil, err
		}
		element.DockerService = dockerService
		return element, nil
	default:
		return nil, fmt.Errorf("unknown kind: %v %s", ppsMeta["kind"], relFilePath)
	}
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

package parse

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/pachyderm/pachyderm/src/pps"
)

var (
	versionToParseFunc = map[string]func(string, *config) (*pps.Pipeline, error){
		"v1": parsePipelineV1,
	}
)

type parser struct{}

func newParser() *parser {
	return &parser{}
}

func (p *parser) ParsePipeline(dirPath string) (*pps.Pipeline, error) {
	return parsePipeline(dirPath)
}

type config struct {
	Version string
	Include []string
}

func parsePipeline(dirPath string) (*pps.Pipeline, error) {
	dirPath = filepath.Clean(dirPath)
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
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	config := &config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func parsePipelineV1(dirPath string, config *config) (*pps.Pipeline, error) {
	filePaths := make([]string, len(config.Include))
	for i, include := range config.Include {
		filePaths[i] = filepath.Join(dirPath, include)
	}
	pipeline := &pps.Pipeline{
		NameToNode:          make(map[string]*pps.Node),
		NameToDockerService: make(map[string]*pps.DockerService),
	}
	names := make(map[string]bool)
	for _, filePath := range filePaths {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		m := make(map[interface{}]interface{})
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		ppsMetaObj, ok := m["pps"]
		if !ok {
			return nil, fmt.Errorf("no pps section for %s", filePath)
		}
		ppsMeta := ppsMetaObj.(map[interface{}]interface{})
		if ppsMeta["kind"] == "" {
			return nil, fmt.Errorf("no kind specified for %s", filePath)
		}
		nameObj, ok := ppsMeta["name"]
		if !ok {
			return nil, fmt.Errorf("no name specified for %s", filePath)
		}
		name := strings.TrimSpace(nameObj.(string))
		if name == "" {
			return nil, fmt.Errorf("no name specified for %s", filePath)
		}
		if _, ok := names[name]; ok {
			return nil, fmt.Errorf("duplicate element: %s", name)
		}
		kindObj, ok := ppsMeta["kind"]
		if !ok {
			return nil, fmt.Errorf("no kind specified for %s", filePath)
		}
		kind := strings.TrimSpace(kindObj.(string))
		switch kind {
		case "node":
			node := &pps.Node{}
			if err := yaml.Unmarshal(data, node); err != nil {
				return nil, err
			}
			pipeline.NameToNode[name] = node
		case "docker_service":
			dockerService := &pps.DockerService{}
			if err := yaml.Unmarshal(data, dockerService); err != nil {
				return nil, err
			}
			if dockerService.Build != "" {
				dockerService.Build = filepath.Clean(filepath.Join(filepath.Dir(filePath), dockerService.Build))
			}
			pipeline.NameToDockerService[name] = dockerService
		default:
			return nil, fmt.Errorf("unknown kind %s for %s", kind, filePath)
		}
	}
	return pipeline, nil
}

func getElementForPipelineFile(dirPath string, filePath string) (*pps.Element, error) {
	filePath := filepath.Join(dirPath, filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	ppsMetaObj, ok := m["pps"]
	if !ok {
		return nil, fmt.Errorf("no pps section for %s", filePath)
	}
	ppsMeta := ppsMetaObj.(map[interface{}]interface{})
	if ppsMeta["kind"] == "" {
		return nil, fmt.Errorf("no kind specified for %s", filePath)
	}
	nameObj, ok := ppsMeta["name"]
	if !ok {
		return nil, fmt.Errorf("no name specified for %s", filePath)
	}
	name := strings.TrimSpace(nameObj.(string))
	if name == "" {
		return nil, fmt.Errorf("no name specified for %s", filePath)
	}
	kindObj, ok := ppsMeta["kind"]
	if !ok {
		return nil, fmt.Errorf("no kind specified for %s", filePath)
	}
	kind := strings.TrimSpace(kindObj.(string))
	switch kind {
	case "node":
		node := &pps.Node{}
		if err := yaml.Unmarshal(data, node); err != nil {
			return nil, err
		}
		element.Node = node
	case "docker_service":
		dockerService := &pps.DockerService{}
		if err := yaml.Unmarshal(data, dockerService); err != nil {
			return nil, err
		}
		if dockerService.Build != "" {
			dockerService.Build = filepath.Clean(filepath.Join(filepath.Dir(filePath), dockerService.Build))
		}
		element.DockerService = dockerService
	default:
		return nil, fmt.Errorf("unknown kind %s for %s", kind, filePath)
	}
	return element, nil
}

func checkFileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

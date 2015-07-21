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
		"v1": parsePipelineV1,
	}
)

type config struct {
	Version string
	Include []string
	Exclude []string
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
	dirPath = filepath.Clean(dirPath)
	var matches []string
	var err error
	if len(config.Include) == 0 {
		matches, err = filepath.Glob(fmt.Sprintf("%s/**", dirPath))
		if err != nil {
			return nil, err
		}
	} else {
		for _, include := range config.Include {
			// TODO(pedge): check if out of scope
			subMatches, err := filepath.Glob(fmt.Sprintf("%s/*/*/*", filepath.Clean(filepath.Join(dirPath, include))))
			if err != nil {
				return nil, err
			}
			matches = append(matches, subMatches...)
		}
	}
	relMatches := make([]string, len(matches))
	for i, match := range matches {
		relMatch, err := filepath.Rel(dirPath, match)
		if err != nil {
			return nil, err
		}
		relMatches[i] = relMatch
	}
	var filteredMatches []string
	for _, relMatch := range relMatches {
		isPipelineFile, err := isPipelineFileV1(relMatch, config.Exclude)
		if err != nil {
			return nil, err
		}
		if isPipelineFile {
			filteredMatches = append(filteredMatches, relMatch)
		}
	}
	pipeline := &pps.Pipeline{
		NameToElement: make(map[string]*pps.Element),
	}
	for _, filteredMatch := range filteredMatches {
		element, err := getElementForPipelineFileV1(dirPath, filteredMatch)
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

func isPipelineFileV1(filePath string, ignorePatterns []string) (bool, error) {
	if !strings.HasSuffix(filePath, ".yml") {
		return false, nil
	}
	for _, ignorePattern := range ignorePatterns {
		if strings.HasPrefix(filePath, ignorePattern) {
			return false, nil
		}
		matched, err := filepath.Match(ignorePattern, filePath)
		if err != nil {
			return false, err
		}
		if matched {
			return false, nil
		}
	}
	return true, nil
}

func getElementForPipelineFileV1(dirPath string, relFilePath string) (*pps.Element, error) {
	fmt.Println(relFilePath)
	return &pps.Element{
		Name: relFilePath,
	}, nil
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

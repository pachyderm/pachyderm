package parse //import "go.pachyderm.com/pachyderm/src/pps/parse"

import "go.pachyderm.com/pachyderm/src/pps"

type Parser interface {
	// ParsePipeline parses the pipeline
	// Id and PipelineSourceId will not be set!
	ParsePipeline(dirPath string) (*pps.Pipeline, error)
}

func NewParser() Parser {
	return newParser()
}

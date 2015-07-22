package parse

import "github.com/pachyderm/pachyderm/src/pps"

type Parser interface {
	ParsePipeline(dirPath string) (*pps.Pipeline, error)
}

func NewParser() Parser {
	return newParser()
}

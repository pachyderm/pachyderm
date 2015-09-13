package source

import "github.com/pachyderm/pachyderm/src/pps"

type Sourcer interface {
	// Id will not be set!
	GetDirPathAndPipeline(pipelineSource *pps.PipelineSource) (string, *pps.Pipeline, error)
}

func NewSourcer() Sourcer {
	return newSourcer()
}

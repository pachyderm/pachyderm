package source

import "github.com/pachyderm/pachyderm/src/pps"

type Sourcer interface {
	GetDirPathAndPipeline(pipelineSource *pps.PipelineSource) (string, *pps.Pipeline, error)
}

func NewSourcer() Sourcer {
	return newSourcer()
}

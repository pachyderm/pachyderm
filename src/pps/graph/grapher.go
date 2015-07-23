package graph

import "github.com/pachyderm/pachyderm/src/pps"

type grapher struct{}

func newGrapher() *grapher {
	return &grapher{}
}

func (g *grapher) GetPipelineInfo(pipeline *pps.Pipeline) (*PipelineInfo, error) {
	return getPipelineInfo(pipeline)
}

func getPipelineInfo(pipeline *pps.Pipeline) (*PipelineInfo, error) {
	return nil, nil
}

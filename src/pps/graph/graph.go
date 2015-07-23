package graph

import "github.com/pachyderm/pachyderm/src/pps"

type NodeInfo struct {
	Parents  []string
	Children []string
}

type PipelineInfo struct {
	NameToNodeInfo map[string]*NodeInfo
}

type Grapher interface {
	GetPipelineInfo(pipeline *pps.Pipeline) (*PipelineInfo, error)
}

// TODO(pedge): cycle detection
func NewGrapher() Grapher {
	return newGrapher()
}

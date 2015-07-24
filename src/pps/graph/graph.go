package graph

import "github.com/pachyderm/pachyderm/src/pps"

type NodeInfo struct {
	Parents  []string
	Children []string
}

type NodeErrorRecorder interface {
	Record(nodeName string, err error)
}

type Run interface {
	Do()
	Cancel()
}

type Grapher interface {
	Build(
		nodeErrorRecorder NodeErrorRecorder,
		nameToNodeInfo map[string]*NodeInfo,
		nameToNodeFunc map[string]func() error,
	) (Run, error)
}

// TODO(pedge): cycle detection
func NewGrapher() Grapher {
	return newGrapher()
}

func GetNameToNodeInfo(nodes map[string]*pps.Node) (map[string]*NodeInfo, error) {
	return getNameToNodeInfo(nodes)
}

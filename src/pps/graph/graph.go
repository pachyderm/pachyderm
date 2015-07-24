package graph

import "github.com/pachyderm/pachyderm/src/pps"

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
		nameToNode map[string]*pps.Node,
		nameToNodeFunc map[string]func() error,
	) (Run, error)
}

// TODO(pedge): cycle detection
func NewGrapher() Grapher {
	//return newGrapher()
	return nil
}

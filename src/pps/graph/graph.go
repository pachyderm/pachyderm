package graph

type NodeInfo struct {
	Parents []string
}

type NodeErrorRecorder interface {
	Record(nodeName string, err error)
}

type Run interface {
	Do() error
	Cancel()
}

type Grapher interface {
	Build(
		nodeErrorRecorder NodeErrorRecorder,
		nameToNodeInfo map[string]*NodeInfo,
		nameToNodeFunc map[string]func() error,
	) (Run, error)
}

func NewGrapher() Grapher {
	// TODO(pedge): cycle detection
	return newGrapher()
}

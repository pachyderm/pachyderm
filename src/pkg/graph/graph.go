package graph

type NodeInfo struct {
	Parents []string
}

type Run interface {
	Do() error
	Cancel()
}

type Grapher interface {
	Build(
		nameToNodeInfo map[string]*NodeInfo,
		nameToNodeFunc map[string]func() error,
	) (Run, error)
}

func NewGrapher() Grapher {
	// TODO(pedge): cycle detection
	return newGrapher()
}

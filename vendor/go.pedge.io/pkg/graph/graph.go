package pkggraph

// NodeInfo represents the information for a node.
type NodeInfo struct {
	Parents []string
}

// Run represents one run of a graph.
type Run interface {
	Do() error
	Cancel()
}

// Grapher provides functionality for creating graphs.
type Grapher interface {
	Build(
		nameToNodeInfo map[string]*NodeInfo,
		nameToNodeFunc map[string]func() error,
	) (Run, error)
}

// NewGrapher creates a new graph.
func NewGrapher() Grapher {
	return newGrapher()
}

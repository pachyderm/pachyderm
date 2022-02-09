package pretty

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	maxLabelLen = 10
)

type vertex struct {
	id     string
	label  string
	hidden bool
	edges  []*vertex
}

func Draw(commit pfs.CommitSet, src *os.FileInfo) {
	// create graph
	vertices := compute(baseGraph(commit))
	draw(vertices)
}

func baseGraph(commit pfs.CommitSet) []*vertex {
	return make([]*vertex, 0)
}

func compute(vertices []*vertex) []*vertex {
	// Assign Layers

	// Order Vertices

	// Assign Coordinates

	return vertices
}

func draw(vertices []*vertex) {

}

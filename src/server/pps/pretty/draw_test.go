package pretty

import (
	"fmt"
	"testing"
)

func sequentialConnectedDAG(layers [][]*vertex) []*vertex {
	vs := make([]*vertex, 0)
	var prevLayer []*vertex
	for _, l := range layers {
		for _, v := range l {
			vs = append(vs, v)
			if prevLayer != nil {
				for _, pv := range prevLayer {
					pv.addEdge(v)
				}
			}
		}
		prevLayer = l
	}
	return vs
}

func TestX(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("configs"), newVertex("datas")},
		{newVertex("clean")},
		{newVertex("model"), newVertex("analysis")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestV(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("configs"), newVertex("datas")},
		{newVertex("transform")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestDiamond(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("datas")},
		{newVertex("features"), newVertex("clean")},
		{newVertex("model")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestChain(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("stats")},
		{newVertex("clean")},
		{newVertex("enhance")},
		{newVertex("dump")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestCrossLayer(t *testing.T) {
	config := newVertex("configs")
	model := newVertex("model")
	layers := [][]*vertex{
		{newVertex("raw_data"), config},
		{newVertex("transform")},
		{newVertex("dashboard")},
		{model},
	}
	// add an extra connection from layer 3 -> 1
	config.addEdge(model)
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func drawMultiAlgos(vs []*vertex) {
	layerers := []layerer{
		layerLongestPath,
	}
	for _, lyr := range layerers {
		fmt.Print(draw(vs, lyr, simpleOrder))
	}
}

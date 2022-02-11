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
		{newVertex("a"), newVertex("b")},
		{newVertex("c")},
		{newVertex("d"), newVertex("e")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestV(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("hello"), newVertex("b")},
		{newVertex("c")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestDiamond(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("a")},
		{newVertex("b"), newVertex("c")},
		{newVertex("d")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestChain(t *testing.T) {
	layers := [][]*vertex{
		{newVertex("a")},
		{newVertex("b")},
		{newVertex("c")},
		{newVertex("d")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(vs)
}

func TestCrossLayer(t *testing.T) {
	b := newVertex("b")
	e := newVertex("e")
	layers := [][]*vertex{
		{newVertex("a"), b},
		{newVertex("c")},
		{newVertex("d")},
		{e},
	}
	// add an extra connection from layer 3 -> 1
	b.addEdge(e)
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

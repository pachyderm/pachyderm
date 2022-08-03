package pretty

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func sequentialConnectedDAG(layers [][]*vertex) []*vertex {
	vs := make([]*vertex, 0)
	var prevLayer []*vertex
	for _, l := range layers {
		for _, v := range l {
			vs = append(vs, v)
			for _, pv := range prevLayer {
				pv.addEdge(v)
			}
		}
		prevLayer = l
	}
	return vs
}

func TestX(t *testing.T) {
	expected := `
           +-----------+    +-----------+
           |  configs  |    |   datas   |
           +-----------+    +-----------+
                  \              /                  
                   \            /                   
                    ----     ---                    
                        \   /                       
                         \ /                        
                    +-----------+
                    |   clean   |
                    +-----------+
                         / \                        
                        /   \                       
                    ----     ---                    
                   /            \                   
                  /              \                  
           +-----------+    +-----------+
           |   model   |    | analysis  |
           +-----------+    +-----------+
`
	layers := [][]*vertex{
		{newVertex("configs"), newVertex("datas")},
		{newVertex("clean")},
		{newVertex("model"), newVertex("analysis")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(t, vs, expected)
}

func TestV(t *testing.T) {
	expected := `
           +-----------+    +-----------+
           |  configs  |    |   datas   |
           +-----------+    +-----------+
                  \              /                  
                   \            /                   
                    ----     ---                    
                        \   /                       
                         \ /                        
                    +-----------+
                    | transform |
                    +-----------+
`
	layers := [][]*vertex{
		{newVertex("configs"), newVertex("datas")},
		{newVertex("transform")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(t, vs, expected)
}

func TestDiamond(t *testing.T) {
	expected := `
                    +-----------+
                    |   datas   |
                    +-----------+
                         / \                        
                        /   \                       
                    ----     ---                    
                   /            \                   
                  /              \                  
           +-----------+    +-----------+
           | features  |    |   clean   |
           +-----------+    +-----------+
                  \              /                  
                   \            /                   
                    ----     ---                    
                        \   /                       
                         \ /                        
                    +-----------+
                    |   model   |
                    +-----------+
`
	layers := [][]*vertex{
		{newVertex("datas")},
		{newVertex("features"), newVertex("clean")},
		{newVertex("model")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(t, vs, expected)
}

func TestChain(t *testing.T) {
	expected := `
       +-----------+
       |   stats   |
       +-----------+
             |            
             |            
             |            
             |            
             |            
       +-----------+
       |   clean   |
       +-----------+
             |            
             |            
             |            
             |            
             |            
       +-----------+
       |  enhance  |
       +-----------+
             |            
             |            
             |            
             |            
             |            
       +-----------+
       |   dump    |
       +-----------+
`
	layers := [][]*vertex{
		{newVertex("stats")},
		{newVertex("clean")},
		{newVertex("enhance")},
		{newVertex("dump")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(t, vs, expected)
}

func TestCrossLayer(t *testing.T) {
	expected := `
           +-----------+    +-----------+
           | raw_data  |    |  configs  |
           +-----------+    +-----------+
                 |               /|                 
                 |              / |                 
                 |  ------------  |                 
                 | /              |                 
                 |/               |                 
           +-----------+          |      
           | transform |          |      
           +-----------+          |      
                 |                |                 
                 |                |                 
                 |                |                 
                 |                |                 
                 |                |                 
           +-----------+          |      
           | dashboard |          |      
           +-----------+          |      
                  \              /                  
                   \            /                   
                    ----     ---                    
                        \   /                       
                         \ /                        
                    +-----------+
                    |   model   |
                    +-----------+                       
`
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
	drawMultiAlgos(t, vs, expected)
}

func TestBiPartite(t *testing.T) {
	expected := `
             +-----------+      +-----------+      +-----------+
             |  configs  |      |   data    |      |  biases   |
             +-----------+      +-----------+      +-----------+
                    \                / \                /                     
                     \              /   \              /                      
                      --------------------------------+                       
                        \   /                     \   /                       
                         \ /                       \ /                        
                    +-----------+             +-----------+
                    |   model   |             | analysis  |
                    +-----------+             +-----------+
`
	layers := [][]*vertex{
		{newVertex("configs"), newVertex("data"), newVertex("biases")},
		{newVertex("model"), newVertex("analysis")},
	}
	vs := sequentialConnectedDAG(layers)
	drawMultiAlgos(t, vs, expected)
}

func drawMultiAlgos(t testing.TB, vs []*vertex, expected string) {
	layerers := []layerer{
		layerLongestPath,
	}
	for _, lyr := range layerers {
		picture := draw(vs, lyr, simpleOrder)
		require.Equal(t, strings.Trim(expected, "\n "), strings.Trim(picture, "\n "))
		fmt.Print(picture)
	}
}

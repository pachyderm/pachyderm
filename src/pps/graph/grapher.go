package graph

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pps"
)

type grapher struct{}

func newGrapher() *grapher {
	return &grapher{}
}

func (g *grapher) GetPipelineInfo(pipeline *pps.Pipeline) (*PipelineInfo, error) {
	return getPipelineInfo(pipeline)
}

func getPipelineInfo(pipeline *pps.Pipeline) (*PipelineInfo, error) {
	nodes := getNameToNode(pipeline)
	nodeToInputs := getNodeNameToInputStrings(nodes)
	nodeToOutputs := getNodeNameToOutputStrings(nodes)
	inputToNodes := getInputStringToNodeNames(nodes)
	outputToNodes := getOutputStringToNodeNames(nodes)
	nodeInfos := make(map[string](*NodeInfo))
	for name := range nodes {
		nodeInfo := &NodeInfo{
			Parents:  make([]string, 0),
			Children: make([]string, 0),
		}
		parents := make(map[string]bool)
		for input := range nodeToInputs[name] {
			for parent := range outputToNodes[input] {
				if parent != name {
					parents[parent] = true
				}
			}
		}
		for parent := range parents {
			nodeInfo.Parents = append(nodeInfo.Parents, parent)
		}
		children := make(map[string]bool)
		for output := range nodeToOutputs[name] {
			for child := range inputToNodes[output] {
				if child != name {
					children[child] = true
				}
			}
		}
		for child := range children {
			nodeInfo.Children = append(nodeInfo.Children, child)
		}
		nodeInfos[name] = nodeInfo
	}
	pipelineInfo := &PipelineInfo{nodeInfos}
	log.Printf("got pipeline info %v\n", pipelineInfo)
	return pipelineInfo, nil
}

func getNodeNameToInputStrings(nodes map[string]*pps.Node) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for name, node := range nodes {
		n := make(map[string]bool)
		if node.Input != nil {
			for hostDir := range node.Input.Host {
				// just need a differentiating string between types
				n[fmt.Sprintf("host://%s", hostDir)] = true
			}
			for pfsRepo := range node.Input.Pfs {
				n[fmt.Sprintf("pfs://%s", pfsRepo)] = true
			}
		}
		m[name] = n
	}
	return m
}

func getNodeNameToOutputStrings(nodes map[string]*pps.Node) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for name, node := range nodes {
		n := make(map[string]bool)
		if node.Output != nil {
			for hostDir := range node.Output.Host {
				// just need a differentiating string between types
				n[fmt.Sprintf("host://%s", hostDir)] = true
			}
			for pfsRepo := range node.Output.Pfs {
				n[fmt.Sprintf("pfs://%s", pfsRepo)] = true
			}
		}
		m[name] = n
	}
	return m
}

func getInputStringToNodeNames(nodes map[string]*pps.Node) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for name, node := range nodes {
		if node.Input != nil {
			for hostDir := range node.Input.Host {
				// just need a differentiating string between types
				s := fmt.Sprintf("host://%s", hostDir)
				if _, ok := m[s]; !ok {
					m[s] = make(map[string]bool)
				}
				m[s][name] = true
			}
			for pfsRepo := range node.Input.Pfs {
				s := fmt.Sprintf("pfs://%s", pfsRepo)
				if _, ok := m[s]; !ok {
					m[s] = make(map[string]bool)
				}
				m[s][name] = true
			}
		}
	}
	return m
}

func getOutputStringToNodeNames(nodes map[string]*pps.Node) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for name, node := range nodes {
		if node.Output != nil {
			for hostDir := range node.Output.Host {
				// just need a differentiating string between types
				s := fmt.Sprintf("host://%s", hostDir)
				if _, ok := m[s]; !ok {
					m[s] = make(map[string]bool)
				}
				m[s][name] = true
			}
			for pfsRepo := range node.Output.Pfs {
				s := fmt.Sprintf("pfs://%s", pfsRepo)
				if _, ok := m[s]; !ok {
					m[s] = make(map[string]bool)
				}
				m[s][name] = true
			}
		}
	}
	return m
}

func getNameToNode(pipeline *pps.Pipeline) map[string]*pps.Node {
	m := make(map[string]*pps.Node)
	for name, element := range pipeline.NameToElement {
		if element.Node != nil {
			m[name] = element.Node
		}
	}
	return m
}

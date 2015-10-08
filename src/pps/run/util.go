package run

import (
	"fmt"

	"go.pedge.io/pkg/graph"

	"go.pachyderm.com/pachyderm/src/pps"
)

func getInputBinds(input *pps.Input) []string {
	if input != nil && input.Host != nil {
		return getBinds(input.Host, "ro")
	}
	return []string{}
}

func getOutputBinds(output *pps.Output) []string {
	if output != nil && output.Host != nil {
		return getBinds(output.Host, "rw")
	}
	return []string{}
}

func getBinds(host map[string]string, postfix string) []string {
	var binds []string
	for key, value := range host {
		binds = append(binds, fmt.Sprintf("%s:%s:%s", key, value, postfix))
	}
	return binds
}

func getNameToNodeInfo(nodes map[string]*pps.Node) (map[string]*pkggraph.NodeInfo, error) {
	nodeToInputs := getNodeNameToInputStrings(nodes)
	outputToNodes := getOutputStringToNodeNames(nodes)
	nodeInfos := make(map[string](*pkggraph.NodeInfo))
	for name := range nodes {
		nodeInfo := &pkggraph.NodeInfo{
			Parents: make([]string, 0),
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
		nodeInfos[name] = nodeInfo
	}
	return nodeInfos, nil
}

func getNodeNameToInputStrings(nodes map[string]*pps.Node) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for name, node := range nodes {
		n := make(map[string]bool)
		if node.Input != nil {
			for _, iNode := range node.Input.Node {
				n[fmt.Sprintf("node://%s", iNode)] = true
			}
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
		s := fmt.Sprintf("node://%s", name)
		for subName := range nodes {
			if name != subName {
				if _, ok := m[s]; !ok {
					m[s] = make(map[string]bool)
				}
				m[s][subName] = true
			}
		}
	}
	return m
}

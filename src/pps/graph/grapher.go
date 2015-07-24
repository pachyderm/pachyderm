package graph

import (
	"fmt"
	"sync"

	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pps"
)

type nodeInfo struct {
	parents  []string
	children []string
}

type run struct {
	nodeRunners    map[string]*nodeRunner
	nameToNodeFunc map[string]func() error
	cancel         chan<- bool
}

func (r *run) Do() {
	runNodes(r.nodeRunners, r.nameToNodeFunc)
}

func (r *run) Cancel() {
	r.cancel <- true
	close(r.cancel)
}

type grapher struct{}

func newGrapher() *grapher {
	return &grapher{}
}

//func (g *grapher) GetNodeRunners(pipeline *pps.Pipeline) ([]NodeRunner, error) {
//return getNodeRunners(pipeline)
//}

//func getNodeRunners(pipeline *pps.Pipeline) ([]NodeRunner, error) {
//return nil, nil
//}

func (g *grapher) Build(
	nodeErrorRecorder NodeErrorRecorder,
	nameToNode map[string]*pps.Node,
	nameToNodeFunc map[string]func() error,
) (Run, error) {
	return build(
		nodeErrorRecorder,
		nameToNode,
		nameToNodeFunc,
	)
}

func build(
	nodeErrorRecorder NodeErrorRecorder,
	nameToNode map[string]*pps.Node,
	nameToNodeFunc map[string]func() error,
) (*run, error) {
	nodeInfos, err := getNameToNodeInfo(nameToNode)
	if err != nil {
		return nil, err
	}
	cancel := make(chan bool)
	nodeRunners, err := getNameToNodeRunner(nodeInfos, nodeErrorRecorder, cancel)
	if err != nil {
		return nil, err
	}
	if err := checkNodeRunners(nodeRunners, nameToNodeFunc); err != nil {
		return nil, err
	}
	return &run{
		nodeRunners,
		nameToNodeFunc,
		cancel,
	}, nil
}

func checkNodeRunners(
	nodeRunners map[string]*nodeRunner,
	nameToNodeFunc map[string]func() error,
) error {
	for name := range nameToNodeFunc {
		if _, ok := nodeRunners[name]; !ok {
			return fmt.Errorf("no node runner for %s", name)
		}
	}
	for name := range nodeRunners {
		if _, ok := nameToNodeFunc[name]; !ok {
			return fmt.Errorf("no node runner for %s", name)
		}
	}
	return nil
}

func runNodes(
	nodeRunners map[string]*nodeRunner,
	nameToNodeFunc map[string]func() error,
) {
	var wg sync.WaitGroup
	for name, nodeRunner := range nodeRunners {
		name := name
		nodeRunner := nodeRunner
		nodeFunc := nameToNodeFunc[name]
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeRunner.run(nodeFunc)
		}()
	}
	wg.Wait()
}

func getNameToNodeRunner(
	nodeInfos map[string]*nodeInfo,
	nodeErrorRecorder NodeErrorRecorder,
	cancel <-chan bool,
) (map[string]*nodeRunner, error) {
	nodeRunners := make(map[string]*nodeRunner, len(nodeInfos))
	for name := range nodeInfos {
		nodeRunners[name] = newNodeRunner(
			name,
			nodeErrorRecorder,
			cancel,
		)
	}
	for name, nodeInfo := range nodeInfos {
		for _, child := range nodeInfo.children {
			errC := make(chan error, 1)
			nodeRunner := nodeRunners[name]
			childNodeRunner := nodeRunners[child]
			nodeRunner.addChild(errC)
			childNodeRunner.addParent(errC)
		}
	}
	return nodeRunners, nil
}

func getNameToNodeInfo(nodes map[string]*pps.Node) (map[string]*nodeInfo, error) {
	nodeToInputs := getNodeNameToInputStrings(nodes)
	nodeToOutputs := getNodeNameToOutputStrings(nodes)
	inputToNodes := getInputStringToNodeNames(nodes)
	outputToNodes := getOutputStringToNodeNames(nodes)
	nodeInfos := make(map[string](*nodeInfo))
	for name := range nodes {
		nodeInfo := &nodeInfo{
			parents:  make([]string, 0),
			children: make([]string, 0),
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
			nodeInfo.parents = append(nodeInfo.parents, parent)
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
			nodeInfo.children = append(nodeInfo.children, child)
		}
		nodeInfos[name] = nodeInfo
	}
	log.Printf("got node infos %v\n", nodeInfos)
	return nodeInfos, nil
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

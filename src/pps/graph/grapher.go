package graph

import (
	"fmt"
	"sync"
)

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

func (g *grapher) Build(
	nodeErrorRecorder NodeErrorRecorder,
	nameToNodeInfo map[string]*NodeInfo,
	nameToNodeFunc map[string]func() error,
) (Run, error) {
	return build(
		nodeErrorRecorder,
		nameToNodeInfo,
		nameToNodeFunc,
	)
}

func build(
	nodeErrorRecorder NodeErrorRecorder,
	nameToNodeInfo map[string]*NodeInfo,
	nameToNodeFunc map[string]func() error,
) (*run, error) {
	cancel := make(chan bool)
	nodeRunners, err := getNameToNodeRunner(nameToNodeInfo, nodeErrorRecorder, cancel)
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
	nodeInfos map[string]*NodeInfo,
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
		for _, child := range nodeInfo.Children {
			errC := make(chan error, 1)
			nodeRunner := nodeRunners[name]
			childNodeRunner := nodeRunners[child]
			nodeRunner.addChild(errC)
			childNodeRunner.addParent(errC)
		}
	}
	return nodeRunners, nil
}

package graph

import (
	"fmt"
	"sync"
)

type run struct {
	nodeRunners map[string]*nodeRunner
	cancel      chan<- bool
}

func (r *run) Do() {
	var wg sync.WaitGroup
	for _, nodeRunner := range r.nodeRunners {
		nodeRunner := nodeRunner
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeRunner.run()
		}()
	}
	wg.Wait()
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
	nodeRunners, err := getNameToNodeRunner(nameToNodeInfo, nameToNodeFunc, nodeErrorRecorder, cancel)
	if err != nil {
		return nil, err
	}
	return &run{
		nodeRunners,
		cancel,
	}, nil
}

func getNameToNodeRunner(
	nodeInfos map[string]*NodeInfo,
	nameToNodeFunc map[string]func() error,
	nodeErrorRecorder NodeErrorRecorder,
	cancel <-chan bool,
) (map[string]*nodeRunner, error) {
	if err := checkNodeInfos(nodeInfos, nameToNodeFunc); err != nil {
		return nil, err
	}
	nodeRunners := make(map[string]*nodeRunner, len(nodeInfos))
	for name := range nodeInfos {
		nodeRunners[name] = newNodeRunner(
			name,
			nameToNodeFunc[name],
			nodeErrorRecorder,
			cancel,
		)
	}
	for name, nodeInfo := range nodeInfos {
		for _, parent := range nodeInfo.Parents {
			errC := make(chan error, 1)
			nodeRunner := nodeRunners[name]
			parentNodeRunner := nodeRunners[parent]
			if err := nodeRunner.addParent(parent, errC); err != nil {
				return nil, err
			}
			if err := parentNodeRunner.addChild(name, errC); err != nil {
				return nil, err
			}
		}
	}
	return nodeRunners, nil
}

func checkNodeInfos(
	nodeInfos map[string]*NodeInfo,
	nameToNodeFunc map[string]func() error,
) error {
	for name := range nameToNodeFunc {
		if _, ok := nodeInfos[name]; !ok {
			return fmt.Errorf("no node info for %s", name)
		}
	}
	for name := range nodeInfos {
		if _, ok := nameToNodeFunc[name]; !ok {
			return fmt.Errorf("no node func for %s", name)
		}
	}
	return nil
}

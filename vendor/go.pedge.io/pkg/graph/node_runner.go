package pkggraph

import (
	"fmt"

	"go.pedge.io/protolog"
)

type nodeRunner struct {
	nodeName      string
	f             func() error
	parentChans   map[string]<-chan error
	childrenChans map[string]chan<- error
	cancel        <-chan bool
}

func newNodeRunner(
	nodeName string,
	f func() error,
	cancel <-chan bool,
) *nodeRunner {
	return &nodeRunner{
		nodeName,
		f,
		make(map[string]<-chan error),
		make(map[string]chan<- error),
		cancel,
	}
}

func (n *nodeRunner) name() string {
	return n.nodeName
}

func (n *nodeRunner) addParent(parentName string, parentChan <-chan error) error {
	chanName := fmt.Sprintf("%s_parent_to_%s", parentName, n.nodeName)
	if _, ok := n.parentChans[chanName]; ok {
		return fmt.Errorf("duplicate channel %s", chanName)
	}
	n.parentChans[chanName] = parentChan
	return nil
}

func (n *nodeRunner) addChild(childName string, childChan chan<- error) error {
	chanName := fmt.Sprintf("%s_parent_to_%s", n.nodeName, childName)
	if _, ok := n.childrenChans[chanName]; ok {
		return fmt.Errorf("duplicate channel %s", chanName)
	}
	n.childrenChans[chanName] = childChan
	return nil
}

func (n *nodeRunner) run() error {
	var err error
	for name, parentChan := range n.parentChans {
		protolog.Debug(&NodeWaiting{Node: n.nodeName, ParentNode: name})
		select {
		case parentErr := <-parentChan:
			if parentErr != nil {
				err = parentErr
			}
			continue
		case <-n.cancel:
			return err
		}
	}
	protolog.Debug(&NodeFinishedWaiting{Node: n.nodeName, ParentError: errorString(err)})
	if err == nil {
		protolog.Info(&NodeStarting{Node: n.nodeName})
		err = n.f()
		protolog.Info(&NodeFinished{Node: n.nodeName, Error: errorString(err)})
	}
	for name, childChan := range n.childrenChans {
		protolog.Debug(&NodeSending{Node: n.nodeName, ChildNode: name, Error: errorString(err)})
		childChan <- err
		close(childChan)
	}
	return err
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

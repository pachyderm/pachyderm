package graph

import (
	"fmt"
	"log"
)

type nodeRunner struct {
	nodeName          string
	f                 func() error
	parentChans       map[string]<-chan error
	childrenChans     map[string]chan<- error
	nodeErrorRecorder NodeErrorRecorder
	cancel            <-chan bool
}

func newNodeRunner(
	nodeName string,
	f func() error,
	nodeErrorRecorder NodeErrorRecorder,
	cancel <-chan bool,
) *nodeRunner {
	return &nodeRunner{
		nodeName,
		f,
		make(map[string]<-chan error),
		make(map[string]chan<- error),
		nodeErrorRecorder,
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

func (n *nodeRunner) run() {
	var err error
	for name, parentChan := range n.parentChans {
		log.Printf("%s is waiting on channel %s\n", n.nodeName, name)
		select {
		case parentErr := <-parentChan:
			if parentErr != nil {
				err = parentErr
			}
			continue
		case <-n.cancel:
			return
		}
	}
	log.Printf("%s is done waiting, had parent error %v\n", n.nodeName, err)
	if err == nil {
		err = n.f()
		log.Printf("%s is finished running func\n", n.nodeName)
		if err != nil {
			n.nodeErrorRecorder.Record(n.nodeName, err)
		}
	}
	for name, childChan := range n.childrenChans {
		log.Printf("%s is sending to channel %s\n", n.nodeName, name)
		childChan <- err
		close(childChan)
	}
}

package graph

type nodeRunner struct {
	nodeName          string
	parentChans       []<-chan error
	childrenChans     []chan<- error
	nodeErrorRecorder NodeErrorRecorder
	cancel            <-chan bool
}

func newNodeRunner(
	nodeName string,
	nodeErrorRecorder NodeErrorRecorder,
	cancel <-chan bool,
) *nodeRunner {
	return &nodeRunner{
		nodeName,
		make([]<-chan error, 0),
		make([]chan<- error, 0),
		nodeErrorRecorder,
		cancel,
	}
}

func (n *nodeRunner) name() string {
	return n.nodeName
}

func (n *nodeRunner) addParent(parentChan <-chan error) {
	n.parentChans = append(n.parentChans, parentChan)
}

func (n *nodeRunner) addChild(childChan chan<- error) {
	n.childrenChans = append(n.childrenChans, childChan)
}

func (n *nodeRunner) run(f func() error) {
	for _, parentChan := range n.parentChans {
		select {
		case <-parentChan:
			continue
		case <-n.cancel:
			return
		}
	}
	err := f()
	if err != nil {
		n.nodeErrorRecorder.Record(n.nodeName, err)
	}
	for _, childChan := range n.childrenChans {
		childChan <- err
	}
}

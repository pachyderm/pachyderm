package server

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/protolog"
)

type pipelineController struct {
	pfsAPIClient     pfs.APIClient
	persistAPIClient persist.APIClient

	pipeline        *persist.Pipeline
	cancelC         chan bool
	finishedCancelC chan bool
}

func newPipelineController(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	pipeline *persist.Pipeline,
) *pipelineController {
	return &pipelineController{
		pfsAPIClient,
		persistAPIClient,
		pipeline,
		make(chan bool),
		make(chan bool),
	}
}

func (p *pipelineController) Start() error {
	go p.run()
	return nil
}

func (p *pipelineController) Cancel() {
	p.cancelC <- true
	close(p.cancelC)
	<-p.finishedCancelC
}

func (p *pipelineController) run() {
	for {
		select {
		case <-p.cancelC:
			p.finishedCancelC <- true
			close(p.finishedCancelC)
			return
		default:
			// TODO(pedge): what to do with the error?
			if err := p.poll(); err != nil {
				protolog.Errorln(err.Error())
			}
		}
	}
}

// TODO(pedge): this is where the list commits logic goes
func (p *pipelineController) poll() error {
	return nil
}

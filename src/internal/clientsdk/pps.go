package clientsdk

import (
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func ForEachPipelineInfo(client pps.API_ListPipelineClient, cb func(*pps.PipelineInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(x); err != nil {
			if err == pacherr.ErrBreak {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListPipelineInfo(client pps.API_ListPipelineClient) ([]*pps.PipelineInfo, error) {
	var pipelineInfos []*pps.PipelineInfo
	if err := ForEachPipelineInfo(client, func(pi *pps.PipelineInfo) error {
		pipelineInfos = append(pipelineInfos, pi)
		return nil
	}); err != nil {
		return nil, err
	}
	return pipelineInfos, nil
}

func ForEachJobSet(client pps.API_ListJobSetClient, cb func(*pps.JobSetInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(x); err != nil {
			if err == pacherr.ErrBreak {
				err = nil
			}
			return err
		}
	}
	return nil
}

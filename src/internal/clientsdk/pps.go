package clientsdk

import (
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
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
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

// ForEachDatumInfo calls cb on each *pps.DatumInfo returned by client.  The
// loop can be broken early by returning pacherr.ErrBreak.  Otherwise, the first
// non-io.EOF error encountered while receiving or processing datum info
// messages will be returned.
func ForEachDatumInfo(client pps.API_ListDatumClient, cb func(*pps.DatumInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListDatum(client pps.API_ListDatumClient) ([]*pps.DatumInfo, error) {
	var results []*pps.DatumInfo
	if err := ForEachDatumInfo(client, func(x *pps.DatumInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachJobSetInfo(client pps.API_ListJobSetClient, cb func(*pps.JobSetInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ForEachJobInfo(client pps.API_ListJobClient, cb func(*pps.JobInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListJobSet(client pps.API_ListJobSetClient) ([]*pps.JobSetInfo, error) {
	var results []*pps.JobSetInfo
	if err := ForEachJobSetInfo(client, func(x *pps.JobSetInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ListJob(client pps.API_ListJobClient) ([]*pps.JobInfo, error) {
	var results []*pps.JobInfo
	if err := ForEachJobInfo(client, func(x *pps.JobInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachLokiLog(client pps.API_GetKubeEventsClient, cb func(*pps.LokiLogMessage) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListLokiLogs(client pps.API_GetKubeEventsClient) ([]*pps.LokiLogMessage, error) {
	var results []*pps.LokiLogMessage
	if err := ForEachLokiLog(client, func(x *pps.LokiLogMessage) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

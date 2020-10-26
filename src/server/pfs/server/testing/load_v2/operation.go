package load

import (
	"context"
	"io"
	"math/rand"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
	"modernc.org/mathutil"
)

type OperationsSpec struct {
	Count              int                  `yaml:"count,omitempty"`
	FuzzOperationSpecs []*FuzzOperationSpec `yaml:"fuzzOperations,omitempty"`
}

func Operations(c *client.APIClient, repo, commit string, spec *OperationsSpec, validator ...*Validator) error {
	for i := 0; i < spec.Count; i++ {
		return FuzzOperation(c, repo, commit, spec.FuzzOperationSpecs, validator...)
	}
	return nil
}

// TODO: Add different types of operations.
type OperationSpec struct {
	PutTarSpec *PutTarSpec `yaml:"putTar,omitempty"`
}

func Operation(c *client.APIClient, repo, commit string, spec *OperationSpec, validator ...*Validator) error {
	return PutTar(c, repo, commit, spec.PutTarSpec, validator...)
}

type PutTarSpec struct {
	FilesSpec      *FilesSpec      `yaml:"files,omitempty"`
	ThroughputSpec *ThroughputSpec `yaml:"throughput,omitempty"`
	CancelSpec     *CancelSpec     `yaml:"cancel,omitempty"`
}

func PutTar(c *client.APIClient, repo, commit string, spec *PutTarSpec, validator ...*Validator) error {
	files, err := Files(spec.FilesSpec)
	if err != nil {
		return err
	}
	r, err := tarutil.Reader(files)
	if err != nil {
		return err
	}
	if spec.ThroughputSpec != nil && shouldExecute(spec.ThroughputSpec.Prob) {
		r = newThroughputLimitReader(r, spec.ThroughputSpec.Limit)
	}
	if spec.CancelSpec != nil && shouldExecute(spec.CancelSpec.Prob) {
		c = newCancelClient(c, spec.CancelSpec)
	}
	if err := c.PutTarV2(repo, commit, r, false); err != nil {
		if c.Ctx().Err() == context.Canceled {
			return nil
		}
		return err
	}
	if len(validator) > 0 {
		return validator[0].AddFiles(files)
	}
	return nil
}

type GetTarSpec struct {
	ThroughputSpec *ThroughputSpec `yaml:"throughput,omitempty"`
	CancelSpec     *CancelSpec     `yaml:"cancel,omitempty"`
}

func GetTar(c *client.APIClient, repo, commit string, spec *GetTarSpec) (io.Reader, error) {
	if spec.CancelSpec != nil && shouldExecute(spec.CancelSpec.Prob) {
		c = newCancelClient(c, spec.CancelSpec)
	}
	r, err := c.GetTarV2(repo, commit, "**")
	if err != nil {
		return nil, err
	}
	if spec.ThroughputSpec != nil && shouldExecute(spec.ThroughputSpec.Prob) {
		r = newThroughputLimitReader(r, spec.ThroughputSpec.Limit)
	}
	return r, nil
}

type ThroughputSpec struct {
	Limit int     `yaml:"limit,omitempty"`
	Prob  float64 `yaml:"prob,omitempty"`
}

type throughputLimitReader struct {
	r                               io.Reader
	bytesSinceSleep, bytesPerSecond int
}

func newThroughputLimitReader(r io.Reader, bytesPerSecond int) *throughputLimitReader {
	return &throughputLimitReader{
		r:              r,
		bytesPerSecond: bytesPerSecond,
	}
}

func (tlr *throughputLimitReader) Read(data []byte) (int, error) {
	var bytesRead int
	for len(data) > 0 {
		size := mathutil.Min(len(data), tlr.bytesPerSecond-tlr.bytesSinceSleep)
		n, err := tlr.r.Read(data[:size])
		data = data[n:]
		bytesRead += n
		if err != nil {
			return bytesRead, err
		}
		if tlr.bytesSinceSleep == tlr.bytesPerSecond {
			time.Sleep(time.Second)
			tlr.bytesSinceSleep = 0
		}
	}
	return bytesRead, nil
}

type CancelSpec struct {
	MaxTime time.Duration `yaml:"maxTime,omitempty"`
	Prob    float64       `yaml:"prob,omitempty"`
}

func newCancelClient(c *client.APIClient, spec *CancelSpec) *client.APIClient {
	cancelCtx, cancel := context.WithCancel(c.Ctx())
	c = c.WithCtx(cancelCtx)
	go func() {
		<-time.After(time.Duration(int64(float64(int64(spec.MaxTime)) * rand.Float64())))
		cancel()
	}()
	return c
}

package pps

import (
	"hash/adler32"

	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

// A Hasher represents a job/pipeline hasher.
type Hasher struct {
	JobModulus      uint64
	PipelineModulus uint64
}

// NewHasher creates a hasher.
func NewHasher(jobModulus uint64, pipelineModulus uint64) *Hasher {
	return &Hasher{
		JobModulus:      jobModulus,
		PipelineModulus: pipelineModulus,
	}
}

// HashJob computes and returns a hash of a job.
func (s *Hasher) HashJob(job *ppsclient.Job) uint64 {
	return uint64(adler32.Checksum([]byte(job.ID))) % s.PipelineModulus
}

// HashPipeline computes and returns a hash of a pipeline.
func (s *Hasher) HashPipeline(pipeline *ppsclient.Pipeline) uint64 {
	return uint64(adler32.Checksum([]byte(pipeline.Name))) % s.JobModulus
}

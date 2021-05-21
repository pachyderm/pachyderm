package pps

import (
	"hash/adler32"
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

// HashJob computes and returns the hash of a pipeline job.
func (s *Hasher) HashJob(pipelineJobID string) uint64 {
	return uint64(adler32.Checksum([]byte(pipelineJobID))) % s.JobModulus
}

// HashPipeline computes and returns the hash of a pipeline.
func (s *Hasher) HashPipeline(pipelineName string) uint64 {
	return uint64(adler32.Checksum([]byte(pipelineName))) % s.PipelineModulus
}

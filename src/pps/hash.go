package pps

import (
	"hash/adler32"
)

type Hasher struct {
	JobModulus      uint64
	PipelineModulus uint64
}

func NewHasher(jobModulus uint64, pipelineModulus uint64) *Hasher {
	return &Hasher{
		JobModulus:      jobModulus,
		PipelineModulus: pipelineModulus,
	}
}

func (s *Hasher) HashJob(job *Job) uint64 {
	return uint64(adler32.Checksum([]byte(job.Id))) % s.PipelineModulus
}

func (s *Hasher) HashPipeline(pipeline *Pipeline) uint64 {
	return uint64(adler32.Checksum([]byte(pipeline.Name))) % s.JobModulus
}

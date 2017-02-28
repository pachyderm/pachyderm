package pps

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/adler32"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
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

// HashJob computes and returns the hash of a job.
func (s *Hasher) HashJob(jobID string) uint64 {
	return uint64(adler32.Checksum([]byte(jobID))) % s.PipelineModulus
}

// HashPipeline computes and returns the hash of a pipeline.
func (s *Hasher) HashPipeline(pipelineName string) uint64 {
	return uint64(adler32.Checksum([]byte(pipelineName))) % s.JobModulus
}

// HashDatum computes and returns the hash of a datum + pipeline.
func HashDatum(data []*pfs.FileInfo, pipelineInfo *pps.PipelineInfo) (string, error) {
	hash := sha256.New()
	for i, fileInfo := range data {
		if _, err := hash.Write([]byte(pipelineInfo.Inputs[i].Name)); err != nil {
			return "", err
		}
		if _, err := hash.Write(fileInfo.Hash); err != nil {
			return "", err
		}
	}
	bytes, err := proto.Marshal(pipelineInfo.Transform)
	if err != nil {
		return "", err
	}
	if _, err := hash.Write(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

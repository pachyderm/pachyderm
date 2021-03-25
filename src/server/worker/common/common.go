package common

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// IsDone returns true if the given context has been canceled, or false otherwise
func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// DatumID computes the ID of a datum.
func DatumID(inputs []*Input) string {
	hash := pfs.NewHash()
	for _, input := range inputs {
		hash.Write([]byte(input.Name))
		binary.Write(hash, binary.BigEndian, int64(len(input.Name)))
		hash.Write([]byte(input.FileInfo.File.Path))
		binary.Write(hash, binary.BigEndian, int64(len(input.FileInfo.File.Path)))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// HashDatum computes the hash of a datum.
func HashDatum(pipelineName string, pipelineSalt string, inputs []*Input) string {
	hash := sha256.New()
	for _, input := range inputs {
		hash.Write([]byte(input.Name))
		hash.Write([]byte(input.FileInfo.File.Path))
		hash.Write([]byte(input.FileInfo.Hash))
	}
	hash.Write([]byte(pipelineName))
	hash.Write([]byte(pipelineSalt))
	return client.DatumTagPrefix(pipelineSalt) + hex.EncodeToString(hash.Sum(nil))
}

// MatchDatum checks if a datum matches a filter.  To match each string in
// filter must correspond match at least 1 datum's Path or Hash. Order of
// filter and inputs is irrelevant.
func MatchDatum(filter []string, inputs []*pps.InputFile) bool {
	// All paths in request.DataFilters must appear somewhere in the log
	// line's inputs, or it's filtered
	matchesData := true
dataFilters:
	for _, dataFilter := range filter {
		for _, input := range inputs {
			if dataFilter == input.Path ||
				dataFilter == base64.StdEncoding.EncodeToString(input.Hash) ||
				dataFilter == hex.EncodeToString(input.Hash) {
				continue dataFilters // Found, move to next filter
			}
		}
		matchesData = false
		break
	}
	return matchesData
}

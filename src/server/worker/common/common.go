package common

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	fmt "fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
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

// DatumID computes the id for a datum, this value is used in ListDatum and
// InspectDatum.
func DatumID(inputs []*Input) string {
	hash := sha256.New()
	for _, input := range inputs {
		for _, fileInfo := range input.FileInfo {
			hash.Write([]byte(fileInfo.File.Path))
			hash.Write(fileInfo.Hash)
		}
	}
	// InputFileID is a single string id for the data from this input, it's used in logs and in
	// the statsTree
	return hex.EncodeToString(hash.Sum(nil))
}

// HashDatum computes and returns the hash of datum + pipeline, with a
// pipeline-specific prefix.
func HashDatum(pipelineName string, pipelineSalt string, inputs []*Input) string {
	hash := sha256.New()
	fmt.Println("hashing: ", len(inputs))
	for _, input := range inputs {
		hash.Write([]byte(input.Name))
		for _, fileInfo := range input.FileInfo {
			hash.Write([]byte(fileInfo.File.Path))
			// if i > 0 { // we want the groups to have different hashes, but we don't want to mess up the previous behavior here
			// 	hash.Write([]byte("group"))
			// }
			hash.Write(fileInfo.Hash)
		}
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

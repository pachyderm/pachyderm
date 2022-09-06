package common

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// MetaPrefix is the prefix for the meta path.
	MetaPrefix = "meta"
	// MetaFileName is the name of the meta file.
	MetaFileName = "meta"
	// PFSPrefix is the prefix for the pfs path.
	PFSPrefix = "pfs"
	// OutputPrefix is the prefix for the output path.
	OutputPrefix = "out"
	// TmpFileName is the name of the tmp file.
	TmpFileName = "tmp"
	TTL         = 15 * time.Minute
)

func MetaFilePath(id string) string {
	return path.Join(MetaPrefix, id, MetaFileName)
}

// DatumID computes the ID of a datum.
func DatumID(inputs []*Input) string {
	hash := pfs.NewHash()
	for _, input := range inputs {
		hash.Write([]byte(input.Name))
		_ = binary.Write(hash, binary.BigEndian, int64(len(input.Name)))
		file := input.FileInfo.File
		hash.Write([]byte(file.Commit.Branch.Repo.Name))
		_ = binary.Write(hash, binary.BigEndian, int64(len(file.Commit.Branch.Repo.Name)))
		hash.Write([]byte(file.Commit.Branch.Name))
		_ = binary.Write(hash, binary.BigEndian, int64(len(file.Commit.Branch.Name)))
		hash.Write([]byte(input.FileInfo.File.Path))
		_ = binary.Write(hash, binary.BigEndian, int64(len(input.FileInfo.File.Path)))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// HashDatum computes the hash of a datum.
func HashDatum(pipelineSalt string, inputs []*Input) string {
	hash := pfs.NewHash()
	id := DatumID(inputs)
	hash.Write([]byte(id))
	for _, input := range inputs {
		hash.Write([]byte(input.FileInfo.Hash))
	}
	hash.Write([]byte(pipelineSalt))
	return hex.EncodeToString(hash.Sum(nil))
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

// TODO: Trim non-meta file shards?
func Shard(pachClient *client.APIClient, fileSetIDs []string) ([]*pfs.PathRange, error) {
	var result []*pfs.PathRange
	for _, fileSetID := range fileSetIDs {
		shards, err := pachClient.ShardFileSet(fileSetID)
		if err != nil {
			return nil, err
		}
		if len(shards) > len(result) {
			result = shards
		}
	}
	return result, nil
}

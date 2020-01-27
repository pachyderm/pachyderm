package transform

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

func iteratorHashes(pipelineInfo *pps.PipelineInfo, dit datum.Iterator) []string {
	result := []string{}
	dit.Reset()
	for dit.Next() {
		result = append(result, common.HashDatum(pipelineInfo.Pipeline.Name, pipelineInfo.Salt, dit.Datum()))
	}
	return result
}

func hashesToMap(hashes []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, hash := range hashes {
		result[hash] = struct{}{}
	}
	return result
}

func requireHashes(t *testing.T, hashes []string, m map[string]struct{}) {
	require.Equal(t, len(hashes), len(m))
	for _, hash := range hashes {
		_, ok := m[hash]
		require.True(t, ok)
	}
}

func TestDatumSets(t *testing.T) {
	pipelineInfo := defaultPipelineInfo()
	jobs := []*pendingJob{}

	// Get the hashes for the first 20 indexes which we will use in this test
	dit := datum.NewMockIterator(&datum.MockIteratorOptions{Length: 20})
	hashes := iteratorHashes(pipelineInfo, dit)

	// Check that all the hashes are 'added' for an empty slate
	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Length: 5})
	added, removed, orphan := calculateDatumSets(pipelineInfo, dit, jobs, hashesToMap([]string{}))

	requireHashes(t, hashes[0:5], added)
	requireHashes(t, hashes[0:0], removed)
	require.True(t, orphan)

	// Check that extending the iterator will give us the new hashes as added
	datumsBase := hashesToMap(hashes[0:5])
	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Length: 7})
	added, removed, orphan = calculateDatumSets(pipelineInfo, dit, jobs, datumsBase)

	requireHashes(t, hashes[5:7], added)
	requireHashes(t, hashes[0:0], removed)
	require.False(t, orphan)

	// Check that sliding the iterator over gives us both added and removed hashes
	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Start: 3, Length: 7})
	added, removed, orphan = calculateDatumSets(pipelineInfo, dit, jobs, datumsBase)

	requireHashes(t, hashes[5:10], added)
	requireHashes(t, hashes[0:3], removed)
	require.False(t, orphan)

	// Check that a completely new range removes all existing datums and is an orphan
	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Start: 5, Length: 5})
	added, removed, orphan = calculateDatumSets(pipelineInfo, dit, jobs, datumsBase)

	requireHashes(t, hashes[5:10], added)
	requireHashes(t, hashes[0:5], removed)
	require.True(t, orphan)

	// Check that a chain of pending jobs results in the correct added/removed hashes
	// base:  0  1  2  3  4
	// job1:    -1           5
	// job2:     1       -4 -5
	// job3: -0 -1        4  5  6  7
	jobs = []*pendingJob{
		&pendingJob{datumsAdded: hashesToMap(hashes[5:6]), datumsRemoved: hashesToMap(hashes[1:2])},
		&pendingJob{datumsAdded: hashesToMap(hashes[1:2]), datumsRemoved: hashesToMap(hashes[4:6])},
		&pendingJob{datumsAdded: hashesToMap(hashes[4:8]), datumsRemoved: hashesToMap(hashes[0:2])},
	}

	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Length: 10})
	added, removed, orphan = calculateDatumSets(pipelineInfo, dit, jobs, datumsBase)

	requireHashes(t, []string{hashes[0], hashes[1], hashes[8], hashes[9]}, added)
	requireHashes(t, hashes[0:0], removed)
	require.False(t, orphan)

	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Start: 7, Length: 3})
	added, removed, orphan = calculateDatumSets(pipelineInfo, dit, jobs, datumsBase)

	requireHashes(t, hashes[8:10], added)
	requireHashes(t, hashes[2:7], removed)
	require.False(t, orphan)

	dit = datum.NewMockIterator(&datum.MockIteratorOptions{Start: 8, Length: 2})
	added, removed, orphan = calculateDatumSets(pipelineInfo, dit, jobs, datumsBase)

	requireHashes(t, hashes[8:10], added)
	requireHashes(t, hashes[2:8], removed)
	require.True(t, orphan)
}

func TestRegistry(t *testing.T) {
	err := withTestEnv(defaultPipelineInfo(), func(env *testEnv) error {
		/*
			reg, err := newRegistry(env.logger, env.driver)
			require.NoError(t, err)
		*/

		return nil
	})
	require.NoError(t, err)
}

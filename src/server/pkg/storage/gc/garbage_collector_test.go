package gc

import (
	"fmt"
	"testing"

	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var _ = fmt.Printf // TODO: remove after debugging is done

// TODO: create clients only once

func getPachClient(t *testing.T) *client.APIClient {
	pachClient, err := client.NewOnUserMachine(false, false, "user")
	require.NoError(t, err)
	return pachClient
}

func getGcClient(t *testing.T) *Client {
	gcc, err := NewClient(getPachClient(t), "localhost", 32228)
	require.NoError(t, err)
	return gcc
}

func initialize(t *testing.T) *Client {
	gcc := getGcClient(t)
	_, err := gcc.db.QueryContext(gcc.pachClient.Ctx(), "delete from chunks *; delete from refs *;")
	require.NoError(t, err)
	return gcc
}

type chunkRow struct {
	chunk    string
	deleting bool
}

func allChunkRows(t *testing.T, gcc *Client) []chunkRow {
	rows, err := gcc.db.QueryContext(gcc.pachClient.Ctx(), "select chunk, deleting from chunks")
	require.NoError(t, err)
	defer rows.Close()

	chunks := []chunkRow{}

	for rows.Next() {
		var chunk string
		var timestamp pq.NullTime
		require.NoError(t, rows.Scan(&chunk, &timestamp))
		chunks = append(chunks, chunkRow{chunk: chunk, deleting: timestamp.Valid})
	}
	require.NoError(t, rows.Err())

	return chunks
}

type refRow struct {
	sourcetype string
	source     string
	chunk      string
}

func allRefRows(t *testing.T, gcc *Client) []refRow {
	rows, err := gcc.db.QueryContext(gcc.pachClient.Ctx(), "select sourcetype, source, chunk from refs")
	require.NoError(t, err)
	defer rows.Close()

	refs := []refRow{}

	for rows.Next() {
		row := refRow{}
		require.NoError(t, rows.Scan(&row.sourcetype, &row.source, &row.chunk))
		refs = append(refs, row)
	}
	require.NoError(t, rows.Err())

	return refs
}

func printState(t *testing.T, gcc *Client) {
	chunkRows := allChunkRows(t, gcc)
	fmt.Printf("Chunks table:\n")
	for _, row := range chunkRows {
		fmt.Printf("  %v\n", row)
	}

	refRows := allRefRows(t, gcc)
	fmt.Printf("Refs table:\n")
	for _, row := range refRows {
		fmt.Printf("  %v\n", row)
	}
}

func makeJobs(count int) []string {
	result := []string{}
	for i := 0; i < count; i++ {
		result = append(result, fmt.Sprintf("job-%d", i))
	}
	return result
}

func makeChunks(count int) []chunk.Chunk {
	result := []chunk.Chunk{}
	for i := 0; i < count; i++ {
		result = append(result, chunk.Chunk{Hash: testutil.UniqueString(fmt.Sprintf("chunk-%d-", i))})
	}
	return result
}

func TestConnectivity(t *testing.T) {
	initialize(t)
}

func TestReserveChunks(t *testing.T) {
	gcc := initialize(t)
	jobs := makeJobs(3)
	chunks := makeChunks(3)

	// No chunks
	require.NoError(t, gcc.ReserveChunks(jobs[0], chunks[0:0]))

	// One chunk
	require.NoError(t, gcc.ReserveChunks(jobs[1], chunks[0:1]))

	// Multiple chunks
	require.NoError(t, gcc.ReserveChunks(jobs[2], chunks))

	expectedChunkRows := []chunkRow{
		{chunks[0].Hash, false},
		{chunks[1].Hash, false},
		{chunks[2].Hash, false},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunkRows(t, gcc))

	expectedRefRows := []refRow{
		{"job", jobs[1], chunks[0].Hash},
		{"job", jobs[2], chunks[0].Hash},
		{"job", jobs[2], chunks[1].Hash},
		{"job", jobs[2], chunks[2].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefRows(t, gcc))
}

func TestUpdateReferences(t *testing.T) {
	gcc := initialize(t)
	jobs := makeJobs(3)
	chunks := makeChunks(5)

	require.NoError(t, gcc.ReserveChunks(jobs[0], chunks[0:3])) // 0, 1, 2
	require.NoError(t, gcc.ReserveChunks(jobs[1], chunks[2:4])) // 2, 3
	require.NoError(t, gcc.ReserveChunks(jobs[2], chunks[4:5])) // 4

	// Currently, no links between chunks:
	// 0 1 2 3 4
	expectedChunkRows := []chunkRow{
		{chunks[0].Hash, false},
		{chunks[1].Hash, false},
		{chunks[2].Hash, false},
		{chunks[3].Hash, false},
		{chunks[4].Hash, false},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunkRows(t, gcc))

	expectedRefRows := []refRow{
		{"job", jobs[0], chunks[0].Hash},
		{"job", jobs[0], chunks[1].Hash},
		{"job", jobs[0], chunks[2].Hash},
		{"job", jobs[1], chunks[2].Hash},
		{"job", jobs[1], chunks[3].Hash},
		{"job", jobs[2], chunks[4].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefRows(t, gcc))

	require.NoError(t, gcc.UpdateReferences(
		[]Reference{{"chunk", chunks[4].Hash, chunks[0]}},
		[]Reference{},
		jobs[0:1],
	))

	// Chunk 1 should be cleaned up as unreferenced
	// 4 2 3 <- referenced by jobs
	// |
	// 0
	expectedChunkRows = []chunkRow{
		{chunks[0].Hash, false},
		{chunks[1].Hash, true},
		{chunks[2].Hash, false},
		{chunks[3].Hash, false},
		{chunks[4].Hash, false},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunkRows(t, gcc))

	expectedRefRows = []refRow{
		{"job", jobs[1], chunks[2].Hash},
		{"job", jobs[1], chunks[3].Hash},
		{"job", jobs[2], chunks[4].Hash},
		{"chunk", chunks[4].Hash, chunks[0].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefRows(t, gcc))

	require.NoError(t, gcc.UpdateReferences(
		[]Reference{
			{"semantic", "semantic-3", chunks[3]},
			{"semantic", "semantic-2", chunks[2]},
		},
		[]Reference{},
		jobs[1:2],
	))

	// No chunks should be cleaned up, the job reference to 2 and 3 were replaced
	// with semantic references
	// 2 3 <- referenced semantically
	// 4 <- referenced by jobs
	// |
	// 0
	expectedChunkRows = []chunkRow{
		{chunks[0].Hash, false},
		{chunks[1].Hash, true},
		{chunks[2].Hash, false},
		{chunks[3].Hash, false},
		{chunks[4].Hash, false},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunkRows(t, gcc))

	expectedRefRows = []refRow{
		{"job", jobs[2], chunks[4].Hash},
		{"chunk", chunks[4].Hash, chunks[0].Hash},
		{"semantic", "semantic-2", chunks[2].Hash},
		{"semantic", "semantic-3", chunks[3].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefRows(t, gcc))

	require.NoError(t, gcc.UpdateReferences(
		[]Reference{{"chunk", chunks[4].Hash, chunks[2]}},
		[]Reference{{"semantic", "semantic-3", chunks[3]}},
		jobs[2:3],
	))

	// Chunk 3 should be cleaned up as the semantic reference was removed
	// Chunk 4 should be cleaned up as the job reference was removed
	// Chunk 0 should be cleaned up later once chunk 4 has been removed
	// 2 <- referenced semantically
	// 0 <- referenced by 4 (deleting)
	expectedChunkRows = []chunkRow{
		{chunks[0].Hash, false},
		{chunks[1].Hash, true},
		{chunks[2].Hash, false},
		{chunks[3].Hash, true},
		{chunks[4].Hash, true},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunkRows(t, gcc))

	expectedRefRows = []refRow{
		{"chunk", chunks[4].Hash, chunks[0].Hash},
		{"chunk", chunks[4].Hash, chunks[2].Hash},
		{"semantic", "semantic-2", chunks[2].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefRows(t, gcc))
}

func TestFuzz(t *testing.T) {
	// Set up several parallel goroutines each with their own client

	// Set up a global ref registry with a mutex to track what the values should end up as

	// Run random add/remove operations, only allow references to go from lower numbers to higher numbers to keep it acyclic

	// Occasionally halt all goroutines and check reference counts in the db

	// Repeat for some amount of time
}

package gc

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var _ = fmt.Printf // TODO: remove after debugging is done

// TODO: create clients only once

func getPachClient(t testing.TB) *client.APIClient {
	pachClient, err := client.NewOnUserMachine(false, false, "user")
	require.NoError(t, err)
	return pachClient
}

func getGcClient(t *testing.T) *Client {
	client, err := NewClient(getPachClient(t), "localhost", 32228)
	require.NoError(t, err)
	return client
}

func makeJob() string {
	return testutil.UniqueString("job")
}

func makeChunk() chunk.Chunk {
	return chunk.Chunk{Hash: testutil.UniqueString("")}
}

func TestConnectivity(t *testing.T) {
	getGcClient(t)
}

func TestReserveChunks(t *testing.T) {
	gcc := getGcClient(t)
	chunks := []chunk.Chunk{makeChunk(), makeChunk(), makeChunk()}

	// No chunks
	require.NoError(t, gcc.ReserveChunks(makeJob(), chunks[0:0]))

	// One chunk
	require.NoError(t, gcc.ReserveChunks(makeJob(), chunks[0:1]))

	// Multiple chunks
	require.NoError(t, gcc.ReserveChunks(makeJob(), chunks))
}

func TestFuzz(t *testing.T) {
	// Set up several parallel goroutines each with their own client

	// Set up a global ref registry with a mutex to track what the values should end up as

	// Run random add/remove operations, only allow references to go from lower numbers to higher numbers to keep it acyclic

	// Occasionally halt all goroutines and check reference counts in the db

	// Repeat for some amount of time
}

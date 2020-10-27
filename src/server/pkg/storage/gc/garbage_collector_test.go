package gc

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

func TestReserveChunk(t *testing.T) {
	require.NoError(t, WithLocalGarbageCollector(func(ctx context.Context, objClient obj.Client, gcClient Client) error {
		chunks := makeChunks(t, objClient, 3)
		tmpID := uuid.NewWithoutDashes()
		expiresAt := getExpiresAt()
		for _, chunk := range chunks {
			require.NoError(t, gcClient.ReserveChunk(ctx, chunk, tmpID, expiresAt))
		}
		expectedChunkRows := []chunkModel{
			{chunks[0], nil},
			{chunks[1], nil},
			{chunks[2], nil},
		}
		require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
		expectedRefRows := []refModel{
			{"temporary", tmpID, chunks[0], time.Time{}, sql.NullTime{}},
			{"temporary", tmpID, chunks[1], time.Time{}, sql.NullTime{}},
			{"temporary", tmpID, chunks[2], time.Time{}, sql.NullTime{}},
		}
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		return nil
	}))
}

func TestCreateDeleteReferences(t *testing.T) {
	require.NoError(t, WithLocalGarbageCollector(func(ctx context.Context, objClient obj.Client, gcClient Client) error {
		chunks := makeChunks(t, objClient, 7)
		// Reserve chunks initially with only temporary references.
		tmpID := uuid.NewWithoutDashes()
		for _, chunk := range chunks {
			require.NoError(t, gcClient.ReserveChunk(ctx, chunk, tmpID, getExpiresAt()))
		}
		var expectedChunkRows []chunkModel
		for _, chunk := range chunks {
			expectedChunkRows = append(expectedChunkRows, chunkModel{chunk, nil})
		}
		require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
		var expectedRefRows []refModel
		for _, chunk := range chunks {
			expectedRefRows = append(expectedRefRows, refModel{"temporary", tmpID, chunk, time.Time{}, sql.NullTime{}})
		}
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		// Create cross-chunk references.
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				chunkFrom, chunkTo := chunks[i], chunks[i*2+1+j]
				require.NoError(t, gcClient.CreateReference(
					ctx,
					&Reference{
						Sourcetype: "chunk",
						Source:     chunkFrom,
						Chunk:      chunkTo,
					},
				))
				expectedRefRows = append(expectedRefRows, refModel{"chunk", chunkFrom, chunkTo, time.Time{}, sql.NullTime{}})
			}
		}
		// Create a semantic reference to the root chunk.
		semanticName := "root"
		require.NoError(t, gcClient.CreateReference(
			ctx,
			&Reference{
				Sourcetype: "semantic",
				Source:     semanticName,
				Chunk:      chunks[0],
			},
		))
		expectedRefRows = append(expectedRefRows, refModel{"semantic", semanticName, chunks[0], time.Time{}, sql.NullTime{}})
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		// Delete the temporary reference.
		require.NoError(t, gcClient.DeleteReference(
			ctx,
			&Reference{
				Sourcetype: "temporary",
				Source:     tmpID,
			},
		))
		time.Sleep(3 * time.Second)
		require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
		expectedRefRows = expectedRefRows[len(chunks):]
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		// Delete the semantic reference.
		require.NoError(t, gcClient.DeleteReference(
			ctx,
			&Reference{
				Sourcetype: "semantic",
				Source:     semanticName,
			},
		))
		time.Sleep(3 * time.Second)
		require.ElementsEqual(t, []chunkModel{}, allChunks(t, gcClient.(*client)))
		require.ElementsEqual(t, []refModel{}, allRefs(t, gcClient.(*client)))
		return nil
	}, WithPolling(time.Second)))
}

func TestRecovery(t *testing.T) {
	require.NoError(t, obj.WithLocalClient(func(objClient obj.Client) error {
		return WithLocalDB(func(db *gorm.DB) error {
			ctx := context.Background()
			gcClient, err := NewClient(db)
			require.NoError(t, err)
			semanticName := "root"
			expectedChunkRows := makeChunkTree(ctx, t, objClient, gcClient.(*client), semanticName, 3, 0)
			require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
			require.NoError(t, gcClient.DeleteReference(
				ctx,
				&Reference{
					Sourcetype: "semantic",
					Source:     semanticName,
				},
			))
			return WithGarbageCollector(objClient, db, func(_ context.Context, _ Client) error {
				time.Sleep(3 * time.Second)
				require.ElementsEqual(t, []chunkModel{}, allChunks(t, gcClient.(*client)))
				require.ElementsEqual(t, []refModel{}, allRefs(t, gcClient.(*client)))
				return nil
			})
		})
	}))
}

func TestTimeout(t *testing.T) {
	require.NoError(t, WithLocalGarbageCollector(func(ctx context.Context, objClient obj.Client, gcClient Client) error {
		numTrees := 10
		var expectedChunkRows []chunkModel
		for i := 0; i < numTrees; i++ {
			tmpID := uuid.NewWithoutDashes()
			expectedChunkRows = append(expectedChunkRows, makeChunkTree(ctx, t, objClient, gcClient.(*client), tmpID, 3, 0.2)...)
		}
		time.Sleep(3 * time.Second)
		require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
		return nil
	}, WithPolling(time.Second)))
}

func makeChunkTree(ctx context.Context, t *testing.T, objClient obj.Client, gcClient *client, semanticName string, levels int, failProb float64) []chunkModel {
	chunks := makeChunks(t, objClient, int(math.Pow(float64(2), float64(levels)))-1)
	// Reserve chunks initially with only temporary references.
	tmpID := uuid.NewWithoutDashes()
	var expectedChunkRows []chunkModel
	for _, chunk := range chunks {
		require.NoError(t, gcClient.ReserveChunk(ctx, chunk, tmpID, time.Now().Add(2*time.Second)))
		expectedChunkRows = append(expectedChunkRows, chunkModel{chunk, nil})
	}
	// Create cross-chunk references.
	nonLeafChunks := int(math.Pow(float64(2), float64(levels-1))) - 1
	for i := 0; i < nonLeafChunks; i++ {
		for j := 0; j < 2; j++ {
			if rand.Float64() < failProb {
				return nil
			}
			chunkFrom, chunkTo := chunks[i], chunks[i*2+1+j]
			require.NoError(t, gcClient.CreateReference(
				ctx,
				&Reference{
					Sourcetype: "chunk",
					Source:     chunkFrom,
					Chunk:      chunkTo,
				},
			))
		}
	}
	// Create a semantic reference to the root chunk.
	require.NoError(t, gcClient.CreateReference(
		ctx,
		&Reference{
			Sourcetype: "semantic",
			Source:     semanticName,
			Chunk:      chunks[0],
		},
	))
	// Delete the temporary reference.
	require.NoError(t, gcClient.DeleteReference(
		ctx,
		&Reference{
			Sourcetype: "temporary",
			Source:     tmpID,
		},
	))
	return expectedChunkRows
}

func makeChunks(t *testing.T, objClient obj.Client, count int) []string {
	result := []string{}
	for i := 0; i < count; i++ {
		chunkID := fmt.Sprintf("chunk-%d-", i) + uuid.NewWithoutDashes()[0:12]
		w, err := objClient.Writer(context.Background(), chunkID)
		require.NoError(t, err)
		require.NoError(t, w.Close())
		result = append(result, chunkID)
	}
	return result
}

func allChunks(t *testing.T, gcClient *client) []chunkModel {
	chunks := []chunkModel{}
	require.NoError(t, gcClient.db.Find(&chunks).Error)
	return chunks
}

func allRefs(t *testing.T, gcClient *client) []refModel {
	refs := []refModel{}
	require.NoError(t, gcClient.db.Find(&refs).Error)
	// Clear the created field because it makes testing difficult.
	for i := range refs {
		refs[i].Created = time.Time{}
	}
	return zeroExpiresAt(refs)
}

// This is necessary to make the times match.  require.ElementsEqual does not seem to handle time.Time well
func zeroExpiresAt(refs []refModel) []refModel {
	for i := range refs {
		refs[i].ExpiresAt = sql.NullTime{}
	}
	return refs
}

// Helper functions for when debugging
//func printState(t *testing.T, gcClient *client) {
//	fmt.Printf("Chunks table:\n")
//	for _, row := range allChunks(t, gcClient) {
//		fmt.Printf("  %v\n", row)
//	}
//
//	fmt.Printf("Refs table:\n")
//	for _, row := range allRefs(t, gcClient) {
//		fmt.Printf("  %v\n", row)
//	}
//}

func getExpiresAt() time.Time {
	return time.Now().Add(defaultTimeout).Truncate(time.Second).UTC()
}

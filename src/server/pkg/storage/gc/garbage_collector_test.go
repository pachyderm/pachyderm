package gc

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

func TestReserveChunk(t *testing.T) {
	require.NoError(t, WithLocalGarbageCollector(func(ctx context.Context, objClient obj.Client, gcClient Client) error {
		chunks := makeChunks(t, objClient, 3)
		tmpID := uuid.NewWithoutDashes()
		for _, chunk := range chunks {
			require.NoError(t, gcClient.ReserveChunk(ctx, chunk, tmpID))
		}
		expectedChunkRows := []chunkModel{
			{chunks[0], nil},
			{chunks[1], nil},
			{chunks[2], nil},
		}
		require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
		expectedRefRows := []refModel{
			{"temporary", tmpID, chunks[0], nil},
			{"temporary", tmpID, chunks[1], nil},
			{"temporary", tmpID, chunks[2], nil},
		}
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		return nil
	}))
}

func TestAddRemoveReferences(t *testing.T) {
	require.NoError(t, WithLocalGarbageCollector(func(ctx context.Context, objClient obj.Client, gcClient Client) error {
		chunks := makeChunks(t, objClient, 7)
		// Reserve chunks initially with only temporary references.
		tmpID := uuid.NewWithoutDashes()
		for _, chunk := range chunks {
			require.NoError(t, gcClient.ReserveChunk(ctx, chunk, tmpID))
		}
		var expectedChunkRows []chunkModel
		for _, chunk := range chunks {
			expectedChunkRows = append(expectedChunkRows, chunkModel{chunk, nil})
		}
		require.ElementsEqual(t, expectedChunkRows, allChunks(t, gcClient.(*client)))
		var expectedRefRows []refModel
		for _, chunk := range chunks {
			expectedRefRows = append(expectedRefRows, refModel{"temporary", tmpID, chunk, nil})
		}
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		// Add cross-chunk references.
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {
				chunkFrom, chunkTo := chunks[i], chunks[i*2+1+j]
				require.NoError(t, gcClient.AddReference(
					ctx,
					&Reference{
						Sourcetype: "chunk",
						Source:     chunkFrom,
						Chunk:      chunkTo,
					},
				))
				expectedRefRows = append(expectedRefRows, refModel{"chunk", chunkFrom, chunkTo, nil})
			}
		}
		// Add a semantic reference to the root chunk.
		semanticName := "root"
		require.NoError(t, gcClient.AddReference(
			ctx,
			&Reference{
				Sourcetype: "semantic",
				Source:     semanticName,
				Chunk:      chunks[0],
			},
		))
		expectedRefRows = append(expectedRefRows, refModel{"semantic", semanticName, chunks[0], nil})
		require.ElementsEqual(t, expectedRefRows, allRefs(t, gcClient.(*client)))
		// Remove the temporary reference.
		require.NoError(t, gcClient.RemoveReference(
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
		// Remove the semantic reference.
		require.NoError(t, gcClient.RemoveReference(
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
			require.NoError(t, gcClient.RemoveReference(
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
	}, WithPolling(time.Second), WithTimeout(time.Second)))
}

func makeChunkTree(ctx context.Context, t *testing.T, objClient obj.Client, gcClient *client, semanticName string, levels int, failProb float64) []chunkModel {
	chunks := makeChunks(t, objClient, int(math.Pow(float64(2), float64(levels)))-1)
	// Reserve chunks initially with only temporary references.
	tmpID := uuid.NewWithoutDashes()
	var expectedChunkRows []chunkModel
	for _, chunk := range chunks {
		require.NoError(t, gcClient.ReserveChunk(ctx, chunk, tmpID))
		expectedChunkRows = append(expectedChunkRows, chunkModel{chunk, nil})
	}
	// Add cross-chunk references.
	nonLeafChunks := int(math.Pow(float64(2), float64(levels-1))) - 1
	for i := 0; i < nonLeafChunks; i++ {
		for j := 0; j < 2; j++ {
			if rand.Float64() < failProb {
				return nil
			}
			chunkFrom, chunkTo := chunks[i], chunks[i*2+1+j]
			require.NoError(t, gcClient.AddReference(
				ctx,
				&Reference{
					Sourcetype: "chunk",
					Source:     chunkFrom,
					Chunk:      chunkTo,
				},
			))
		}
	}
	// Add a semantic reference to the root chunk.
	require.NoError(t, gcClient.AddReference(
		ctx,
		&Reference{
			Sourcetype: "semantic",
			Source:     semanticName,
			Chunk:      chunks[0],
		},
	))
	// Remove the temporary reference.
	require.NoError(t, gcClient.RemoveReference(
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
		chunkID := testutil.UniqueString(fmt.Sprintf("chunk-%d-", i))
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
		refs[i].Created = nil
	}
	return refs
}

// Helper functions for when debugging
func printMetrics(t *testing.T, f func() error) error {
	registry := prometheus.NewRegistry()
	initPrometheus(registry)
	defer func() {
		stats, err := registry.Gather()
		require.NoError(t, err)
		for _, family := range stats {
			fmt.Printf("%s (%d)\n", *family.Name, len(family.Metric))
			for _, metric := range family.Metric {
				labels := []string{}
				for _, pair := range metric.Label {
					labels = append(labels, fmt.Sprintf("%s:%s", *pair.Name, *pair.Value))
				}
				labelStr := strings.Join(labels, ",")
				if len(labelStr) == 0 {
					labelStr = "no labels"
				}

				if metric.Counter != nil {
					fmt.Printf(" %s: %d\n", labelStr, int64(*metric.Counter.Value))
				}

				if metric.Summary != nil {
					fmt.Printf(" %s: %d, %f\n", labelStr, *metric.Summary.SampleCount, *metric.Summary.SampleSum)
					for _, quantile := range metric.Summary.Quantile {
						fmt.Printf("  %f: %f\n", *quantile.Quantile, *quantile.Value)
					}
				}
			}
		}
	}()
	return f()
}

func printState(t *testing.T, gcClient *client) {
	fmt.Printf("Chunks table:\n")
	for _, row := range allChunks(t, gcClient) {
		fmt.Printf("  %v\n", row)
	}

	fmt.Printf("Refs table:\n")
	for _, row := range allRefs(t, gcClient) {
		fmt.Printf("  %v\n", row)
	}
}

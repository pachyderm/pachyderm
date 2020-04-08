package gc

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

func TestConnectivity(t *testing.T) {
	ctx := context.Background()
	initialize(t, ctx, nil)
}

func TestReserveChunk(t *testing.T) {
	ctx := context.Background()
	client := initialize(t, ctx, nil)
	chunks := makeChunks(3)
	tmpID := uuid.NewWithoutDashes()
	for _, chunk := range chunks {
		require.NoError(t, client.ReserveChunk(ctx, chunk, tmpID))
	}
	expectedChunkRows := []chunkModel{
		{chunks[0], nil},
		{chunks[1], nil},
		{chunks[2], nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))
	expectedRefRows := []refModel{
		{"temporary", tmpID, chunks[0], nil},
		{"temporary", tmpID, chunks[1], nil},
		{"temporary", tmpID, chunks[2], nil},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))
}

func TestAddRemoveReferences(t *testing.T) {
	ctx := context.Background()
	client := initialize(t, ctx, nil)
	chunks := makeChunks(7)
	// Reserve chunks initially with only temporary references.
	tmpID := uuid.NewWithoutDashes()
	for _, chunk := range chunks {
		require.NoError(t, client.ReserveChunk(ctx, chunk, tmpID))
	}
	var expectedChunkRows []chunkModel
	for _, chunk := range chunks {
		expectedChunkRows = append(expectedChunkRows, chunkModel{chunk, nil})
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))
	var expectedRefRows []refModel
	for _, chunk := range chunks {
		expectedRefRows = append(expectedRefRows, refModel{"temporary", tmpID, chunk, nil})
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))
	// Add cross-chunk references.
	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			chunkFrom, chunkTo := chunks[i], chunks[i*2+1+j]
			require.NoError(t, client.AddReference(
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
	require.NoError(t, client.AddReference(
		ctx,
		&Reference{
			Sourcetype: "semantic",
			Source:     semanticName,
			Chunk:      chunks[0],
		},
	))
	expectedRefRows = append(expectedRefRows, refModel{"semantic", semanticName, chunks[0], nil})
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))
	// Remove the temporary reference.
	require.NoError(t, client.RemoveReference(
		ctx,
		&Reference{
			Sourcetype: "temporary",
			Source:     tmpID,
		},
	))
	flushAllDeletes(client.server)
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))
	expectedRefRows = expectedRefRows[len(chunks):]
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))
	// Remove the semantic reference.
	require.NoError(t, client.RemoveReference(
		ctx,
		&Reference{
			Sourcetype: "semantic",
			Source:     semanticName,
		},
	))
	time.Sleep(3 * time.Second)
	flushAllDeletes(client.server)
	require.ElementsEqual(t, []chunkModel{}, allChunks(t, client))
	require.ElementsEqual(t, []refModel{}, allRefs(t, client))
}

func TestRecovery(t *testing.T) {
	ctx := context.Background()
	// Create a no-op server that will not perform deletions.
	client := initialize(t, ctx, &testServer{})
	// Create a chunk tree.
	semanticName := "root"
	expectedChunkRows := makeChunkTree(t, ctx, client, semanticName, 3, 0)
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))
	require.NoError(t, client.RemoveReference(
		ctx,
		&Reference{
			Sourcetype: "semantic",
			Source:     semanticName,
		},
	))
	// Create a real server and confirm that it performs the deletion at startup.
	client = newClient(t, ctx, nil, nil)
	time.Sleep(3 * time.Second)
	require.ElementsEqual(t, []chunkModel{}, allChunks(t, client))
	require.ElementsEqual(t, []refModel{}, allRefs(t, client))
}

func TestTimeout(t *testing.T) {
	ctx := context.Background()
	server := newServer(t, ctx, nil, nil, time.Second)
	client := initialize(t, ctx, server)
	numTrees := 10
	var expectedChunkRows []chunkModel
	for i := 0; i < numTrees; i++ {
		tmpID := uuid.NewWithoutDashes()
		expectedChunkRows = append(expectedChunkRows, makeChunkTree(t, ctx, client, tmpID, 3, 0.2)...)
	}
	time.Sleep(3 * time.Second)
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))
}

func makeChunkTree(t *testing.T, ctx context.Context, client *client, semanticName string, levels int, failProb float64) []chunkModel {
	chunks := makeChunks(int(math.Pow(float64(2), float64(levels))) - 1)
	// Reserve chunks initially with only temporary references.
	tmpID := uuid.NewWithoutDashes()
	var expectedChunkRows []chunkModel
	for _, chunk := range chunks {
		require.NoError(t, client.ReserveChunk(ctx, chunk, tmpID))
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
			require.NoError(t, client.AddReference(
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
	require.NoError(t, client.AddReference(
		ctx,
		&Reference{
			Sourcetype: "semantic",
			Source:     semanticName,
			Chunk:      chunks[0],
		},
	))
	// Remove the temporary reference.
	require.NoError(t, client.RemoveReference(
		ctx,
		&Reference{
			Sourcetype: "temporary",
			Source:     tmpID,
		},
	))
	return expectedChunkRows
}

func initialize(t *testing.T, ctx context.Context, server Server) *client {
	client := newClient(t, ctx, server, nil)
	clearData(t, client)
	return client
}

func newClient(t *testing.T, ctx context.Context, server Server, metrics prometheus.Registerer) *client {
	if server == nil {
		server = newServer(t, ctx, nil, metrics, 0)
	}
	c, err := NewClient(server, "localhost", 32228, metrics)
	require.NoError(t, err)
	return c.(*client)
}

func newServer(t *testing.T, ctx context.Context, deleter Deleter, metrics prometheus.Registerer, timeout time.Duration) Server {
	if deleter == nil {
		deleter = &testDeleter{}
	}
	s, err := NewServer(ctx, deleter, "localhost", 32228, metrics, timeout)
	require.NoError(t, err)
	return s
}

func clearData(t *testing.T, client *client) {
	require.NoError(t, client.db.Exec("delete from chunks *; delete from refs *;").Error)
}

func makeTemporaryIDs(count int) []string {
	result := []string{}
	for i := 0; i < count; i++ {
		result = append(result, fmt.Sprintf("temporary-%d", i))
	}
	return result
}

func makeChunks(count int) []string {
	result := []string{}
	for i := 0; i < count; i++ {
		result = append(result, testutil.UniqueString(fmt.Sprintf("chunk-%d-", i)))
	}
	return result
}

func allChunks(t *testing.T, client *client) []chunkModel {
	chunks := []chunkModel{}
	require.NoError(t, client.db.Find(&chunks).Error)
	return chunks
}

func allRefs(t *testing.T, client *client) []refModel {
	refs := []refModel{}
	require.NoError(t, client.db.Find(&refs).Error)
	// Clear the created field because it makes testing difficult.
	for i := range refs {
		refs[i].Created = nil
	}
	return refs
}

// Dummy deleter object for testing, so we don't need an object storage
type testDeleter struct{}

func (td *testDeleter) Delete(ctx context.Context, chunks []string) error {
	return nil
}

// Dummy server for testing recovery.
type testServer struct{}

func (ts *testServer) DeleteChunk(_ context.Context, _ string) {}

func (ts *testServer) FlushDelete(_ context.Context, _ string) error {
	return nil
}

// Helper function to ensure the server is done deleting to avoid races
func flushAllDeletes(s Server) {
	if _, ok := s.(*testServer); ok {
		return
	}
	// bork bork
	si := s.(*server)
	for {
		t := func() *trigger {
			si.mutex.Lock()
			defer si.mutex.Unlock()
			for _, t := range si.deleting {
				return t
			}
			return nil
		}()
		if t == nil {
			return
		}
		t.Wait()
	}
}

// Helper functions for when debugging
func printMetrics(t *testing.T, metrics prometheus.Gatherer) {
	stats, err := metrics.Gather()
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
}

func printState(t *testing.T, client *client) {
	fmt.Printf("Chunks table:\n")
	for _, row := range allChunks(t, client) {
		fmt.Printf("  %v\n", row)
	}

	fmt.Printf("Refs table:\n")
	for _, row := range allRefs(t, client) {
		fmt.Printf("  %v\n", row)
	}
}

type fuzzDeleter struct {
	mutex sync.Mutex
	users map[string]int
}

func (fd *fuzzDeleter) forEach(chunks []string, cb func(string) error) error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for _, chunk := range chunks {
		if err := cb(chunk); err != nil {
			return err
		}
	}
	return nil
}

// Delete will make sure that only one thing tries to delete or use the chunk
// at one time.
func (fd *fuzzDeleter) Delete(ctx context.Context, chunks []string) error {
	err := fd.forEach(chunks, func(hash string) error {
		if fd.users[hash] != 0 {
			return fmt.Errorf("Failed to delete chunk (%s), already in use: %d", hash, fd.users[hash])
		}
		fd.users[hash] = -1
		return nil
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Float32()*50) * time.Millisecond)

	err = fd.forEach(chunks, func(hash string) error {
		if fd.users[hash] != -1 {
			return fmt.Errorf("Failed to release chunk, something changed it during deletion: %d", fd.users[hash])
		}
		delete(fd.users, hash)
		return nil
	})
	return err
}

// updating will make sure the chunks are not deleted during the call to cb
func (fd *fuzzDeleter) updating(chunks []string) error {
	err := fd.forEach(chunks, func(hash string) error {
		if fd.users[hash] == -1 {
			return fmt.Errorf("Failed to use chunk, currently being deleted")
		}
		fd.users[hash]++
		return nil
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Float32()*50) * time.Millisecond)

	err = fd.forEach(chunks, func(hash string) error {
		if fd.users[hash] < 1 {
			return fmt.Errorf("Failed to release chunk, inconsistency")
		}
		fd.users[hash]--
		return nil
	})
	return err
}

func TestFuzz(t *testing.T) {
	metrics := prometheus.NewRegistry()
	deleter := &fuzzDeleter{users: make(map[string]int)}
	ctx := context.Background()
	server := newServer(t, ctx, deleter, metrics, 0)
	client := newClient(t, ctx, server, metrics)
	seed := time.Now().UTC().UnixNano()
	seed = 1561162472664846205
	rng := rand.New(rand.NewSource(seed))
	clearData(t, client)

	fmt.Printf("Fuzzer running with seed: %v\n", seed)
	fmt.Printf("first 10 rng:\n")
	for i := 0; i < 10; i++ {
		fmt.Printf("  %d\n", rng.Intn(1000))
	}

	type jobData struct {
		id     string
		add    []Reference
		remove []Reference
	}

	runJob := func(ctx context.Context, client Client, job jobData) error {
		// Make a list of chunks we'll be referencing and reserve them
		chunkMap := make(map[string]bool)
		for _, item := range job.add {
			chunkMap[item.Chunk] = true
			if item.Sourcetype == "chunk" {
				chunkMap[item.Source] = true
			}
		}
		reservedChunks := []string{}
		for hash := range chunkMap {
			reservedChunks = append(reservedChunks, hash)
		}

		for _, chunk := range reservedChunks {
			if err := client.ReserveChunk(ctx, chunk, job.id); err != nil {
				return err
			}
		}

		// Check that no one deletes the chunk while we have it reserved
		if err := deleter.updating(reservedChunks); err != nil {
			return err
		}

		for _, add := range job.add {
			if err := client.AddReference(ctx, &add); err != nil {
				return err
			}
		}
		for _, remove := range job.remove {
			if err := client.RemoveReference(ctx, &remove); err != nil {
				return err
			}
		}
		return client.RemoveReference(ctx, &Reference{
			Sourcetype: "temporary",
			Source:     job.id,
		})
	}

	startWorkers := func(numWorkers int, jobChannel chan jobData) *errgroup.Group {
		eg, ctx := errgroup.WithContext(context.Background())
		for i := 0; i < numWorkers; i++ {
			eg.Go(func() error {
				client := newClient(t, ctx, server, metrics)
				client.db.DB().SetMaxOpenConns(1)
				client.db.DB().SetMaxIdleConns(1)
				for x := range jobChannel {
					if err := runJob(ctx, client, x); err != nil {
						fmt.Printf("client failed: %v\n", err)
						closeErr := client.db.Close()
						if closeErr != nil {
							fmt.Printf("error when closing: %v\n", closeErr)
						}
						return err
					}
				}
				return client.db.Close()
			})
		}
		return eg
	}

	// Global ref registry to track what the database should look like
	chunkRefs := make(map[int][]Reference)
	freeChunkIDs := []int{}
	maxID := -1

	usedChunkID := func(lowerBound int) int {
		ids := []int{}
		// TODO: this might get annoyingly slow, but go doesn't have good sorted data structures
		for id := range chunkRefs {
			if id >= lowerBound {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return 0
		}
		return ids[rng.Intn(len(ids))]
	}

	unusedChunkID := func(lowerBound int) int {
		for i, id := range freeChunkIDs {
			if id >= lowerBound {
				freeChunkIDs = append(freeChunkIDs[0:i], freeChunkIDs[i+1:len(freeChunkIDs)]...)
				return id
			}
		}
		maxID++
		return maxID
	}

	addRef := func() Reference {
		var dest int
		ref := Reference{}
		roll := rng.Float32()
		if len(chunkRefs) == 0 || roll > 0.9 {
			// Make a new chunk with a semantic reference
			dest = unusedChunkID(0)
			ref.Sourcetype = "semantic"
			ref.Source = testutil.UniqueString("semantic-")
			ref.Chunk = fmt.Sprintf("%d", dest)
		} else if roll > 0.6 {
			// Add a semantic reference to an existing chunk
			dest = usedChunkID(0)
			ref.Sourcetype = "semantic"
			ref.Source = testutil.UniqueString("semantic-")
			ref.Chunk = fmt.Sprintf("%d", dest)
		} else {
			source := usedChunkID(0)
			ref.Sourcetype = "chunk"
			ref.Source = fmt.Sprintf("%d", source)
			if roll > 0.5 {
				// Add a cross-chunk reference to a new chunk
				dest = unusedChunkID(source + 1)
				ref.Chunk = fmt.Sprintf("%d", dest)
			} else {
				// Add a cross-chunk reference to an existing chunk
				dest = usedChunkID(source + 1)
				if dest == 0 {
					dest = unusedChunkID(source + 1)
				}
				ref.Chunk = fmt.Sprintf("%d", dest)
			}
		}

		existingChunkRefs := chunkRefs[dest]
		if existingChunkRefs == nil {
			chunkRefs[dest] = []Reference{ref}
		} else {
			// Make sure ref isn't a duplicate
			for _, x := range existingChunkRefs {
				if ref.Sourcetype == x.Sourcetype && ref.Source == x.Source {
					ref.Sourcetype = "semantic"
					ref.Source = testutil.UniqueString("semantic-")
					break
				}
			}

			chunkRefs[dest] = append(existingChunkRefs, ref)
		}

		return ref
	}

	// Recursively remove references from the specified chunk (simulating a delete)
	var cleanupRef func(id int)
	cleanupRef = func(id int) {
		require.Equal(t, 0, len(chunkRefs[id]))

		idStr := fmt.Sprintf("%d", id)
		delete(chunkRefs, id)
		freeChunkIDs = append(freeChunkIDs, id)

		for recursiveID, refs := range chunkRefs {
			i := 0
			for _, ref := range refs {
				if ref.Source != idStr {
					refs[i] = ref
					i++
				} else {
					fmt.Printf("removing chunk reference: %s/%s:%s\n", ref.Sourcetype, ref.Source, ref.Chunk)
				}
			}
			refs = refs[:i]
			chunkRefs[recursiveID] = refs

			if len(refs) == 0 {
				fmt.Printf("recursively removing unreferenced chunk: %d\n", recursiveID)
				cleanupRef(recursiveID)
			}
		}
	}

	removeRef := func() Reference {
		if len(chunkRefs) == 0 {
			panic("no references to remove")
		}

		i := rng.Intn(len(chunkRefs))
		for k, v := range chunkRefs {
			if i == 0 {
				j := rng.Intn(len(v))
				ref := v[j]
				chunkRefs[k] = append(v[0:j], v[j+1:len(v)]...)
				if len(chunkRefs[k]) == 0 {
					fmt.Printf("removing unreferenced chunk: %d\n", k)
					cleanupRef(k)
				}
				return ref
			}
			i--
		}
		panic("unreachable")
	}

	makeJobData := func() jobData {
		jd := jobData{id: testutil.UniqueString("job-")}

		for len(jd.add) == 0 && len(jd.remove) == 0 {
			// Run random add/remove operations, only allow references to go from lower numbers to higher numbers to keep it acyclic
			for rng.Float32() > 0.5 {
				numAdds := rng.Intn(10)
				for i := 0; i < numAdds; i++ {
					jd.add = append(jd.add, addRef())
				}
			}

			for rng.Float32() > 0.6 {
				numRemoves := rng.Intn(7)
				for i := 0; i < numRemoves && len(chunkRefs) > 0; i++ {
					jd.remove = append(jd.remove, removeRef())
				}
			}
		}

		fmt.Printf("job: %v\n", jd)

		return jd
	}

	// chunkRefs := make(map[int][]Reference)
	verifyData := func() {
		chunks := allChunks(t, client)
		//refs := allRefs(t, client)

		chunkIDs := []int{}
		chunksByID := make(map[int]bool)
		expectedChunkIDs := []int{}
		expectedChunksByID := make(map[int]bool)

		for _, c := range chunks {
			require.Nil(t, c.Deleting)
			id, err := strconv.ParseInt(c.Chunk, 10, 64)
			require.NoError(t, err)
			chunkIDs = append(chunkIDs, int(id))
			chunksByID[int(id)] = true
		}

		for id := range chunkRefs {
			expectedChunkIDs = append(expectedChunkIDs, id)
			expectedChunksByID[id] = true
		}

		for id := range chunksByID {
			if !expectedChunksByID[id] {
				fmt.Printf("Extra chunk in database: %v\n", id)
			}
		}

		for id := range expectedChunksByID {
			if !chunksByID[id] {
				fmt.Printf("Missing chunk in database: %v\n", id)
			}
		}

		require.ElementsEqual(t, expectedChunkIDs, chunkIDs)

		//for _, r := range refs {
		//	// Do a thing
		//}
	}

	numWorkers := 1
	numJobs := 10
	// Occasionally halt all goroutines and check data consistency
	for i := 0; i < 1; i++ {
		jobChan := make(chan jobData, numJobs)
		eg := startWorkers(numWorkers, jobChan)
		for i := 0; i < numJobs; i++ {
			jobChan <- makeJobData()
		}
		close(jobChan)
		require.NoError(t, eg.Wait())

		flushAllDeletes(server)
		verifyData()
	}

	printMetrics(t, metrics)
}

func TestMetrics(t *testing.T) {
}

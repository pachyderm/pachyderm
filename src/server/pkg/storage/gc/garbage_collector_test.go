package gc

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

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

func printState(t *testing.T, client *clientImpl) {
	fmt.Printf("Chunks table:\n")
	for _, row := range allChunks(t, client) {
		fmt.Printf("  %v\n", row)
	}

	fmt.Printf("Refs table:\n")
	for _, row := range allRefs(t, client) {
		fmt.Printf("  %v\n", row)
	}
}

// Dummy deleter object for testing, so we don't need an object storage
type testDeleter struct{}

func (td *testDeleter) Delete(ctx context.Context, chunks []string) error {
	return nil
}

// Helper function to ensure the server is done deleting to avoid races
func flushAllDeletes(server Server) {
	// bork bork
	si := server.(*serverImpl)
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

func makeServer(t *testing.T, deleter Deleter, metrics prometheus.Registerer) Server {
	if deleter == nil {
		deleter = &testDeleter{}
	}
	server, err := MakeServer(deleter, "localhost", 32228, metrics)
	require.NoError(t, err)
	return server
}

func makeClient(t *testing.T, server Server, metrics prometheus.Registerer) *clientImpl {
	if server == nil {
		server = makeServer(t, nil, metrics)
	}
	client, err := MakeClient(server, "localhost", 32228, metrics)
	require.NoError(t, err)
	return client.(*clientImpl)
}

func clearData(t *testing.T, client *clientImpl) {
	require.NoError(t, client.db.Exec("delete from chunks *; delete from refs *;").Error)
}

func initialize(t *testing.T) *clientImpl {
	client := makeClient(t, nil, nil)
	clearData(t, client)
	return client
}

func allChunks(t *testing.T, client *clientImpl) []chunkModel {
	chunks := []chunkModel{}
	require.NoError(t, client.db.Find(&chunks).Error)
	return chunks
}

func allRefs(t *testing.T, client *clientImpl) []refModel {
	refs := []refModel{}
	require.NoError(t, client.db.Find(&refs).Error)
	return refs
}

func makeJobs(count int) []string {
	result := []string{}
	for i := 0; i < count; i++ {
		result = append(result, fmt.Sprintf("job-%d", i))
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

func TestConnectivity(t *testing.T) {
	initialize(t)
}

func TestReserveChunks(t *testing.T) {
	ctx := context.Background()
	client := initialize(t)
	jobs := makeJobs(3)
	chunks := makeChunks(3)

	// No chunks
	require.NoError(t, client.ReserveChunks(ctx, jobs[0], chunks[0:0]))

	// One chunk
	require.NoError(t, client.ReserveChunks(ctx, jobs[1], chunks[0:1]))

	// Multiple chunks
	require.NoError(t, client.ReserveChunks(ctx, jobs[2], chunks))

	expectedChunkRows := []chunkModel{
		{chunks[0], nil},
		{chunks[1], nil},
		{chunks[2], nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))

	expectedRefRows := []refModel{
		{"job", jobs[1], chunks[0]},
		{"job", jobs[2], chunks[0]},
		{"job", jobs[2], chunks[1]},
		{"job", jobs[2], chunks[2]},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))
}

func TestUpdateReferences(t *testing.T) {
	ctx := context.Background()
	client := initialize(t)
	jobs := makeJobs(3)
	chunks := makeChunks(5)

	require.NoError(t, client.ReserveChunks(ctx, jobs[0], chunks[0:3])) // 0, 1, 2
	require.NoError(t, client.ReserveChunks(ctx, jobs[1], chunks[2:4])) // 2, 3
	require.NoError(t, client.ReserveChunks(ctx, jobs[2], chunks[4:5])) // 4

	// Currently, no links between chunks:
	// 0 1 2 3 4
	expectedChunkRows := []chunkModel{
		{chunks[0], nil},
		{chunks[1], nil},
		{chunks[2], nil},
		{chunks[3], nil},
		{chunks[4], nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))

	expectedRefRows := []refModel{
		{"job", jobs[0], chunks[0]},
		{"job", jobs[0], chunks[1]},
		{"job", jobs[0], chunks[2]},
		{"job", jobs[1], chunks[2]},
		{"job", jobs[1], chunks[3]},
		{"job", jobs[2], chunks[4]},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))

	require.NoError(t, client.UpdateReferences(
		ctx,
		[]Reference{{"chunk", chunks[4], chunks[0]}},
		[]Reference{},
		jobs[0],
	))
	flushAllDeletes(client.server)

	// Chunk 1 should be cleaned up as unreferenced
	// 4 2 3 <- referenced by jobs
	// |
	// 0
	expectedChunkRows = []chunkModel{
		{chunks[0], nil},
		{chunks[2], nil},
		{chunks[3], nil},
		{chunks[4], nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))

	expectedRefRows = []refModel{
		{"job", jobs[1], chunks[2]},
		{"job", jobs[1], chunks[3]},
		{"job", jobs[2], chunks[4]},
		{"chunk", chunks[4], chunks[0]},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))

	require.NoError(t, client.UpdateReferences(
		ctx,
		[]Reference{
			{"semantic", "semantic-3", chunks[3]},
			{"semantic", "semantic-2", chunks[2]},
			{"chunk", chunks[2], chunks[3]},
		},
		[]Reference{},
		jobs[1],
	))
	flushAllDeletes(client.server)

	// No chunks should be cleaned up, the job reference to 2 and 3 were replaced
	// with semantic references
	// 2 <- referenced semantically
	// |
	// 3 <- referenced semantically
	//
	// 4 <- referenced by jobs
	// |
	// 0
	expectedChunkRows = []chunkModel{
		{chunks[0], nil},
		{chunks[2], nil},
		{chunks[3], nil},
		{chunks[4], nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))

	expectedRefRows = []refModel{
		{"job", jobs[2], chunks[4]},
		{"chunk", chunks[4], chunks[0]},
		{"semantic", "semantic-2", chunks[2]},
		{"semantic", "semantic-3", chunks[3]},
		{"chunk", chunks[2], chunks[3]},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))

	require.NoError(t, client.UpdateReferences(
		ctx,
		[]Reference{{"chunk", chunks[4], chunks[2]}},
		[]Reference{{"semantic", "semantic-3", chunks[3]}, {"chunk", chunks[2], chunks[3]}},
		jobs[2],
	))
	flushAllDeletes(client.server)

	// Chunk 3 should be cleaned up as the semantic reference was removed
	// Chunk 4 should be cleaned up as the job reference was removed
	// Chunk 0 should be cleaned up later once chunk 4 has been removed
	// 2 <- referenced semantically
	expectedChunkRows = []chunkModel{
		{chunks[2], nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, client))

	expectedRefRows = []refModel{
		{"semantic", "semantic-2", chunks[2]},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, client))
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
	server := makeServer(t, deleter, metrics)
	client := makeClient(t, server, metrics)
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

		if err := client.ReserveChunks(ctx, job.id, reservedChunks); err != nil {
			return err
		}

		// Check that no one deletes the chunk while we have it reserved
		if err := deleter.updating(reservedChunks); err != nil {
			return err
		}

		return client.UpdateReferences(ctx, job.add, job.remove, job.id)
	}

	startWorkers := func(numWorkers int, jobChannel chan jobData) *errgroup.Group {
		eg, ctx := errgroup.WithContext(context.Background())
		for i := 0; i < numWorkers; i++ {
			eg.Go(func() error {
				client := makeClient(t, server, metrics)
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

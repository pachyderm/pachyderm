package gc

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

var _ = fmt.Printf // TODO: remove after debugging is done
var _ = sql.OpenDB
var _ = math.Abs

// Dummy deleter object for testing, so we don't need an object storage
type testDeleter struct{}

func (td *testDeleter) Delete(ctx context.Context, chunks []chunk.Chunk) error {
	return nil
}

func makeServer(t *testing.T, deleter Deleter, metrics prometheus.Registerer) Server {
	if deleter == nil {
		deleter = &testDeleter{}
	}
	server, err := MakeServer(deleter, "localhost", 32228, metrics)
	require.NoError(t, err)
	return server
}

func makeClient(t *testing.T, server Server, metrics prometheus.Registerer) *ClientImpl {
	if server == nil {
		server = makeServer(t, nil, metrics)
	}
	gcc, err := MakeClient(context.Background(), server, "localhost", 32228, metrics)
	require.NoError(t, err)
	return gcc.(*ClientImpl)
}

func clearData(t *testing.T, ctx context.Context, gcc *ClientImpl) {
	require.NoError(t, gcc.db.Exec("delete from chunks *; delete from refs *;").Error)
}

func initialize(t *testing.T, ctx context.Context) *ClientImpl {
	gcc := makeClient(t, nil, nil)
	clearData(t, ctx, gcc)
	return gcc
}

func allChunks(t *testing.T, ctx context.Context, gcc *ClientImpl) []chunkModel {
	chunks := []chunkModel{}
	require.NoError(t, gcc.db.Find(&chunks).Error)
	return chunks
}

func allRefs(t *testing.T, ctx context.Context, gcc *ClientImpl) []refModel {
	refs := []refModel{}
	require.NoError(t, gcc.db.Find(&refs).Error)
	return refs
}

func printState(t *testing.T, ctx context.Context, gcc *ClientImpl) {
	fmt.Printf("Chunks table:\n")
	for _, row := range allChunks(t, ctx, gcc) {
		fmt.Printf("  %v\n", row)
	}

	fmt.Printf("Refs table:\n")
	for _, row := range allRefs(t, ctx, gcc) {
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
	initialize(t, context.Background())
}

func TestReserveChunks(t *testing.T) {
	ctx := context.Background()
	gcc := initialize(t, ctx)
	jobs := makeJobs(3)
	chunks := makeChunks(3)

	// No chunks
	require.NoError(t, gcc.ReserveChunks(ctx, jobs[0], chunks[0:0]))

	// One chunk
	require.NoError(t, gcc.ReserveChunks(ctx, jobs[1], chunks[0:1]))

	// Multiple chunks
	require.NoError(t, gcc.ReserveChunks(ctx, jobs[2], chunks))

	expectedChunkRows := []chunkModel{
		{chunks[0].Hash, nil},
		{chunks[1].Hash, nil},
		{chunks[2].Hash, nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, ctx, gcc))

	expectedRefRows := []refModel{
		{"job", jobs[1], chunks[0].Hash},
		{"job", jobs[2], chunks[0].Hash},
		{"job", jobs[2], chunks[1].Hash},
		{"job", jobs[2], chunks[2].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, ctx, gcc))
}

func TestUpdateReferences(t *testing.T) {
	ctx := context.Background()
	gcc := initialize(t, ctx)
	jobs := makeJobs(3)
	chunks := makeChunks(5)

	// deletes are handled asynchronously on the server, flush everything to avoid race conditions
	flush := func() {
		require.NoError(t, gcc.server.FlushDeletes(ctx, chunks))
	}

	require.NoError(t, gcc.ReserveChunks(ctx, jobs[0], chunks[0:3])) // 0, 1, 2
	require.NoError(t, gcc.ReserveChunks(ctx, jobs[1], chunks[2:4])) // 2, 3
	require.NoError(t, gcc.ReserveChunks(ctx, jobs[2], chunks[4:5])) // 4

	// Currently, no links between chunks:
	// 0 1 2 3 4
	expectedChunkRows := []chunkModel{
		{chunks[0].Hash, nil},
		{chunks[1].Hash, nil},
		{chunks[2].Hash, nil},
		{chunks[3].Hash, nil},
		{chunks[4].Hash, nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, ctx, gcc))

	expectedRefRows := []refModel{
		{"job", jobs[0], chunks[0].Hash},
		{"job", jobs[0], chunks[1].Hash},
		{"job", jobs[0], chunks[2].Hash},
		{"job", jobs[1], chunks[2].Hash},
		{"job", jobs[1], chunks[3].Hash},
		{"job", jobs[2], chunks[4].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, ctx, gcc))

	require.NoError(t, gcc.UpdateReferences(
		ctx,
		[]Reference{{"chunk", chunks[4].Hash, chunks[0]}},
		[]Reference{},
		jobs[0],
	))
	flush()

	// Chunk 1 should be cleaned up as unreferenced
	// 4 2 3 <- referenced by jobs
	// |
	// 0
	expectedChunkRows = []chunkModel{
		{chunks[0].Hash, nil},
		{chunks[2].Hash, nil},
		{chunks[3].Hash, nil},
		{chunks[4].Hash, nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, ctx, gcc))

	expectedRefRows = []refModel{
		{"job", jobs[1], chunks[2].Hash},
		{"job", jobs[1], chunks[3].Hash},
		{"job", jobs[2], chunks[4].Hash},
		{"chunk", chunks[4].Hash, chunks[0].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, ctx, gcc))

	require.NoError(t, gcc.UpdateReferences(
		ctx,
		[]Reference{
			{"semantic", "semantic-3", chunks[3]},
			{"semantic", "semantic-2", chunks[2]},
		},
		[]Reference{},
		jobs[1],
	))
	flush()

	// No chunks should be cleaned up, the job reference to 2 and 3 were replaced
	// with semantic references
	// 2 3 <- referenced semantically
	// 4 <- referenced by jobs
	// |
	// 0
	expectedChunkRows = []chunkModel{
		{chunks[0].Hash, nil},
		{chunks[2].Hash, nil},
		{chunks[3].Hash, nil},
		{chunks[4].Hash, nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, ctx, gcc))

	expectedRefRows = []refModel{
		{"job", jobs[2], chunks[4].Hash},
		{"chunk", chunks[4].Hash, chunks[0].Hash},
		{"semantic", "semantic-2", chunks[2].Hash},
		{"semantic", "semantic-3", chunks[3].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, ctx, gcc))

	require.NoError(t, gcc.UpdateReferences(
		ctx,
		[]Reference{{"chunk", chunks[4].Hash, chunks[2]}},
		[]Reference{{"semantic", "semantic-3", chunks[3]}},
		jobs[2],
	))

	// We do two flushes because this update does a transitive delete and we
	// don't want this test to be flaky by landing on an intermediary state
	flush()
	flush()

	// Chunk 3 should be cleaned up as the semantic reference was removed
	// Chunk 4 should be cleaned up as the job reference was removed
	// Chunk 0 should be cleaned up later once chunk 4 has been removed
	// 2 <- referenced semantically
	expectedChunkRows = []chunkModel{
		{chunks[2].Hash, nil},
	}
	require.ElementsEqual(t, expectedChunkRows, allChunks(t, ctx, gcc))

	expectedRefRows = []refModel{
		{"semantic", "semantic-2", chunks[2].Hash},
	}
	require.ElementsEqual(t, expectedRefRows, allRefs(t, ctx, gcc))
}

type fuzzDeleter struct {
	mutex sync.Mutex
	users map[string]int
}

func (fd *fuzzDeleter) forEach(chunks []chunk.Chunk, cb func(string) error) error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	for _, chunk := range chunks {
		if err := cb(chunk.Hash); err != nil {
			return err
		}
	}
	return nil
}

// Delete will make sure that only one thing tries to delete or use the chunk
// at one time.
func (fd *fuzzDeleter) Delete(ctx context.Context, chunks []chunk.Chunk) error {
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
func (fd *fuzzDeleter) updating(chunks []chunk.Chunk) error {
	err := fd.forEach(chunks, func(hash string) error {
		if fd.users[hash] == -1 {
			return fmt.Errorf("Failed to use chunk, currently being deleted")
		}
		fd.users[hash] += 1
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
		fd.users[hash] -= 1
		return nil
	})
	return err
}

func TestFuzz(t *testing.T) {
	metrics := prometheus.NewRegistry()
	deleter := &fuzzDeleter{users: make(map[string]int)}
	server := makeServer(t, deleter, metrics)
	client := makeClient(t, server, metrics)
	clearData(t, context.Background(), client)

	type jobData struct {
		id     string
		add    []Reference
		remove []Reference
	}

	runJob := func(ctx context.Context, gcc Client, job jobData) error {
		// Make a list of chunks we'll be referencing and reserve them
		chunkMap := make(map[string]bool)
		for _, item := range job.add {
			chunkMap[item.chunk.Hash] = true
			if item.sourcetype == "chunk" {
				chunkMap[item.source] = true
			}
		}
		reservedChunks := []chunk.Chunk{}
		for hash := range chunkMap {
			reservedChunks = append(reservedChunks, chunk.Chunk{Hash: hash})
		}

		if err := gcc.ReserveChunks(ctx, job.id, reservedChunks); err != nil {
			return err
		}

		// Check that no one deletes the chunk while we have it reserved
		if err := deleter.updating(reservedChunks); err != nil {
			return err
		}

		return gcc.UpdateReferences(ctx, job.add, job.remove, job.id)
	}

	startWorkers := func(numWorkers int, jobChannel chan jobData) *errgroup.Group {
		eg, ctx := errgroup.WithContext(context.Background())
		for i := 0; i < numWorkers; i++ {
			eg.Go(func() error {
				gcc := makeClient(t, server, metrics)
				gcc.db.DB().SetMaxOpenConns(1)
				gcc.db.DB().SetMaxIdleConns(1)
				for x := range jobChannel {
					if err := runJob(ctx, gcc, x); err != nil {
						fmt.Printf("client failed: %v\n", err)
						closeErr := gcc.db.Close()
						if closeErr != nil {
							fmt.Printf("error when closing: %v\n", closeErr)
						}
						return err
					}
				}
				return gcc.db.Close()
			})
		}
		return eg
	}

	// Global ref registry to track what the database should look like
	chunkRefs := make(map[int][]Reference)
	freeChunkIds := []int{}
	maxId := -1

	usedChunkId := func(lowerBound int) int {
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
		return ids[rand.Intn(len(ids))]
	}

	unusedChunkId := func(lowerBound int) int {
		for i, id := range freeChunkIds {
			if id >= lowerBound {
				freeChunkIds = append(freeChunkIds[0:i], freeChunkIds[i+1:len(freeChunkIds)]...)
				return id
			}
		}
		maxId++
		return maxId
	}

	addRef := func() Reference {
		var dest int
		ref := Reference{}
		roll := rand.Float32()
		if len(chunkRefs) == 0 || roll > 0.9 {
			// Make a new chunk with a semantic reference
			dest = unusedChunkId(0)
			ref.sourcetype = "semantic"
			ref.source = testutil.UniqueString("semantic-")
			ref.chunk = chunk.Chunk{Hash: fmt.Sprintf("chunk-%d", dest)}
		} else if roll > 0.6 {
			// Add a semantic reference to an existing chunk
			dest = usedChunkId(0)
			ref.sourcetype = "semantic"
			ref.source = testutil.UniqueString("semantic-")
			ref.chunk = chunk.Chunk{Hash: fmt.Sprintf("chunk-%d", dest)}
		} else {
			source := usedChunkId(0)
			ref.sourcetype = "chunk"
			ref.source = fmt.Sprintf("chunk-%d", source)
			if roll > 0.5 {
				// Add a cross-chunk reference to a new chunk
				dest = unusedChunkId(source + 1)
				ref.chunk = chunk.Chunk{Hash: fmt.Sprintf("chunk-%d", dest)}
			} else {
				// Add a cross-chunk reference to an existing chunk
				dest = usedChunkId(source + 1)
				if dest == 0 {
					dest = unusedChunkId(source + 1)
				}
				ref.chunk = chunk.Chunk{Hash: fmt.Sprintf("chunk-%d", dest)}
			}
		}

		existingChunkRefs := chunkRefs[dest]
		if existingChunkRefs == nil {
			chunkRefs[dest] = []Reference{ref}
		} else {
			// Make sure ref isn't a duplicate
			for _, x := range existingChunkRefs {
				if ref.sourcetype == x.sourcetype && ref.source == x.source {
					ref.sourcetype = "semantic"
					ref.source = testutil.UniqueString("semantic-")
					break
				}
			}

			chunkRefs[dest] = append(existingChunkRefs, ref)
		}

		return ref
	}

	removeRef := func() Reference {
		if len(chunkRefs) == 0 {
			panic("no references to remove")
		}

		i := rand.Intn(len(chunkRefs))
		for k, v := range chunkRefs {
			if i == 0 {
				j := rand.Intn(len(v))
				ref := v[j]
				chunkRefs[k] = append(v[0:j], v[j+1:len(v)]...)
				if len(chunkRefs[k]) == 0 {
					delete(chunkRefs, k)
					freeChunkIds = append(freeChunkIds, k)
				}
				return ref
			}
			i--
		}
		panic("unreachable")
	}

	makeJobData := func() jobData {
		jd := jobData{id: testutil.UniqueString("job-")}

		// Run random add/remove operations, only allow references to go from lower numbers to higher numbers to keep it acyclic
		for rand.Float32() > 0.5 {
			numAdds := rand.Intn(10)
			for i := 0; i < numAdds; i++ {
				jd.add = append(jd.add, addRef())
			}
		}

		for rand.Float32() > 0.6 {
			numRemoves := rand.Intn(7)
			for i := 0; i < numRemoves && len(chunkRefs) > 0; i++ {
				jd.remove = append(jd.remove, removeRef())
			}
		}

		return jd
	}

	verifyData := func() {
		fmt.Printf("verifyData\n")
	}

	numWorkers := 10
	numJobs := 1000
	// Occasionally halt all goroutines and check data consistency
	for i := 0; i < 5; i++ {
		jobChan := make(chan jobData, numJobs)
		eg := startWorkers(numWorkers, jobChan)
		for i := 0; i < numJobs; i++ {
			jobChan <- makeJobData()
		}
		close(jobChan)
		require.NoError(t, eg.Wait())

		verifyData()
	}

	printMetrics(t, metrics)
}

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

func TestMetrics(t *testing.T) {
	metrics := prometheus.NewRegistry()
	counter :=
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "pachyderm",
				Subsystem: "gc",
				Name:      "test_counter",
				Help:      "disregard this",
			},
			[]string{"result"},
		)

	err := metrics.Register(counter)
	require.NoError(t, err)

	counter.WithLabelValues("five").Inc()

	printMetrics(t, metrics)
}

func TestGorm(t *testing.T) {
	db, err := openDatabase("localhost", 32228)
	require.NoError(t, err)
	initializeDb(db)
}

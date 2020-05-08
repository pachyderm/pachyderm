package testing

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"modernc.org/mathutil"
)

type loadConfig struct {
	pachdConfig *serviceenv.PachdFullConfiguration
	branchGens  []branchGenerator
}

func newLoadConfig(opts ...loadConfigOption) *loadConfig {
	config := &loadConfig{}
	config.pachdConfig = newPachdConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type loadConfigOption func(*loadConfig)

// (bryce) will use later, commenting to make linter happy.
//func withPachdConfig(opts ...pachdConfigOption) loadConfigOption {
//	return func(config *loadConfig) {
//		config.pachdConfig = newPachdConfig(opts...)
//	}
//}

func withBranchGenerator(opts ...branchGeneratorOption) loadConfigOption {
	return func(config *loadConfig) {
		config.branchGens = append(config.branchGens, newBranchGenerator(opts...))
	}
}

func newPachdConfig(opts ...pachdConfigOption) *serviceenv.PachdFullConfiguration {
	config := &serviceenv.PachdFullConfiguration{}
	config.NewStorageLayer = true
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelZeroSize = units.MB
	config.StorageGCPolling = "30s"
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// (bryce) this should probably be moved to the corresponding packages with configuration available
type pachdConfigOption func(*serviceenv.PachdFullConfiguration)

// (bryce) will use later, commenting to make linter happy.
//func withMemoryThreshold(memoryThreshold int64) pachdConfigOption {
//	return func(c *serviceenv.PachdFullConfiguration) {
//		c.StorageMemoryThreshold = memoryThreshold
//	}
//}
//
//func withShardThreshold(shardThreshold int64) pachdConfigOption {
//	return func(c *serviceenv.PachdFullConfiguration) {
//		c.StorageShardThreshold = shardThreshold
//	}
//}
//
//func withLevelZeroSize(levelZeroSize int64) pachdConfigOption {
//	return func(c *serviceenv.PachdFullConfiguration) {
//		c.StorageLevelZeroSize = levelZeroSize
//	}
//}
//
//func withGCPolling(polling string) pachdConfigOption {
//	return func(c *serviceenv.PachdFullConfiguration) {
//		c.StorageGCPolling = polling
//	}
//}

type branchGenerator func(*client.APIClient, string, *loadState) error

func newBranchGenerator(opts ...branchGeneratorOption) branchGenerator {
	config := &branchConfig{
		name:      "master",
		validator: newValidator(),
	}
	for _, opt := range opts {
		opt(config)
	}
	return func(c *client.APIClient, repo string, state *loadState) error {
		for _, gen := range config.commitGens {
			if err := gen(c, repo, config.name, state, config.validator); err != nil {
				return err
			}
		}
		return nil
	}
}

type branchConfig struct {
	name       string
	commitGens []commitGenerator
	validator  *validator
}

type branchGeneratorOption func(config *branchConfig)

func withCommitGenerator(opts ...commitGeneratorOption) branchGeneratorOption {
	return func(config *branchConfig) {
		config.commitGens = append(config.commitGens, newCommitGenerator(opts...))
	}
}

type commitGenerator func(*client.APIClient, string, string, *loadState, *validator) error

func newCommitGenerator(opts ...commitGeneratorOption) commitGenerator {
	config := &commitConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return func(c *client.APIClient, repo, branch string, state *loadState, validator *validator) error {
		for i := 0; i < config.count; i++ {
			commit, err := c.StartCommit(repo, branch)
			if err != nil {
				return err
			}
			for _, gen := range config.putTarGens {
				fs := make(map[string][]byte)
				r, err := gen(state, fs)
				if err != nil {
					return err
				}
				if config.putThroughputConfig != nil && rand.Float64() < config.putThroughputConfig.prob {
					r = newThroughputLimitReader(r, config.putThroughputConfig.limit)
				}
				if config.putCancelConfig != nil && rand.Float64() < config.putCancelConfig.prob {
					// (bryce) not sure if we want to do anything with errors here?
					cancelOperation(config.putCancelConfig, c, func(c *client.APIClient) error {
						err := c.PutTar(repo, commit.ID, r)
						if err == nil {
							validator.recordFileSet(fs)
						}
						return err

					})
					continue
				}
				if err := c.PutTar(repo, commit.ID, r); err != nil {
					return err
				}
				validator.recordFileSet(fs)
			}
			if err := c.FinishCommit(repo, commit.ID); err != nil {
				return err
			}
			getTar := func(c *client.APIClient) error {
				r, err := c.GetTar(repo, commit.ID, "/")
				if err != nil {
					return err
				}
				if config.getThroughputConfig != nil && rand.Float64() < config.getThroughputConfig.prob {
					r = newThroughputLimitReader(r, config.getThroughputConfig.limit)
				}
				return validator.validate(r)
			}
			if config.getCancelConfig != nil && rand.Float64() < config.getCancelConfig.prob {
				cancelOperation(config.getCancelConfig, c, getTar)
				continue
			}
			if err := getTar(c); err != nil {
				return err
			}
		}
		return nil
	}
}

type commitConfig struct {
	count                                    int
	putTarGens                               []putTarGenerator
	putThroughputConfig, getThroughputConfig *throughputConfig
	putCancelConfig, getCancelConfig         *cancelConfig
}

type throughputConfig struct {
	limit int
	prob  float64
}

func newThroughputConfig(limit int, prob ...float64) *throughputConfig {
	config := &throughputConfig{
		limit: limit,
		prob:  1.0,
	}
	if len(prob) > 0 {
		config.prob = prob[0]
	}
	return config
}

type throughputLimitReader struct {
	r                               io.Reader
	bytesSinceSleep, bytesPerSecond int
}

func newThroughputLimitReader(r io.Reader, bytesPerSecond int) *throughputLimitReader {
	return &throughputLimitReader{
		r:              r,
		bytesPerSecond: bytesPerSecond,
	}
}

func (tlr *throughputLimitReader) Read(data []byte) (int, error) {
	var bytesRead int
	for len(data) > 0 {
		size := mathutil.Min(len(data), tlr.bytesPerSecond-tlr.bytesSinceSleep)
		n, err := tlr.r.Read(data[:size])
		data = data[n:]
		bytesRead += n
		if err != nil {
			return bytesRead, err
		}
		if tlr.bytesSinceSleep == tlr.bytesPerSecond {
			time.Sleep(time.Second)
			tlr.bytesSinceSleep = 0
		}
	}
	return bytesRead, nil
}

type cancelConfig struct {
	maxTime time.Duration
	prob    float64
}

func newCancelConfig(maxTime time.Duration, prob ...float64) *cancelConfig {
	config := &cancelConfig{
		maxTime: maxTime,
		prob:    1.0,
	}
	if len(prob) > 0 {
		config.prob = prob[0]
	}
	return config
}

func cancelOperation(cc *cancelConfig, c *client.APIClient, f func(c *client.APIClient) error) error {
	cancelCtx, cancel := context.WithCancel(c.Ctx())
	c = c.WithCtx(cancelCtx)
	go func() {
		<-time.After(time.Duration(int64(float64(int64(cc.maxTime)) * rand.Float64())))
		cancel()
	}()
	return f(c)
}

type commitGeneratorOption func(config *commitConfig)

func withCommitCount(count int) commitGeneratorOption {
	return func(config *commitConfig) {
		config.count = count
	}
}

func withPutTarGenerator(opts ...putTarGeneratorOption) commitGeneratorOption {
	return func(config *commitConfig) {
		config.putTarGens = append(config.putTarGens, newPutTarGenerator(opts...))
	}
}

func withPutThroughputLimit(limit int, prob ...float64) commitGeneratorOption {
	return func(config *commitConfig) {
		config.putThroughputConfig = newThroughputConfig(limit, prob...)
	}
}

func withGetThroughputLimit(limit int, prob ...float64) commitGeneratorOption {
	return func(config *commitConfig) {
		config.getThroughputConfig = newThroughputConfig(limit, prob...)
	}
}

func withPutCancel(maxTime time.Duration, prob ...float64) commitGeneratorOption {
	return func(config *commitConfig) {
		config.putCancelConfig = newCancelConfig(maxTime, prob...)
	}
}

func withGetCancel(maxTime time.Duration, prob ...float64) commitGeneratorOption {
	return func(config *commitConfig) {
		config.getCancelConfig = newCancelConfig(maxTime, prob...)
	}
}

type putTarGenerator func(*loadState, fileSetSpec) (io.Reader, error)

func newPutTarGenerator(opts ...putTarGeneratorOption) putTarGenerator {
	config := &putTarConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return func(state *loadState, files fileSetSpec) (io.Reader, error) {
		putTarBuf := &bytes.Buffer{}
		fileBuf := &bytes.Buffer{}
		tw := tar.NewWriter(io.MultiWriter(putTarBuf, fileBuf))
		for i := 0; i < config.count; i++ {
			for _, gen := range config.fileGens {
				name, err := gen(tw, state)
				if err != nil {
					return nil, err
				}
				// Record serialized tar entry for validation.
				files.recordFile(name, fileBuf)
				fileBuf.Reset()
			}
		}
		if err := tw.Close(); err != nil {
			return nil, err
		}
		return putTarBuf, nil
	}
}

type putTarConfig struct {
	count    int
	fileGens []fileGenerator
}

type putTarGeneratorOption func(config *putTarConfig)

func withFileCount(count int) putTarGeneratorOption {
	return func(config *putTarConfig) {
		config.count = count
	}
}

func withFileGenerator(gen fileGenerator) putTarGeneratorOption {
	return func(config *putTarConfig) {
		config.fileGens = append(config.fileGens, gen)
	}
}

type fileGenerator func(*tar.Writer, *loadState) (string, error)

func newRandomFileGenerator(opts ...randomFileGeneratorOption) fileGenerator {
	config := &randomFileConfig{
		fileSizeBuckets: defaultFileSizeBuckets(),
	}
	for _, opt := range opts {
		opt(config)
	}
	// (bryce) might want some validation for the file size buckets (total prob adds up to 1.0)
	return func(tw *tar.Writer, state *loadState) (string, error) {
		name := uuid.NewWithoutDashes()
		var totalProb float64
		sizeProb := rand.Float64()
		var min, max int
		for _, bucket := range config.fileSizeBuckets {
			totalProb += bucket.prob
			if sizeProb <= totalProb {
				min, max = bucket.min, bucket.max
				break
			}
		}
		size := min
		if max > min {
			size += rand.Intn(max - min)
		}
		state.Lock(func() {
			if size > state.sizeLeft {
				size = state.sizeLeft
			}
			state.sizeLeft -= size
		})
		return name, writeFile(tw, name, chunk.RandSeq(size))
	}
}

type randomFileConfig struct {
	fileSizeBuckets []fileSizeBucket
}

type fileSizeBucket struct {
	min, max int
	prob     float64
}

var (
	fileSizeBuckets = []fileSizeBucket{
		fileSizeBucket{
			min: 1 * units.KB,
			max: 10 * units.KB,
		},
		fileSizeBucket{
			min: 10 * units.KB,
			max: 100 * units.KB,
		},
		fileSizeBucket{
			min: 1 * units.MB,
			max: 10 * units.MB,
		},
		fileSizeBucket{
			min: 10 * units.MB,
			max: 100 * units.MB,
		},
	}
	// (bryce) will use later, commenting to make linter happy.
	//edgeCaseFileSizeBuckets = []fileSizeBucket{
	//	fileSizeBucket{
	//		min: 0,
	//	},
	//	fileSizeBucket{
	//		min: 1,
	//	},
	//	fileSizeBucket{
	//		min: 1,
	//		max: 100,
	//	},
	//}
)

func defaultFileSizeBuckets() []fileSizeBucket {
	buckets := append([]fileSizeBucket{}, fileSizeBuckets...)
	buckets[0].prob = 0.4
	buckets[1].prob = 0.4
	buckets[2].prob = 0.2
	return buckets
}

type randomFileGeneratorOption func(*randomFileConfig)

func withFileSizeBuckets(buckets []fileSizeBucket) randomFileGeneratorOption {
	return func(config *randomFileConfig) {
		config.fileSizeBuckets = buckets
	}
}

func writeFile(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name: "/" + name,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	if err != nil {
		return err
	}
	return tw.Flush()
}

// (bryce) this should be somewhere else (probably testutil).
func seedRand(customSeed ...int64) {
	seed := time.Now().UTC().UnixNano()
	if len(customSeed) > 0 {
		seed = customSeed[0]
	}
	rand.Seed(seed)
	fmt.Println("seed: ", strconv.FormatInt(seed, 10))
}

func TestLoad(t *testing.T) {
	seedRand()
	// (bryce) this is so dumb, but through a combination of the linter and not being
	// able to deploy the new storage layer in CI (particularly postgres), I have decided
	// to just ignore the error that will be produced by running TestLoad without postgres
	// setup for the time being.
	testLoad(fuzzLoad())
}

func fuzzLoad() *loadConfig {
	return newLoadConfig(
		withBranchGenerator(
			fuzzCommits()...,
		),
	)
}

func fuzzCommits() []branchGeneratorOption {
	var branchOpts []branchGeneratorOption
	for i := 0; i < 5; i++ {
		var commitOpts []commitGeneratorOption
		commitOpts = append(commitOpts, fuzzThroughputLimit()...)
		commitOpts = append(commitOpts, fuzzCancel()...)
		commitOpts = append(commitOpts,
			withCommitCount(rand.Intn(5)),
			withPutTarGenerator(
				withFileCount(rand.Intn(5)),
				withFileGenerator(fuzzFiles()),
			),
		)
		branchOpts = append(branchOpts, withCommitGenerator(commitOpts...))
	}
	return branchOpts
}

func fuzzFiles() fileGenerator {
	buckets := append([]fileSizeBucket{}, fileSizeBuckets...)
	rand.Shuffle(len(buckets), func(i, j int) { buckets[i], buckets[j] = buckets[j], buckets[i] })
	totalProb := 1.0
	for i := 0; i < len(buckets); i++ {
		buckets[i].prob = rand.Float64() * totalProb
		totalProb -= buckets[i].prob
	}
	buckets[len(buckets)-1].prob += totalProb
	return newRandomFileGenerator(withFileSizeBuckets(buckets))
}

func fuzzThroughputLimit() []commitGeneratorOption {
	return []commitGeneratorOption{
		withPutThroughputLimit(
			1*units.MB,
			0.05,
		),
		withGetThroughputLimit(
			1*units.MB,
			0.05,
		),
	}
}

func fuzzCancel() []commitGeneratorOption {
	return []commitGeneratorOption{
		withPutCancel(
			5*time.Second,
			0.05,
		),
		withGetCancel(
			5*time.Second,
			0.05,
		),
	}
}

func testLoad(loadConfig *loadConfig) error {
	return testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		c := env.PachClient
		state := &loadState{
			sizeLeft: units.GB,
		}
		repo := "test"
		if err := c.CreateRepo(repo); err != nil {
			return err
		}
		var eg errgroup.Group
		for _, branchGen := range loadConfig.branchGens {
			// (bryce) need a ctx here.
			branchGen := branchGen
			eg.Go(func() error {
				return branchGen(c, repo, state)
			})
		}
		return eg.Wait()
	}, loadConfig.pachdConfig)
}

type loadState struct {
	sizeLeft int
	mu       sync.Mutex
}

func (ls *loadState) Lock(f func()) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	f()
}

type validator struct {
	files      fileSetSpec
	sampleProb float64
}

func newValidator() *validator {
	return &validator{
		files:      make(map[string][]byte),
		sampleProb: 1.0,
	}
}

func (v *validator) recordFileSet(files fileSetSpec) {
	for file, data := range files {
		v.files.recordFile(file, bytes.NewReader(data))
	}
}

func (v *validator) validate(r io.Reader) error {
	hdr, err := tar.NewReader(r).Next()
	if err != nil {
		// We expect an empty tar stream if no files were uploaded.
		if err == io.EOF && len(v.files) == 0 {
			return nil
		}
		return err
	}
	if hdr.Name != "/" {
		return errors.Errorf("expected root header, got %v", hdr.Name)
	}
	var filesSorted []string
	for file := range v.files {
		filesSorted = append(filesSorted, file)
	}
	sort.Strings(filesSorted)
	for _, file := range filesSorted {
		buf := &bytes.Buffer{}
		if _, err := io.CopyN(buf, r, int64(len(v.files[file]))); err != nil {
			return err
		}
		if !bytes.Equal(v.files[file], buf.Bytes()) {
			return errors.Errorf("file %v's header and/or content is incorrect", file)
		}
	}
	return nil
}

type fileSetSpec map[string][]byte

func (fs fileSetSpec) recordFile(name string, r io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}
	fs[name] = buf.Bytes()
	return nil
}

func (fs fileSetSpec) makeTarStream() io.Reader {
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)
	for name, data := range fs {
		if err := writeFile(tw, name, data); err != nil {
			panic(err)
		}
	}
	if err := tw.Close(); err != nil {
		panic(err)
	}
	return buf
}

func finfosToPaths(finfos []*pfs.FileInfoNewStorage) (paths []string) {
	for _, finfo := range finfos {
		paths = append(paths, finfo.File.Path)
	}
	return paths
}

func TestListFileNS(t *testing.T) {
	config := &serviceenv.PachdFullConfiguration{}
	config.NewStorageLayer = true
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(repo))

		commit1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)

		fsSpec := fileSetSpec{
			"dir1/file1.1": []byte{},
			"dir1/file1.2": []byte{},
			"dir2/file2.1": []byte{},
			"dir2/file2.2": []byte{},
		}
		err = env.PachClient.PutTar(repo, commit1.ID, fsSpec.makeTarStream())
		require.NoError(t, err)

		err = env.PachClient.FinishCommit(repo, commit1.ID)
		require.NoError(t, err)

		finfos, err := env.PachClient.ListFileNS(repo, commit1.ID, "/dir1/*")
		require.NoError(t, err)
		t.Log(finfos)
		require.ElementsEqual(t, []string{"/dir1/file1.1", "/dir1/file1.2"}, finfosToPaths(finfos))
		return nil
	}, config)
	require.NoError(t, err)
}

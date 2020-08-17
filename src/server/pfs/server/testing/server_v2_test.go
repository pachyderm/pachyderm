package testing

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"sort"
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
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/stretchr/testify/assert"
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

// TODO Will use later, commenting to make linter happy.
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
	config.StorageV2 = true
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelZeroSize = units.MB
	config.StorageGCPolling = "30s"
	config.StorageCompactionMaxFanIn = 2
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// TODO This should probably be moved to the corresponding packages with configuration available
type pachdConfigOption func(*serviceenv.PachdFullConfiguration)

// TODO Will use later, commenting to make linter happy.
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
				fs := make(map[string]tarutil.File)
				r, err := gen(state, fs)
				if err != nil {
					return err
				}
				if config.putThroughputConfig != nil && rand.Float64() < config.putThroughputConfig.prob {
					r = newThroughputLimitReader(r, config.putThroughputConfig.limit)
				}
				if config.putCancelConfig != nil && rand.Float64() < config.putCancelConfig.prob {
					// TODO Not sure if we want to do anything with errors here?
					cancelOperation(config.putCancelConfig, c, func(c *client.APIClient) error {
						err := c.PutTarV2(repo, commit.ID, r)
						if err == nil {
							validator.recordFileSet(fs)
						}
						return err

					})
					continue
				}
				if err := c.PutTarV2(repo, commit.ID, r); err != nil {
					return err
				}
				if rand.Float64() < config.deleteProb {
					file := validator.deleteRandomFile()
					if file != "" {
						if err := c.DeleteFilesV2(repo, commit.ID, []string{file}); err != nil {
							return err
						}
					}
				}
				validator.recordFileSet(fs)
			}
			if err := c.FinishCommit(repo, commit.ID); err != nil {
				return err
			}
			getTar := func(c *client.APIClient) error {
				r, err := c.GetTarV2(repo, commit.ID, "**")
				if err != nil {
					return err
				}
				if config.getThroughputConfig != nil && rand.Float64() < config.getThroughputConfig.prob {
					r = newThroughputLimitReader(r, config.getThroughputConfig.limit)
				}
				validator.validate(r)
				return nil
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
	deleteProb                               float64
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

func withDeleteProb(prob float64) commitGeneratorOption {
	return func(config *commitConfig) {
		config.deleteProb = prob
	}
}

type putTarGenerator func(*loadState, fileSetSpec) (io.Reader, error)

func newPutTarGenerator(opts ...putTarGeneratorOption) putTarGenerator {
	config := &putTarConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return func(state *loadState, files fileSetSpec) (io.Reader, error) {
		buf := &bytes.Buffer{}
		if err := tarutil.WithWriter(buf, func(tw *tar.Writer) error {
			for i := 0; i < config.count; i++ {
				for _, gen := range config.fileGens {
					file, err := gen(tw, state)
					if err != nil {
						return err
					}
					// Record serialized tar entry for validation.
					files.recordFile(file)
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return buf, nil
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

type fileGenerator func(*tar.Writer, *loadState) (tarutil.File, error)

func newRandomFileGenerator(opts ...randomFileGeneratorOption) fileGenerator {
	config := &randomFileConfig{
		fileSizeBuckets: defaultFileSizeBuckets(),
	}
	for _, opt := range opts {
		opt(config)
	}
	// TODO Might want some validation for the file size buckets (total prob adds up to 1.0)
	return func(tw *tar.Writer, state *loadState) (tarutil.File, error) {
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
		file := tarutil.NewMemFile("/"+name, chunk.RandSeq(size))
		if err := tarutil.WriteFile(tw, file); err != nil {
			return nil, err
		}
		return file, nil
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
	// TODO Will use later, commenting to make linter happy.
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

func TestLoad(t *testing.T) {
	// TODO Remove once postgres runs in CI.
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	msg := tu.SeedRand()
	require.NoError(t, testLoad(t, fuzzLoad()), msg)
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
		commitOpts = append(commitOpts, fuzzDelete()...)
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

func fuzzDelete() []commitGeneratorOption {
	return []commitGeneratorOption{
		withDeleteProb(0.3),
	}
}

func testLoad(t *testing.T, loadConfig *loadConfig) error {
	return testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		c := env.PachClient
		state := &loadState{
			t:        t,
			sizeLeft: units.GB,
		}
		repo := "test"
		if err := c.CreateRepo(repo); err != nil {
			return err
		}
		var eg errgroup.Group
		for _, branchGen := range loadConfig.branchGens {
			// TODO Need a ctx here.
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
	t        *testing.T
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
		files:      make(map[string]tarutil.File),
		sampleProb: 1.0,
	}
}

func (v *validator) recordFileSet(files fileSetSpec) {
	for _, file := range files {
		v.files.recordFile(file)
	}
}

func (v *validator) validate(r io.Reader) (retErr error) {
	var namesSorted []string
	for name := range v.files {
		namesSorted = append(namesSorted, name)
	}
	sort.Strings(namesSorted)
	if len(namesSorted) > 0 {
		namesSorted = append([]string{"/"}, namesSorted...)
	}
	defer func() {
		if retErr == nil {
			if len(namesSorted) != 0 {
				retErr = errors.Errorf("got back less files than expected")
			}
		}
	}()
	return tarutil.Iterate(r, func(file tarutil.File) error {
		if len(namesSorted) == 0 {
			return errors.Errorf("got back more files than expected")
		}
		hdr, err := file.Header()
		if err != nil {
			return err
		}
		if hdr.Name == "/" && namesSorted[0] == "/" {
			namesSorted = namesSorted[1:]
			return nil
		}
		ok, err := tarutil.Equal(v.files[namesSorted[0]], file)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf("file %v's header and/or content is incorrect", namesSorted[0])
		}
		namesSorted = namesSorted[1:]
		return nil
	})
}

func (v *validator) deleteRandomFile() string {
	for name := range v.files {
		delete(v.files, name)
		return name
	}
	return ""
}

type fileSetSpec map[string]tarutil.File

func (fs fileSetSpec) recordFile(file tarutil.File) error {
	hdr, err := file.Header()
	if err != nil {
		return err
	}
	fs[hdr.Name] = file
	return nil
}

func (fs fileSetSpec) makeTarStream() io.Reader {
	buf := &bytes.Buffer{}
	if err := tarutil.WithWriter(buf, func(tw *tar.Writer) error {
		for _, file := range fs {
			if err := tarutil.WriteFile(tw, file); err != nil {
				panic(err)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return buf
}

func finfosToPaths(finfos []*pfs.FileInfo) (paths []string) {
	for _, finfo := range finfos {
		paths = append(paths, finfo.File.Path)
	}
	return paths
}

func TestListFileV2(t *testing.T) {
	// TODO: remove once postgres runs in CI
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}

	config := newPachdConfig()
	require.NoError(t, testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(repo))

		commit1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)

		fsSpec := fileSetSpec{}
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("dir1/file1.1", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("dir1/file1.2", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("dir2/file2.1", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("dir2/file2.2", []byte{})))
		err = env.PachClient.PutTarV2(repo, commit1.ID, fsSpec.makeTarStream())
		require.NoError(t, err)

		err = env.PachClient.FinishCommit(repo, commit1.ID)
		require.NoError(t, err)
		// should list a directory but not siblings
		finfos := []*pfs.FileInfo{}
		err = env.PachClient.ListFileV2(repo, commit1.ID, "/dir1", func(finfo *pfs.FileInfo) error {
			finfos = append(finfos, finfo)
			return nil
		})
		require.NoError(t, err)
		require.ElementsEqual(t, []string{"/dir1/file1.1", "/dir1/file1.2"}, finfosToPaths(finfos))
		// should list the root
		finfos = []*pfs.FileInfo{}
		err = env.PachClient.ListFileV2(repo, commit1.ID, "/", func(finfo *pfs.FileInfo) error {
			finfos = append(finfos, finfo)
			return nil
		})
		require.NoError(t, err)
		require.ElementsEqual(t, []string{"/dir1/", "/dir2/"}, finfosToPaths(finfos))

		return nil
	}, config))
}

func TestGlobFileV2(t *testing.T) {
	// TODO: remove once postgres runs in CI
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	config := newPachdConfig()
	config.StorageV2 = true
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(repo))
		commit1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		fsSpec := fileSetSpec{}
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir1/file1.1", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir1/file1.2", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir2/file2.1", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir2/file2.2", []byte{})))
		err = env.PachClient.PutTarV2(repo, commit1.ID, fsSpec.makeTarStream())
		require.NoError(t, err)
		err = env.PachClient.FinishCommit(repo, commit1.ID)
		require.NoError(t, err)
		globFile := func(x string) []string {
			ys := []string{}
			client, err := env.PachClient.GlobFileV2(env.PachClient.Ctx(), &pfs.GlobFileRequest{
				Commit:  commit1,
				Pattern: x,
			})
			require.NoError(t, err)
			for finfo, err := client.Recv(); err != io.EOF; finfo, err = client.Recv() {
				require.NoError(t, err)
				ys = append(ys, finfo.File.Path)
			}
			return ys
		}
		assert.ElementsMatch(t, []string{"/dir1/file1.2", "/dir2/file2.2"}, globFile("**.2"))
		assert.ElementsMatch(t, []string{"/dir1/file1.1", "/dir1/file1.2"}, globFile("/dir1/*"))
		assert.ElementsMatch(t, []string{"/", "/dir1/", "/dir2/"}, globFile("/*"))
		return nil
	}, config)
	require.NoError(t, err)
}

func TestWalkFileV2(t *testing.T) {
	// TODO: remove once postgres runs in CI
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	config := newPachdConfig()
	config.StorageV2 = true
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(repo))
		commit1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		fsSpec := fileSetSpec{}
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir1/file1.1", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir1/file1.2", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir2/file2.1", []byte{})))
		require.NoError(t, fsSpec.recordFile(tarutil.NewMemFile("/dir2/file2.2", []byte{})))
		err = env.PachClient.PutTarV2(repo, commit1.ID, fsSpec.makeTarStream())
		require.NoError(t, err)
		err = env.PachClient.FinishCommit(repo, commit1.ID)
		require.NoError(t, err)
		walkFile := func(x string) []string {
			ys := []string{}
			client, err := env.PachClient.WalkFileV2(env.PachClient.Ctx(), &pfs.WalkFileRequest{
				File: &pfs.File{
					Commit: commit1,
					Path:   x,
				},
			})
			require.NoError(t, err)
			for finfo, err := client.Recv(); err != io.EOF; finfo, err = client.Recv() {
				require.NoError(t, err)
				ys = append(ys, finfo.File.Path)
			}
			return ys
		}
		assert.ElementsMatch(t, []string{"/dir1/", "/dir1/file1.1", "/dir1/file1.2"}, walkFile("/dir1"))
		assert.ElementsMatch(t, []string{"/dir1/file1.1"}, walkFile("/dir1/file1.1"))
		assert.Len(t, walkFile("/"), 7)
		return nil
	}, config)
	require.NoError(t, err)
}

func TestCompaction(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	config := &serviceenv.PachdFullConfiguration{}
	config.StorageV2 = true
	config.StorageCompactionMaxFanIn = 10
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(repo))
		commit1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)

		const (
			nFileSets   = 100
			filesPer    = 10
			fileSetSize = 1e3
		)
		for i := 0; i < nFileSets; i++ {
			fsSpec := fileSetSpec{}
			for j := 0; j < filesPer; j++ {
				name := fmt.Sprintf("file%02d", j)
				data, err := ioutil.ReadAll(randomReader(fileSetSize))
				if err != nil {
					return err
				}
				file := tarutil.NewMemFile(name, data)
				hdr, err := file.Header()
				if err != nil {
					return err
				}
				fsSpec[hdr.Name] = file
			}
			if err := env.PachClient.PutTarV2(repo, commit1.ID, fsSpec.makeTarStream()); err != nil {
				return err
			}
			runtime.GC()
		}
		if err := env.PachClient.FinishCommit(repo, commit1.ID); err != nil {
			return err
		}
		return nil
	}, config)
	t.Log(err)
	require.NoError(t, err)
}

var (
	randSeed = int64(0)
	randMu   sync.Mutex
)

func getRand() *rand.Rand {
	randMu.Lock()
	seed := randSeed
	randSeed++
	randMu.Unlock()
	return rand.New(rand.NewSource(seed))
}

func randomReader(n int) io.Reader {
	return io.LimitReader(getRand(), int64(n))
}

func TestDiffFileV2(t *testing.T) {
	// TODO: remove once postgres runs in CI
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	config := newPachdConfig()
	config.StorageV2 = true
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		if testing.Short() {
			t.Skip("Skipping integration tests in short mode")
		}
		repo := tu.UniqueString("TestDiff")
		require.NoError(t, env.PachClient.CreateRepo(repo))

		putFile := func(repo, commit, fileName string, data []byte) {
			fsspec := fileSetSpec{
				fileName: tarutil.NewMemFile(fileName, data),
			}
			err := env.PachClient.PutTarV2(repo, commit, fsspec.makeTarStream())
			require.NoError(t, err)
		}

		diffFile := func(newRepo, newCommit, newPath, oldRepo, oldCommit, oldPath string, shallow bool) (newFiles, oldFiles []*pfs.FileInfo) {
			newFile := &pfs.File{
				Commit: &pfs.Commit{ID: newCommit, Repo: &pfs.Repo{Name: newRepo}},
				Path:   newPath,
			}
			var oldFile *pfs.File
			if oldRepo != "" {
				oldFile = &pfs.File{
					Commit: &pfs.Commit{ID: oldCommit, Repo: &pfs.Repo{Name: oldRepo}},
					Path:   oldPath,
				}
			}
			client, err := env.PachClient.DiffFileV2(env.Context, &pfs.DiffFileRequest{
				NewFile: newFile,
				OldFile: oldFile,
			})
			defer client.CloseSend()
			require.NoError(t, err)
			for res, err := client.Recv(); err != io.EOF; res, err = client.Recv() {
				require.NoError(t, err)
				if res.NewFile != nil {
					newFiles = append(newFiles, res.NewFile)
				}
				if res.OldFile != nil {
					oldFiles = append(oldFiles, res.OldFile)
				}
			}
			return newFiles, oldFiles
		}

		// Write foo
		c1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		putFile(repo, c1.ID, "foo", []byte("foo\n"))
		require.NoError(t, env.PachClient.FinishCommit(repo, c1.ID))

		newFiles, oldFiles := diffFile(repo, c1.ID, "", "", "", "", false)
		require.Equal(t, 2, len(newFiles))
		require.Equal(t, "/foo", newFiles[1].File.Path)
		require.Equal(t, 0, len(oldFiles))

		// Change the value of foo
		c2, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFilesV2(repo, c2.ID, []string{"/foo"}))
		putFile(repo, c2.ID, "foo", []byte("not foo\n"))
		require.NoError(t, err)
		require.NoError(t, env.PachClient.FinishCommit(repo, c2.ID))

		newFiles, oldFiles = diffFile(repo, c2.ID, "", "", "", "", false)
		require.Equal(t, 2, len(newFiles))
		require.Equal(t, "/foo", newFiles[1].File.Path)
		require.Equal(t, 2, len(oldFiles))
		require.Equal(t, "/foo", oldFiles[1].File.Path)

		// Write bar
		c3, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		putFile(repo, c3.ID, "/bar", []byte("bar\n"))
		require.NoError(t, err)
		require.NoError(t, env.PachClient.FinishCommit(repo, c3.ID))

		newFiles, oldFiles = diffFile(repo, c3.ID, "", "", "", "", false)
		require.Equal(t, 2, len(newFiles))
		require.Equal(t, "/bar", newFiles[1].File.Path)
		require.Equal(t, 1, len(oldFiles))

		// Delete bar
		c4, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFilesV2(repo, c4.ID, []string{"/bar"}))
		require.NoError(t, env.PachClient.FinishCommit(repo, c4.ID))

		newFiles, oldFiles = diffFile(repo, c4.ID, "", "", "", "", false)
		require.Equal(t, 1, len(newFiles))
		require.Equal(t, 2, len(oldFiles))
		require.Equal(t, "/bar", oldFiles[1].File.Path)

		// Write dir/fizz and dir/buzz
		c5, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		putFile(repo, c5.ID, "/dir/fizz", []byte("fizz\n"))
		putFile(repo, c5.ID, "/dir/buzz", []byte("buzz\n"))
		require.NoError(t, env.PachClient.FinishCommit(repo, c5.ID))

		newFiles, oldFiles = diffFile(repo, c5.ID, "", "", "", "", false)
		require.Equal(t, 4, len(newFiles))
		require.Equal(t, 1, len(oldFiles))

		// Modify dir/fizz
		c6, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		putFile(repo, c6.ID, "/dir/fizz", []byte("fizz\n"))
		require.NoError(t, err)
		require.NoError(t, env.PachClient.FinishCommit(repo, c6.ID))

		newFiles, oldFiles = diffFile(repo, c6.ID, "", "", "", "", false)
		require.Equal(t, 3, len(newFiles))
		require.Equal(t, "/dir/fizz", newFiles[2].File.Path)
		require.Equal(t, 3, len(oldFiles))
		require.Equal(t, "/dir/fizz", oldFiles[2].File.Path)

		return nil
	}, config)
	require.NoError(t, err)
}

func TestInspectFileV2(t *testing.T) {
	// TODO: remove once postgres runs in CI
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	config := newPachdConfig()
	config.StorageV2 = true
	require.NoError(t, testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		ctx := env.Context
		putFile := func(repo, commit, path string, data []byte) error {
			fsSpec := fileSetSpec{}
			fsSpec.recordFile(tarutil.NewMemFile(path, data))
			return env.PachClient.PutTarV2(repo, commit, fsSpec.makeTarStream())
		}
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(repo))

		fileContent1 := "foo\n"
		commit1, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		err = putFile(repo, commit1.ID, "foo/bar", []byte(fileContent1))
		require.NoError(t, err)
		// TODO: can't read uncommitted filesets yet.
		// fileInfo, err := env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
		// 	File: &pfs.File{
		// 		Commit: commit1,
		// 		Path:   "foo",
		// 	},
		// })
		// require.NoError(t, err)
		// require.NotNil(t, fileInfo)

		require.NoError(t, env.PachClient.FinishCommit(repo, commit1.ID))

		fileInfo, err := env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
			File: &pfs.File{
				Commit: commit1,
				Path:   "foo/bar",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, fileInfo)

		fileContent2 := "barbar\n"
		commit2, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		err = putFile(repo, commit2.ID, "foo", []byte(fileContent2))
		require.NoError(t, err)

		// TODO: can't read uncommitted filesets yet.
		// fileInfo, err = env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
		// 	File: &pfs.File{
		// 		Commit: commit2,
		// 		Path:   "foo",
		// 	},
		// })
		// require.NoError(t, err)
		// require.NotNil(t, fileInfo)

		require.NoError(t, env.PachClient.FinishCommit(repo, commit2.ID))

		fileInfo, err = env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
			File: &pfs.File{
				Commit: commit2,
				Path:   "foo",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, fileInfo)

		fileContent3 := "bar\n"
		commit3, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		err = putFile(repo, commit3.ID, "bar", []byte(fileContent3))
		require.NoError(t, err)
		require.NoError(t, env.PachClient.FinishCommit(repo, commit3.ID))
		fileInfo, err = env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
			File: &pfs.File{
				Commit: commit3,
				Path:   "bar",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, fileInfo)
		return nil
	}, config))
}

func TestCopyFile2(t *testing.T) {
	// TODO: remove once postgres runs in CI
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	conf := newPachdConfig()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		ctx := env.Context
		putFile := func(repo, commit, path string, data []byte) error {
			fsspec := fileSetSpec{
				path: tarutil.NewMemFile(path, data),
			}
			return env.PachClient.PutTarV2(repo, commit, fsspec.makeTarStream())
		}
		repo := tu.UniqueString("TestCopyFile")
		require.NoError(t, env.PachClient.CreateRepo(repo))

		masterCommit, err := env.PachClient.StartCommit(repo, "master")
		require.NoError(t, err)
		numFiles := 5
		for i := 0; i < numFiles; i++ {
			err = putFile(repo, masterCommit.ID, fmt.Sprintf("files/%d", i), []byte(fmt.Sprintf("foo %d\n", i)))
			require.NoError(t, err)
		}
		require.NoError(t, env.PachClient.FinishCommit(repo, masterCommit.ID))

		for i := 0; i < numFiles; i++ {
			_, err = env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
				File: &pfs.File{
					Commit: masterCommit,
					Path:   fmt.Sprintf("files/%d", i),
				},
			})
			require.NoError(t, err)
		}

		otherCommit, err := env.PachClient.StartCommit(repo, "other")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.CopyFile(repo, masterCommit.ID, "files", repo, otherCommit.ID, "files", false))
		require.NoError(t, env.PachClient.CopyFile(repo, masterCommit.ID, "files/0", repo, otherCommit.ID, "file0", false))
		require.NoError(t, env.PachClient.FinishCommit(repo, otherCommit.ID))

		for i := 0; i < numFiles; i++ {
			_, err = env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
				File: &pfs.File{
					Commit: otherCommit,
					Path:   fmt.Sprintf("files/%d", i),
				},
			})
			require.NoError(t, err)
		}
		_, err = env.PachClient.InspectFileV2(ctx, &pfs.InspectFileRequest{
			File: &pfs.File{
				Commit: otherCommit,
				Path:   "file0",
			},
		})
		require.NoError(t, err)
		return nil
	}, conf)
	require.NoError(t, err)
}

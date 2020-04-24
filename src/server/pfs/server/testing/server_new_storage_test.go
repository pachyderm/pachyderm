package testing

import (
	"archive/tar"
	"bytes"
	"io"
	"math/rand"
	"sort"
	"testing"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"golang.org/x/sync/errgroup"
)

//func TestCompaction(t *testing.T) {
//	config := &serviceenv.PachdFullConfiguration{}
//	config.NewStorageLayer = true
//	config.StorageMemoryThreshold = units.GB
//	config.StorageShardThreshold = units.GB
//	config.StorageLevelZeroSize = units.GB
//	config.StorageGCPolling = "30s"
//	testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
//		c := env.PachClient
//		repo := "test"
//		require.NoError(t, c.CreateRepo(repo))
//		var eg errgroup.Group
//		for i := 0; i < 1; i++ {
//			branch := "test" + strconv.Itoa(i)
//			eg.Go(func() error {
//				var commit *pfs.Commit
//				var testFiles []string
//				t.Run("PutTar", func(t *testing.T) {
//					var err error
//					for i := 0; i < 10; i++ {
//						commit, err = c.StartCommit(repo, branch)
//						require.NoError(t, err)
//						buf := &bytes.Buffer{}
//						tw := tar.NewWriter(buf)
//						// Create files.
//						for j := 0; j < 10; j++ {
//							testFile := strconv.Itoa(i*10 + j)
//							writeTestFile(t, tw, testFile)
//							testFiles = append(testFiles, testFile)
//						}
//						require.NoError(t, tw.Close())
//						require.NoError(t, c.PutTar(repo, commit.ID, buf))
//						require.NoError(t, c.FinishCommit(repo, commit.ID))
//					}
//				})
//				getTarContent := func(r io.Reader) string {
//					tr := tar.NewReader(r)
//					_, err := tr.Next()
//					require.NoError(t, err)
//					buf := &bytes.Buffer{}
//					_, err = io.Copy(buf, tr)
//					require.NoError(t, err)
//					return buf.String()
//				}
//				t.Run("GetTar", func(t *testing.T) {
//					r, err := c.GetTar(repo, commit.ID, "/0")
//					require.NoError(t, err)
//					require.Equal(t, "0", getTarContent(r))
//					r, err = c.GetTar(repo, commit.ID, "/50")
//					require.NoError(t, err)
//					require.Equal(t, "50", getTarContent(r))
//					r, err = c.GetTar(repo, commit.ID, "/99")
//					require.NoError(t, err)
//					require.Equal(t, "99", getTarContent(r))
//				})
//				t.Run("GetTarConditional", func(t *testing.T) {
//					downloadProb := 0.25
//					require.NoError(t, c.GetTarConditional(repo, commit.ID, "/*", func(fileInfo *pfs.FileInfoNewStorage, r io.Reader) error {
//						if rand.Float64() < downloadProb {
//							require.Equal(t, strings.TrimPrefix(fileInfo.File.Path, "/"), getTarContent(r))
//						}
//						return nil
//					}))
//				})
//				t.Run("GetTarConditionalDirectory", func(t *testing.T) {
//					fullBuf := &bytes.Buffer{}
//					fullTw := tar.NewWriter(fullBuf)
//					rootHdr := &tar.Header{
//						Typeflag: tar.TypeDir,
//						Name:     "/",
//					}
//					require.NoError(t, fullTw.WriteHeader(rootHdr))
//					sort.Strings(testFiles)
//					for _, testFile := range testFiles {
//						writeTestFile(t, fullTw, testFile)
//					}
//					require.NoError(t, fullTw.Close())
//					require.NoError(t, c.GetTarConditional(repo, commit.ID, "/", func(fileInfo *pfs.FileInfoNewStorage, r io.Reader) error {
//						buf := &bytes.Buffer{}
//						_, err := io.Copy(buf, r)
//						require.NoError(t, err)
//						require.Equal(t, 0, bytes.Compare(fullBuf.Bytes(), buf.Bytes()))
//						return nil
//					}))
//				})
//				return nil
//			})
//		}
//		return eg.Wait()
//	}, config)
//}

var (
	shortLoadConfigs = []*loadConfig{
		newLoadConfig(),
	}
	longLoadConfigs = []*loadConfig{}
)

type loadConfig struct {
	pachdConfig *serviceenv.PachdFullConfiguration
	branchGens  []branchGenerator
}

func newLoadConfig(opts ...loadConfigOption) *loadConfig {
	config := &loadConfig{}
	config.pachdConfig = newPachdConfig()
	config.branchGens = append(config.branchGens, newBranchGenerator(withCommitGenerator(newCommitGenerator())))
	for _, opt := range opts {
		opt(config)
	}
	return config
}

type loadConfigOption func(*loadConfig)

func withBranchGenerator(gen branchGenerator) loadConfigOption {
	return func(config *loadConfig) {
		config.branchGens = append(config.branchGens, gen)
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

func withMemoryThreshold(memoryThreshold int64) pachdConfigOption {
	return func(c *serviceenv.PachdFullConfiguration) {
		c.StorageMemoryThreshold = memoryThreshold
	}
}

func withShardThreshold(shardThreshold int64) pachdConfigOption {
	return func(c *serviceenv.PachdFullConfiguration) {
		c.StorageShardThreshold = shardThreshold
	}
}

func withLevelZeroSize(levelZeroSize int64) pachdConfigOption {
	return func(c *serviceenv.PachdFullConfiguration) {
		c.StorageLevelZeroSize = levelZeroSize
	}
}

func withGCPolling(polling string) pachdConfigOption {
	return func(c *serviceenv.PachdFullConfiguration) {
		c.StorageGCPolling = polling
	}
}

type branchGenerator func(*client.APIClient, string) error

type branchConfig struct {
	name       string
	commitGens []commitGenerator
	validator  *validator
}

func newBranchGenerator(opts ...branchGeneratorOption) branchGenerator {
	return func(c *client.APIClient, repo string) error {
		config := &branchConfig{
			name:      "master",
			validator: newValidator(),
		}
		for _, opt := range opts {
			opt(config)
		}
		for _, gen := range config.commitGens {
			if err := gen(c, repo, config.name, config.validator); err != nil {
				return err
			}
		}
		return nil
	}
}

type branchGeneratorOption func(config *branchConfig)

func withCommitGenerator(gen commitGenerator) branchGeneratorOption {
	return func(config *branchConfig) {
		config.commitGens = append(config.commitGens, gen)
	}
}

type commitGenerator func(*client.APIClient, string, string, *validator) error

type commitConfig struct {
	putTarGens []putTarGenerator
}

func newCommitGenerator(opts ...commitGeneratorOption) commitGenerator {
	return func(c *client.APIClient, repo, branch string, validator *validator) error {
		config := &commitConfig{
			putTarGens: []putTarGenerator{newPutTarGenerator(1, 3, 10*units.MB, 50*units.MB)},
		}
		for _, opt := range opts {
			opt(config)
		}
		commit, err := c.StartCommit(repo, branch)
		if err != nil {
			return err
		}
		for _, gen := range config.putTarGens {
			r, err := gen(validator)
			if err != nil {
				return err
			}
			if err := c.PutTar(repo, commit.ID, r); err != nil {
				return err
			}
		}
		if err := c.FinishCommit(repo, commit.ID); err != nil {
			return err
		}
		r, err := c.GetTar(repo, commit.ID, "/")
		if err != nil {
			return err
		}
		return validator.validate(r)
	}
}

type commitGeneratorOption func(config *commitConfig)

func withPutTarGenerator(gen putTarGenerator) commitGeneratorOption {
	return func(config *commitConfig) {
		config.putTarGens = append(config.putTarGens, gen)
	}
}

type putTarGenerator func(*validator) (io.Reader, error)

func newPutTarGenerator(minFiles, maxFiles, minSize, maxSize int) putTarGenerator {
	return func(validator *validator) (io.Reader, error) {
		numFiles := minFiles + rand.Intn(maxFiles-minFiles)
		putTarBuf := &bytes.Buffer{}
		fileBuf := &bytes.Buffer{}
		tw := tar.NewWriter(io.MultiWriter(putTarBuf, fileBuf))
		for i := 0; i < numFiles; i++ {
			name := uuid.NewWithoutDashes()
			// Write serialized tar entry.
			writeFile(tw, name, chunk.RandSeq(minSize+rand.Intn(maxSize-minSize)))
			// Record serialized tar entry for validation.
			validator.recordFile(name, fileBuf)
			fileBuf.Reset()
		}
		if err := tw.Close(); err != nil {
			return nil, err
		}
		return putTarBuf, nil
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

type validator struct {
	files      map[string][]byte
	sampleProb float64
}

func newValidator() *validator {
	return &validator{
		files:      make(map[string][]byte),
		sampleProb: 1.0,
	}
}

func (v *validator) recordFile(name string, r io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}
	v.files[name] = buf.Bytes()
	return nil
}

func (v *validator) validate(r io.Reader) error {
	hdr, err := tar.NewReader(r).Next()
	if err != nil {
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

func TestLoadShort(t *testing.T) {
	for _, loadConfig := range shortLoadConfigs {
		require.NoError(t, testLoad(loadConfig))
	}
}

func TestLoadLong(t *testing.T) {
	t.Skip("Skipping long load tests")
	for _, loadConfig := range longLoadConfigs {
		require.NoError(t, testLoad(loadConfig))
	}
}

func testLoad(loadConfig *loadConfig) error {
	return testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		c := env.PachClient
		repo := "test"
		if err := c.CreateRepo(repo); err != nil {
			return err
		}
		var eg errgroup.Group
		for _, branchGen := range loadConfig.branchGens {
			// (bryce) need a ctx here.
			eg.Go(func() error {
				return branchGen(c, repo)
			})
		}
		return eg.Wait()
	}, loadConfig.pachdConfig)
}

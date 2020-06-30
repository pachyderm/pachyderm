package datum

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	pfstesting "github.com/pachyderm/pachyderm/src/server/pfs/server/testing"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

type hasher struct{}

func (h *hasher) Hash(inputs []*common.InputV2) string {
	return common.HashDatumV2("", "", inputs)
}

func TestSet(t *testing.T) {
	config := pfstesting.NewPachdConfig()
	require.NoError(t, testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		c := env.PachClient
		inputRepo := tu.UniqueString(t.Name() + "_input")
		require.NoError(t, c.CreateRepo(inputRepo))
		outputRepo := tu.UniqueString(t.Name() + "_output")
		require.NoError(t, c.CreateRepo(outputRepo))
		inputCommit, err := c.StartCommit(inputRepo, "master")
		require.NoError(t, err)
		buf := &bytes.Buffer{}
		tw := tar.NewWriter(buf)
		for i := 0; i < 50; i++ {
			require.NoError(t, writeFile(tw, &testFile{
				name: fmt.Sprintf("/foo%v", i),
				data: []byte("input"),
			}))
		}
		require.NoError(t, c.PutTarV2(inputRepo, inputCommit.ID, buf))
		require.NoError(t, c.FinishCommit(inputRepo, inputCommit.ID))
		inputName := "test"
		in := client.NewPFSInput(inputRepo, "/foo*")
		in.Pfs.Name = inputName
		in.Pfs.Commit = inputCommit.ID
		outputCommit, err := c.StartCommit(outputRepo, "master")
		require.NoError(t, err)
		var allInputs [][]*common.InputV2
		// Create datum fileset.
		require.NoError(t, c.WithFileOperationClientV2(outputRepo, outputCommit.ID, func(foc *client.FileOperationClient) error {
			require.NoError(t, withTmpDir(func(storageRoot string) error {
				require.NoError(t, WithSet(c, storageRoot, &hasher{}, func(s *Set) error {
					di, err := NewIteratorV2(c, in)
					if err != nil {
						return err
					}
					return di.Iterate(func(inputs []*common.InputV2) error {
						allInputs = append(allInputs, inputs)
						return s.WithDatum(context.Background(), inputs, func(d *Datum) error {
							return copyFile(path.Join(d.PFSStorageRoot(), OutputPrefix), path.Join(d.PFSStorageRoot(), inputName), func(_ []byte) []byte {
								return []byte("output")
							})
						})
					})
				}, WithMetaOutput(foc)))
				return nil
			}))
			return nil
		}))
		require.NoError(t, c.FinishCommit(outputRepo, outputCommit.ID))
		// Check output.
		require.NoError(t, withTmpDir(func(storageRoot string) error {
			require.NoError(t, WithSet(c, storageRoot, &hasher{}, func(s *Set) error {
				fsi := NewFileSetIterator(c, outputRepo, outputCommit.ID)
				return fsi.Iterate(func(meta *Meta) error {
					require.Equal(t, allInputs[0], meta.Inputs)
					allInputs = allInputs[1:]
					return nil
				})
			}))
			return nil
		}))
		return nil
	}, config))
}

// TODO: Should refactor this and fileset tests to have a general purpose tool for creating tar streams.
type testFile struct {
	name string
	data []byte
}

func writeFile(w *tar.Writer, f *testFile) error {
	hdr := &tar.Header{
		Name: f.name,
		Size: int64(len(f.data)),
	}
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := w.Write(f.data)
	return err
}

func copyFile(outputPrefix, inputPath string, cb func([]byte) []byte) (retErr error) {
	buf := &bytes.Buffer{}
	inputF, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := inputF.Close(); retErr == nil {
			retErr = err
		}
	}()
	if _, err := io.Copy(buf, inputF); err != nil {
		return err
	}
	data := cb(buf.Bytes())
	outputName := path.Base(inputPath)
	outputPath := path.Join(outputPrefix, outputName)
	outputF, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := outputF.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = outputF.Write(data)
	return err
}

func withTmpDir(cb func(string) error) (retErr error) {
	storageRoot := path.Join(os.TempDir(), "datum_test")
	if err := os.RemoveAll(storageRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(storageRoot, 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(storageRoot); retErr == nil {
			retErr = err
		}
	}()
	return cb(storageRoot)
}

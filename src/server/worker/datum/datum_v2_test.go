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
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

type hasher struct{}

func (h *hasher) Hash(inputs []*common.InputV2) string {
	// TODO: implement this for skipping testing.
	return ""
	//return common.HashDatum("", "", inputs)
}

func TestSet(t *testing.T) {
	config := &serviceenv.PachdFullConfiguration{}
	config.StorageV2 = true
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
				name: fmt.Sprintf("foo%v", i),
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
		foc, err := c.NewFileOperationClientV2(outputRepo, outputCommit.ID)
		require.NoError(t, err)
		// Create datum fileset.
		require.NoError(t, WithSet(c, "", &hasher{}, func(s *Set) error {
			di, err := NewIteratorV2(c, in)
			if err != nil {
				return err
			}
			return di.Iterate(func(inputs []*common.InputV2) error {
				return s.WithDatum(context.Background(), inputs, func(d *Datum) error {
					return copyFile(path.Join(d.StorageRoot(), OutputPrefix), path.Join(d.StorageRoot(), inputName), func(_ []byte) []byte {
						return []byte("output")
					})
				})
			})
		}, WithMetaOutput(foc)))
		require.NoError(t, c.FinishCommit(outputRepo, outputCommit.ID))
		// Check output.
		require.NoError(t, WithSet(c, "", &hasher{}, func(s *Set) error {
			di := NewFileSetIterator(c, outputRepo, outputCommit.ID)
			return di.Iterate(func(inputs []*common.InputV2) error {
				return s.WithDatum(context.Background(), inputs, func(d *Datum) error {
					return nil
				})
			})
		}))
		return nil
	}))
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

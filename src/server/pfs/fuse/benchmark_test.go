package fuse

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
)

/*

Aim is to use <10 GB disk space per benchmark (I only have 100G free on my dev
machine, and we don't want to fill up the CI runners).

mount-server benchmark scenarios:

* A is like 50 million files, broken up into 50,000 directories and like ~500 GB
  total?
* B has fewer files, but they're several GB each (I think 100GB-1TB total
  dataset as well), so they want to lazily read files

*/

func BenchmarkA(b *testing.B) {
	env := testpachd.NewRealEnv(b, dockertestenv.NewTestDBConfig(b))
	require.NoError(b, env.PachClient.CreateRepo("repo"))
	random.SeedRand(123)
	src := random.String(GB)
	err := env.PachClient.PutFile(client.NewCommit("repo", "master", ""), "file", strings.NewReader(src))
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		withMount(b, env.PachClient, nil, func(mountPoint string) {
			data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
			require.NoError(b, err)
			require.Equal(b, GB, len(data))
			b.SetBytes(GB)
		})
	}
}

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

This suggests a range of benchmarks along a set of axes/spectrum:

* Read/write
* Distribution/metadata intensity
* Random reads etc
* A single 10GB file

These can be handled with fio.

https://fio.readthedocs.io/en/latest/fio_man.html#cmdoption-arg-filename-format

This suggests that we can get fio to generate a directory structure based on:
$jobname
The name of the worker thread or process.
$jobnum
The incremental number of the worker thread or process.
$filenum
The incremental number of the file for that worker thread or process.

Unless we start 50,000 threads, or run fio 50,000 times, I don't see how we can
create 50k directories. Maybe we combine approaches: have 10k runs of fio w/5
threads, so 50k folders, each run writing 10,000 10kb files.

We can also use the approach in
https://www.flamingbytes.com/posts/fio-multi-files/ with : separators to split,
say 100k files over 10 directories, hence needing to do fewer runs.

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

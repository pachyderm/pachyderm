package server

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

const (
	clients          = 8
	objectsPerClient = 50
	objectSize       = 15 * 1024 * 1024
)

func BenchmarkManyObjects(b *testing.B) {
	b.Run("Put", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var eg errgroup.Group
			for i := 0; i < clients; i++ {
				i := i
				c := GetPachClient(b, GetBasicConfig())
				rand := rand.New(rand.NewSource(int64(i)))
				eg.Go(func() error {
					for j := 0; j < objectsPerClient; j++ {
						r := workload.NewReader(rand, objectSize)
						if n == 0 {
							if _, _, err := c.PutObject(r, fmt.Sprintf("%d.%d", i, j)); err != nil {
								return err
							}
						} else {
							if _, _, err := c.PutObject(r); err != nil {
								return err
							}
						}
					}
					return nil
				})
			}
			b.SetBytes(clients * objectsPerClient * objectSize)
			require.NoError(b, eg.Wait())
		}
	})
	b.Run("Get", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var eg errgroup.Group
			for i := 0; i < clients; i++ {
				i := i
				c := GetPachClient(b, GetBasicConfig())
				eg.Go(func() error {
					for j := 0; j < objectsPerClient; j++ {
						err := c.GetTag(fmt.Sprintf("%d.%d", i, j), ioutil.Discard)
						if err != nil {
							return err
						}
					}
					return nil
				})
			}
			b.SetBytes(clients * objectsPerClient * objectSize)
			require.NoError(b, eg.Wait())
		}
	})
	b.Run("CacheGet", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var eg errgroup.Group
			for i := 0; i < clients; i++ {
				i := i
				c := GetPachClient(b, GetBasicConfig())
				eg.Go(func() error {
					for j := 0; j < objectsPerClient; j++ {
						err := c.GetTag(fmt.Sprintf("%d.%d", i, j), ioutil.Discard)
						if err != nil {
							return err
						}
					}
					return nil
				})
			}
			b.SetBytes(clients * objectsPerClient * objectSize)
			require.NoError(b, eg.Wait())
		}
	})
}

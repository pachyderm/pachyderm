//go:build livek8s
// +build livek8s

package server

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/local"
)

func init() {
	if v := os.Getenv("LOCAL_TEST"); v != "" {
		go func() {
			if err := local.RunLocal(); err != nil {
				panic(err)
			}
		}()
	}
}

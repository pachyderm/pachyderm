package fileset

import (
	"testing"

	_ "github.com/lib/pq"
)

func TestPGStore(t *testing.T) {
	PathStoreTestSuite(t, func(cb func(PathStore)) {
		WithTestPathStore(t, cb)
	})
}

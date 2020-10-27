package fileset

import "testing"

func TestPGStore(t *testing.T) {
	StoreTestSuite(t, func(cb func(s Store)) {
		WithTestStore(t, cb)
	})
}

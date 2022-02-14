package obj

import "testing"

func TestCacheClient(t *testing.T) {
	t.Parallel()
	TestSuite(t, func(t testing.TB) Client {
		fast := newTestLocalClient(t)
		slow := newTestLocalClient(t)
		return NewCacheClient(slow, fast, 3)
	})
}

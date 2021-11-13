package obj

import "testing"

func TestLimitedClient(t *testing.T) {
	t.Parallel()
	TestSuite(t, func(t testing.TB) Client {
		c := newTestLocalClient(t)
		return NewLimitedClient(c, 2, 2)
	})
}

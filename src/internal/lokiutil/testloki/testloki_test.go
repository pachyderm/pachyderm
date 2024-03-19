package testloki

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestStartLoki(t *testing.T) {
	ctx := pctx.TestContext(t)
	loki, err := New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("set up loki: %v", err)
	}
	t.Cleanup(func() {
		if err := loki.Close(); err != nil {
			t.Fatalf("clean up loki: %v", err)
		}
	})
}

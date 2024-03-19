package testing

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
)

func TestTestPachd(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	_, err := c.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "not implemented"; !strings.Contains(got, want) {
		t.Errorf("unexpected error text:\n  got: %v\n want: contains(%q)", got, want)
	}
}

func TestRealEnv(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	_, err := env.PachClient.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "not implemented"; !strings.Contains(got, want) {
		t.Errorf("unexpected error text:\n  got: %v\n want: contains(%q)", got, want)
	}
}

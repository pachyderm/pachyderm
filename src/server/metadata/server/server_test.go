package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
)

func TestEditMetadata(t *testing.T) {
	ctx := pctx.TestContext(t)
	server := &APIServer{}
	_, err := server.EditMetadata(ctx, &metadata.EditMetadataRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

package pctx

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
)

func TestBackground(t *testing.T) {
	_, h := log.TestWithCapture(t)
	log.Info(Background(""), "hi")
	h.HasALog(t)
}

func TestTODO(t *testing.T) {
	_, h := log.TestWithCapture(t)
	log.Info(TODO(), "hi")
	h.HasALog(t)
}

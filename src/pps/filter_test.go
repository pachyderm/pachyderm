package pps_test

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestDatumStateFilter(t *testing.T) {
	var (
		ctx = context.Background()
		f   = pps.ListDatumRequest_Filter{State: []pps.DatumState{pps.DatumState_FAILED}}
		d   = &pps.DatumInfo{State: pps.DatumState_UNKNOWN}
	)
	if f.Allow(d) {
		t.Errorf("%v allowed %v", f, d.State)
	}
	d.State = pps.DatumState_FAILED
	if !f.Allow(d) {
		t.Errorf("%v disallowed matching state", f)
	}
}

package pps_test

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestDatumStateFilter(t *testing.T) {
	var (
		ctx = context.Background()
		f   = pps.NewDatumStateFilter(pps.DatumState_FAILED)
		d   = pps.DatumInfo{State: pps.DatumState_UNKNOWN}
	)
	if f.Allow(ctx, d) {
		t.Errorf("%v allowed %v", f, d.State)
	}
	d.State = pps.DatumState_FAILED
	if !f.Allow(ctx, d) {
		t.Errorf("%v disallowed matching state", f)
	}
}

func TestNot(t *testing.T) {
	var (
		ctx = context.Background()
		f   = pps.NewNotFilter(pps.NewDatumStateFilter(pps.DatumState_SUCCESS))
	)
	if f.Allow(ctx, &pps.DatumInfo{State: pps.DatumState_SUCCESS}) {
		t.Errorf("%v allowed %v", f, pps.DatumState_SUCCESS)
	}
	if !f.Allow(ctx, &pps.DatumInfo{}) {
		t.Errorf("%v disallowed %v", f, pps.DatumState(0))
	}
}

func TestAnd(t *testing.T) {
	var (
		ctx = context.Background()
		f   = pps.NewAndFilter(
			pps.NewDatumStateFilter(pps.DatumState_FAILED),
			pps.NewDatumStateFilter(pps.DatumState_SKIPPED),
		)
	)
	for _, s := range []pps.DatumState{pps.DatumState_UNKNOWN, pps.DatumState_FAILED, pps.DatumState_SKIPPED} {
		if f.Allow(ctx, &pps.DatumInfo{State: s}) {
			t.Errorf("%v allowed %v", f, s)
		}
	}
	f = pps.NewAndFilter(
		pps.NewOrFilter(
			pps.NewDatumStateFilter(pps.DatumState_FAILED),
			pps.NewDatumStateFilter(pps.DatumState_STARTING),
		),
		pps.NewOrFilter(
			pps.NewDatumStateFilter(pps.DatumState_SUCCESS),
			pps.NewDatumStateFilter(pps.DatumState_FAILED),
		),
	)
	if f.Allow(ctx, &pps.DatumInfo{}) {
		t.Errorf("%v allowed %v", f, pps.DatumState(0))
	}
	if !f.Allow(ctx, &pps.DatumInfo{State: pps.DatumState_FAILED}) {
		t.Errorf("%v disallowed %v", f, pps.DatumState_FAILED)
	}
}

func TestOr(t *testing.T) {
	var (
		ctx = context.Background()
		f   = pps.NewOrFilter(
			pps.NewDatumStateFilter(pps.DatumState_FAILED),
			pps.NewDatumStateFilter(pps.DatumState_SKIPPED),
		)
	)
	if f.Allow(ctx, &pps.DatumInfo{}) {
		t.Errorf("%v allowed %v", f, pps.DatumState(0))
	}
	for _, s := range []pps.DatumState{pps.DatumState_FAILED, pps.DatumState_SKIPPED} {
		if !f.Allow(ctx, &pps.DatumInfo{State: s}) {
			t.Errorf("%v disallowed %v", f, s)
		}
	}
}

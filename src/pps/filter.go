package pps

import (
	context "context"
)

func (r *ListDatumRequest_Filter) Allow(ctx context.Context, item *DatumInfo) bool {
	if r == nil {
		return true
	}
	match := false
	for _, s := range r.State {
		if s == item.State {
			match = true
		}
	}
	if len(r.State) > 0 && !match {
		return false
	}

	return true
}

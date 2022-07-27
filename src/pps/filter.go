package pps

// Allow returns true if the filter allows the item.  Currently, this means if
// the item’s state matches one of the states given in the filter.
func (r *ListDatumRequest_Filter) Allow(item *DatumInfo) bool {
	// A missing filter allows all items.
	if r == nil {
		return true
	}
	// An empty filter allows all items.
	if len(r.State) == 0 {
		return true
	}
	for _, s := range r.State {
		if s == item.State {
			return true
		}
	}
	return false
}

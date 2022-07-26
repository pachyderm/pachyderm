package pps

import (
	context "context"
)

type Allower interface {
	// Allow returns true if the filter allows the item.
	Allow(ctx context.Context, item any) bool
}

// Allow returns true if the filter allows the item, false otherwise.  A nil
// filter is always true.
func (f *Filter) Allow(ctx context.Context, item any) bool {
	if f == nil {
		return true
	}
	// We have to implement polymorphism manually this way because we can’t
	// add methods to the isFilter_Filter type used by protoc.
	switch f := f.Filter.(type) {
	case *Filter_NotFilter:
		return f.NotFilter.Allow(ctx, item)
	case *Filter_AndFilter:
		return f.AndFilter.Allow(ctx, item)
	case *Filter_OrFilter:
		return f.OrFilter.Allow(ctx, item)
	case *Filter_DatumStateFilter:
		return f.DatumStateFilter.Allow(ctx, item)
	}
	return false
}

// NewNotFilter returns a new Filter which inverts its argument filter.
func NewNotFilter(f *Filter) *Filter {
	return &Filter{
		Filter: &Filter_NotFilter{
			NotFilter: &NotFilter{
				Operand: f,
			},
		},
	}
}

// Allow inverts its operand.  A nil NotFilter is always false.
func (f *NotFilter) Allow(ctx context.Context, item any) bool {
	if f == nil {
		return false
	}
	return !f.Operand.Allow(ctx, item)
}

// NewAndFilter returns a new Filter which matches if all of its argument
// filters match.
func NewAndFilter(ff ...*Filter) *Filter {
	return &Filter{
		Filter: &Filter_AndFilter{
			AndFilter: &AndFilter{
				Operands: ff,
			},
		},
	}
}

// And returns true if all of its operands are true.  A nil or empty AndFilter
// is always true.
func (f *AndFilter) Allow(ctx context.Context, item any) bool {
	if f == nil {
		return true
	}
	for _, f := range f.Operands {
		if !f.Allow(ctx, item) {
			return false
		}
	}
	return true
}

// NewOrFilter returns a new Filter which matches if any of the argument filters
// match.
func NewOrFilter(ff ...*Filter) *Filter {
	return &Filter{
		Filter: &Filter_OrFilter{
			OrFilter: &OrFilter{
				Operands: ff,
			},
		},
	}
}

// Allow returns true if any of its operands is true.  It returns false if given
// no operands.
func (f *OrFilter) Allow(ctx context.Context, item any) bool {
	if f == nil {
		return false
	}
	for _, f := range f.Operands {
		if f.Allow(ctx, item) {
			return true
		}
	}
	return false
}

// NewDatumStateFilter returns a new Filter which matches the given datum state.
func NewDatumStateFilter(s DatumState) *Filter {
	return &Filter{
		Filter: &Filter_DatumStateFilter{
			DatumStateFilter: &DatumStateFilter{
				Value: s,
			},
		},
	}
}

// Allow returns true if its argument is a DatumInfo with a State field matching
// the filter’s Value field.  A nil filter is always false.
func (f *DatumStateFilter) Allow(ctx context.Context, item any) bool {
	if f == nil {
		return false
	}
	switch item := item.(type) {
	case DatumInfo:
		return item.State == f.Value
	case *DatumInfo:
		return item.State == f.Value
	default:
		return false
	}
}

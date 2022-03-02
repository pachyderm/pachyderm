package sdata

import "fmt"

// ErrTupleFields is returned by writers to indicate that
// a tuple is not the right shape to be written to them.
type ErrTupleFields struct {
	Writer TupleWriter
	Tuple  Tuple
	Fields []string
}

func (e ErrTupleFields) Error() string {
	return fmt.Sprintf("tuple has invalid fields for this writer (%T). expected: (%v), only have %d", e.Writer, e.Fields, (e.Tuple))
}

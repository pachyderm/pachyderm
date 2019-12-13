package stats

type statsErr struct {
	err string
}

func (s statsErr) Error() string {
	return s.err
}

func (s statsErr) String() string {
	return s.err
}

// These are the package-wide error values.
// All error identification should use these values.
var (
	EmptyInputErr = statsErr{"Input must not be empty."}
	NaNErr        = statsErr{"Not a number."}
	NegativeErr   = statsErr{"Must not contain negative values."}
	ZeroErr       = statsErr{"Must not contain zero values."}
	BoundsErr     = statsErr{"Input is outside of range."}
	SizeErr       = statsErr{"Must be the same length."}
	InfValue      = statsErr{"Value is infinite."}
	YCoordErr     = statsErr{"Y Value must be greater than zero."}
)

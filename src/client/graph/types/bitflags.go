package types

type BitFlags interface {
	// represents bitflats as UInt64
	UInt64()
	// returns list of flags that are set
	HighFlags() []string
	// returns list of flags that are unset
	LowFlags() []string
	// sets a flag if it exists, returns err otherwise
	SetFlag(string) error
	// clears a flag or returns error
	ClearFlag(string) error
	// sets flags or returns errors
	SetFlags(...string) []error
	// clears flags or returns error
	ClearFlags(...string) []error
	// tests if flag is set
	TestFlag(string) bool
	// tests flags and returns bool if all flags are set
	TestFlags(...string) bool
	// initializes a flags object from a uint64
	FromUInt64(uint64) error
}

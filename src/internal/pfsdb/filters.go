package pfsdb

type (
	SortOrder        int
	FilterExpression string
)

const (
	SortNone SortOrder = iota
	SortAscend
	SortDescend
)

const (
	ValueIn FilterExpression = "%s in (%s)"
)

type Filter struct {
	Field      string
	Value      any
	Expression FilterExpression
}

type ListResourceConfig struct {
	Limit     int
	Offset    int
	Filters   []*Filter
	SortOrder SortOrder
}

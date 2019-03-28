package fileset

const (
	Index = "index"
)

type Range struct {
	lower, upper int64
}

// GetRange does a range search for a prefix.
// This is only applicable when the index is created in a sorted fashion.
func (i *Index) GetRange(prefix string) *Range {
	return nil
}

func New() *Index {
	return &Index{}
}

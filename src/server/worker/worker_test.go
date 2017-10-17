package worker

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"testing"
)

func TestMergeRanges(t *testing.T) {
	require.Equal(t, []*DatumRange(nil), mergeRanges(nil))
	require.Equal(t, []*DatumRange{}, mergeRanges([]*DatumRange{}))
	require.Equal(t, []*DatumRange{&DatumRange{Upper: 10}}, mergeRanges([]*DatumRange{&DatumRange{Upper: 10}}))
	require.Equal(t, []*DatumRange{
		&DatumRange{Upper: 30},
	}, mergeRanges([]*DatumRange{
		&DatumRange{Lower: 20, Upper: 30},
		&DatumRange{Lower: 10, Upper: 20},
		&DatumRange{Upper: 10},
	}))
	require.Equal(t, []*DatumRange{
		&DatumRange{Upper: 10},
		&DatumRange{Lower: 20, Upper: 30},
	}, mergeRanges([]*DatumRange{
		&DatumRange{Lower: 20, Upper: 30},
		&DatumRange{Upper: 10},
	}))
}

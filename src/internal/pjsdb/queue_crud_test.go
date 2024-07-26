package pjsdb_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestForEachQueue(t *testing.T) {
	ctx, db := DB(t)
	s := FilesetStorage(t, db)
	withTx(t, ctx, db, s, func(d dependencies) {
		prog1 := mockFileset(t, d, "/program", "#!/bin/bash; echo 'hello';")
		prog2 := mockFileset(t, d, "/program", "#!/bin/bash; echo 'hi';")
		prog3 := mockFileset(t, d, "/program", "#!/bin/bash; echo 'hey';")
		for i := 0; i < 25; i++ {
			createJobWithFilesets(t, d, 0, prog1)
			createJobWithFilesets(t, d, 0, prog2)
			createJobWithFilesets(t, d, 0, prog3)
		}
	})
	num := 0
	err := pjsdb.ForEachQueue(ctx, db, pjsdb.IterateQueuesRequest{}, func(queue pjsdb.Queue) error {
		num++
		require.Len(t, queue.Jobs, 25)
		require.Len(t, queue.Programs, 25)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, num)
}

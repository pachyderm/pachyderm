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
		var err error
		targetFs := mockFileset(t, d, "/spec", "#!/bin/bash; echo 'hello';")
		targetFs2 := mockFileset(t, d, "/spec", "#!/bin/bash; echo 'hi';")
		for i := 0; i < 25; i++ {
			createJobWithFilesets(t, d, 0, targetFs, nil)
			createJobWithFilesets(t, d, 0, targetFs2, nil)
			_, err = createJob(t, d, 0)
			require.NoError(t, err)
		}
	})
	num := 0
	err := pjsdb.ForEachQueue(ctx, db, pjsdb.IterateQueuesRequest{}, func(queue pjsdb.Queue) error {
		num++
		require.Len(t, queue.Jobs, 25)
		require.Len(t, queue.Specs, 25)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, num)
}

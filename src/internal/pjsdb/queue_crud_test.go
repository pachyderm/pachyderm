package pjsdb_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

func TestForEachQueue(t *testing.T) {
	ctx, db := DB(t)
	s := FilesetStorage(t, db)
	withTx(t, ctx, db, s, func(d dependencies) {
		prog1, prog1Hash := mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hello';")
		prog2, prog2Hash := mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hi';")
		prog3, prog3Hash := mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hey';")
		for i := 0; i < 25; i++ {
			createJobWithFilesets(t, d, 0, prog1, prog1Hash)
			createJobWithFilesets(t, d, 0, prog2, prog2Hash)
			createJobWithFilesets(t, d, 0, prog3, prog3Hash)
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

func TestDequeue(t *testing.T) {
	ctx, db := DB(t)
	s := FilesetStorage(t, db)
	var prog1, prog2, prog3 fileset.PinnedFileset
	var prog1Hash, prog2Hash, prog3Hash []byte
	// create 3 queues and put in 25 jobs in each queue
	const NumOfPrograms = 25
	withTx(t, ctx, db, s, func(d dependencies) {
		prog1, prog1Hash = mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hello';")
		prog2, prog2Hash = mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hi';")
		prog3, prog3Hash = mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hey';")
		for i := 0; i < NumOfPrograms; i++ {
			createJobWithFilesets(t, d, 0, prog1, prog1Hash)
			createJobWithFilesets(t, d, 0, prog2, prog2Hash)
			createJobWithFilesets(t, d, 0, prog3, prog3Hash)
		}
	})

	var dequeued uint64 = 0
	for i := 0; i < NumOfPrograms; i++ {
		withTx(t, ctx, db, s, func(d dependencies) {
			dequeued++
			// for now the program hash is also the program
			resp, err := pjsdb.DequeueAndProcess(ctx, d.tx, prog1Hash)
			require.NoError(t, err)
			// the job is dequeued in FIFO order
			require.Equal(t, uint64(resp.ID), dequeued)
			dequeued++
			resp, err = pjsdb.DequeueAndProcess(ctx, d.tx, prog2Hash)
			require.NoError(t, err)
			require.Equal(t, uint64(resp.ID), dequeued)
			dequeued++
			resp, err = pjsdb.DequeueAndProcess(ctx, d.tx, prog3Hash)
			require.NoError(t, err)
			require.Equal(t, uint64(resp.ID), dequeued)
		})
		withTx(t, ctx, db, s, func(d dependencies) {
			err := pjsdb.ForEachQueue(ctx, db, pjsdb.IterateQueuesRequest{}, func(queue pjsdb.Queue) error {
				require.Len(t, queue.Jobs, NumOfPrograms-int(dequeued)/3)
				require.Len(t, queue.Programs, NumOfPrograms-int(dequeued)/3)
				return nil
			})
			require.NoError(t, err)
		})
	}
	withTx(t, ctx, db, s, func(d dependencies) {
		// All three queues are empty. An error is expected if more dequeue operation is performed
		_, err := pjsdb.DequeueAndProcess(ctx, d.tx, prog1Hash)
		require.YesError(t, err)
		require.True(t, errors.As(err, &pjsdb.DequeueFromEmptyQueueError{}))
	})
}

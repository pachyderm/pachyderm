package chunk

type GC struct {
	objc obj.Client
	period time.Duration
	tracker chunk.Tracker
}

type IDIterator func(ctx context.Context, func(chunk.ChunkID) error) error

func NewGC(db *sqlx.DB, objc obj.Client, tracker chunk.Tracker, pollingPeriod time.Duration) *GC {
	return &GC{
		objc: objc,
		tracker: tracker,
		period: pollingPeriod,
		db: db,
	}	
}

func (gc *GC) Run(ctx context.Context) error {
	ticker := time.NewTicker(gc.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := gc.runOnce(ctx); err != nil {
				log.Error(err)
			}
		}	
	}
	return nil
}

const deleteableQuery = `
	SELECT hash_id FROM storage.chunks
	WHERE int_id IN (
		SELECT to FROM storage.chunk_refs
		WHERE count(from) = 0
		GROUP BY to
	)
	AND hash_id NOT IN (SELECT chunk_id FROM storage.paths)
`

func (gc *GC) runOnce(ctx context.Context) error {	
	tmpID := fmt.Sprintf("gc-%d", time.Now().UnixNano())
	if err := gc.tracker.NewChunkSet(ctx, ); err != nil {
		return nil, err
	}
	ctx, cf := context.WithCancel(ctx)
	eg, ctx2 := errgroup.WithContext(ctx)
	eg.Go(func() error {
		
	})
	eg.Go(func() error {
		rows, err := gc.db.QueryContext(ctx, deleteableQuery)
		if err != nil {
			return err
		}
		for rows.Next() {
			var chunkID []byte
			if err := rows.Scan(&chunkId); err != nil {
				return err
			}
			if err := chunk.DeleteChunk(ctx, gc.tracker, gc.objc, chunkID); err != nil {
				return err
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		cf()
		return nil
	})
	return eg.Wait()	
}


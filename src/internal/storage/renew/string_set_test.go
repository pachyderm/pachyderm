package renew

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

func TestStringSet(t *testing.T) {
	idsRenewed := make(map[string]struct{})
	rf := func(ctx context.Context, id string, ttl time.Duration) error {
		idsRenewed[id] = struct{}{}
		return nil
	}
	idsCreated := make(map[string]struct{})
	cf := func(ctx context.Context, ids []string, ttl time.Duration) (string, error) {
		id := uuid.NewWithoutDashes()
		idsCreated[id] = struct{}{}
		return id, nil
	}
	require.NoError(t, WithStringSet(context.Background(), 15*time.Second, rf, cf, func(ctx context.Context, ss *StringSet) error {
		time.Sleep(time.Second)
		for i := 0; i < 101; i++ {
			if err := ss.Add(ctx, uuid.NewWithoutDashes()); err != nil {
				return err
			}
		}
		for i := 0; i < 100; i++ {
			id := uuid.NewWithoutDashes()
			if err := ss.Add(ctx, id); err != nil {
				return err
			}
			idsCreated[id] = struct{}{}
		}
		time.Sleep(15 * time.Second)
		return nil
	}))
	require.Equal(t, 101, len(idsCreated))
	require.Equal(t, idsCreated, idsRenewed)
}

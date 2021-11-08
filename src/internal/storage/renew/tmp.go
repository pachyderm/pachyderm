package renew

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// TmpTrackerPrefix is the tracker prefix for temporary objects.
const TmpTrackerPrefix = "tmp/"

type tmpDeleter struct{}

// NewTmpDeleter creates a new temporary deleter.
func NewTmpDeleter() track.Deleter {
	return &tmpDeleter{}
}

func (*tmpDeleter) DeleteTx(tx *pachsql.Tx, _ string) error {
	return nil
}

// NewTmpComposer returns a ComposeFunc which creates meaningless temporary objects
// Use NewTmpDeleter to get a no-op track.Deleter to handle these objects.
func NewTmpComposer(tr track.Tracker, name string) ComposeFunc {
	if name == "" {
		panic("must provide non-empty name for tmp ComposeFunc")
	}
	prefix := fmt.Sprintf("%s/%s-%s", TmpTrackerPrefix, name, uuid.NewWithoutDashes())
	var n int
	var mu sync.Mutex
	return func(ctx context.Context, ids []string, ttl time.Duration) (string, error) {
		mu.Lock()
		n2 := n
		n++
		mu.Unlock()
		tmpID := fmt.Sprintf("%s/%08x", prefix, n2)
		if err := track.Create(ctx, tr, tmpID, ids, ttl); err != nil {
			return "", err
		}
		return tmpID, nil
	}
}

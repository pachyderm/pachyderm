package track

import "context"

// Drop makes an object eligible for deletion, if nothing is referencing it.
// It does this with a negative TTL, but if that strategy ever becomes deprecated
// callers of Drop will benefit from the new strategy.
func Drop(ctx context.Context, track Tracker, id string) error {
	_, err := track.SetTTLPrefix(ctx, id, -1)
	return err
}

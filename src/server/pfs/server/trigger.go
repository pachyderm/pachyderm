package server

import (
	"github.com/docker/go-units"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// isTriggerred checks to see if a branch should be updated from oldHead to
// newHead based on a trigger.
func (d *driver) isTriggerred(t *pfs.Trigger, oldHead, newHead *pfs.CommitInfo) (bool, error) {
	if t.Size_ != "" {
		size, err := units.FromHumanSize(t.Size_)
		if err != nil {
			return false, err
		}
		if int64(newHead.SizeBytes-oldHead.SizeBytes) >= size {
			return true, nil
		}
	}
	return false, nil
}

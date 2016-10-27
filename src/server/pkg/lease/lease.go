package lease

import "time"

type Resource interface {
	ID() string
}

type LeaseManager interface {
	// lease the resource r for duration d.  When the lease expires, invoke `revoke`
	// a lease can be refreshed by calling Lease() again on the same resource
	Lease(r Resource, d time.Duration, revoke func())
	// return the resource r.
	Return(r Resource)
}

type leaseManager struct {
	timers map[string]*time.Timer
}

func NewLeaseManager() LeaseManager {
	return &leaseManager{
		timers: make(map[string]*time.Timer),
	}
}

func (l *leaseManager) Lease(r Resource, d time.Duration, revoke func()) {
	timer := time.AfterFunc(d, revoke)
	// cancel the old timer and add the new one
	if t, ok := c.timers[r.ID()]; ok {
		t.Stop()
	}
	c.timers[r.ID()] = timer
}

func (l *leaseManager) Return(r Resource) {
	if t, ok := c.timers[r.ID()]; ok {
		t.Stop()
	}
}

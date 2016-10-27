package lease

import "time"

type LeaseManager interface {
	// lease the resource r for duration d.  When the lease expires, invoke `revoke`
	// a lease can be refreshed by calling Lease() again on the same resource
	Lease(r string, d time.Duration, revoke func())
	// return the resource r.
	Return(r string)
}

type leaseManager struct {
	timers map[string]*time.Timer
}

func NewLeaseManager() LeaseManager {
	return &leaseManager{
		timers: make(map[string]*time.Timer),
	}
}

func (l *leaseManager) Lease(r string, d time.Duration, revoke func()) {
	timer := time.AfterFunc(d, revoke)
	// cancel the old timer and add the new one
	if t, ok := l.timers[r]; ok {
		t.Stop()
	}
	l.timers[r] = timer
}

func (l *leaseManager) Return(r string) {
	if t, ok := l.timers[r]; ok {
		t.Stop()
	}
}

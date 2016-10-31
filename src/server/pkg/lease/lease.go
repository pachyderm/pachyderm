package lease

import "time"

// Manager manages resources via leases
type Manager interface {
	// lease the resource r for duration d.  When the lease expires, invoke `revoke`
	// a lease can be refreshed by calling Lease() again on the same resource
	Lease(r string, d time.Duration, revoke func())
	// return the resource r.
	Return(r string)
}

type manager struct {
	timers map[string]*time.Timer
}

// NewManager creates a new lease manager
func NewManager() Manager {
	return &manager{
		timers: make(map[string]*time.Timer),
	}
}

func (l *manager) Lease(r string, d time.Duration, revoke func()) {
	timer := time.AfterFunc(d, revoke)
	// cancel the old timer and add the new one
	if t, ok := l.timers[r]; ok {
		t.Stop()
	}
	l.timers[r] = timer
}

func (l *manager) Return(r string) {
	if t, ok := l.timers[r]; ok {
		t.Stop()
	}
}

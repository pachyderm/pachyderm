package lease

import "time"

// Leaser manages resources via leases
type Leaser interface {
	// lease the resource r for duration d.  When the lease expires, invoke `revoke`
	// a lease can be refreshed by calling Lease() again on the same resource
	Lease(r string, d time.Duration, revoke func())
	// return the resource r.
	Return(r string)
}

type leaser struct {
	timers map[string]*time.Timer
}

// NewLeaser creates a new lease leaser
func NewLeaser() Leaser {
	return &leaser{
		timers: make(map[string]*time.Timer),
	}
}

func (l *leaser) Lease(r string, d time.Duration, revoke func()) {
	timer := time.AfterFunc(d, revoke)
	// cancel the old timer and add the new one
	if t, ok := l.timers[r]; ok {
		t.Stop()
	}
	l.timers[r] = timer
}

func (l *leaser) Return(r string) {
	if t, ok := l.timers[r]; ok {
		t.Stop()
	}
}

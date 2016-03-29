package google_type

import (
	"sync"
	"time"
)

var (
	// SystemDater is a Dater that uses the system UTC time.
	SystemDater = &systemDater{}
)

// NewDate is a convienence function to create a new Date.
func NewDate(day int32, month int32, year int32) *Date {
	return &Date{
		Day:   day,
		Month: month,
		Year:  year,
	}
}

// Now returns the current date at UTC.
func Now() *Date {
	return TimeToDate(time.Now().UTC())
}

// TimeToDate converts a golang Time to a Date.
func TimeToDate(t time.Time) *Date {
	return NewDate(int32(t.Day()), int32(t.Month()), int32(t.Year()))
}

// GoTime converts a Date to a golang Time.
func (d *Date) GoTime() time.Time {
	if d == nil {
		return time.Unix(0, 0).UTC()
	}
	return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), 0, 0, 0, 0, time.UTC)
}

// Before returns true if d is before j.
func (d *Date) Before(j *Date) bool {
	if j == nil {
		return false
	}
	if d == nil {
		return true
	}
	if d.Year < j.Year {
		return true
	}
	if d.Year > j.Year {
		return false
	}
	if d.Month < j.Month {
		return true
	}
	if d.Month > j.Month {
		return false
	}
	return d.Day < j.Day
}

// Equal returns true if i equals j.
func (d *Date) Equal(j *Date) bool {
	return ((d == nil) == (j == nil)) && ((d == nil) || (*d == *j))
}

// InRange returns whether d is within start to end, inclusive.
// The given date is expected to not be nil.
// If start is nil, it checks whether d is less than or equal to end.
// If end is nil it checks whether d is greater than or equal to end.
// If start and end are nil, it returns true.
func (d *Date) InRange(start *Date, end *Date) bool {
	if start == nil && end == nil {
		return true
	}
	if start == nil {
		return d.Before(end) || d.Equal(end)
	}
	if end == nil {
		return start.Before(d) || start.Equal(d)
	}
	return d.Equal(start) || d.Equal(end) || (start.Before(d) && d.Before(end))
}

// Dater provides the current date.
type Dater interface {
	Now() *Date
}

// FakeDater is a Dater for testing.
type FakeDater interface {
	Dater
	Set(day int32, month int32, year int32)
}

// NewFakeDater returns a new FakeDater with the initial date.
func NewFakeDater(day int32, month int32, year int32) FakeDater {
	return newFakeDater(day, month, year)
}

type systemDater struct{}

func (s *systemDater) Now() *Date {
	return Now()
}

type fakeDater struct {
	curDate *Date
	lock    *sync.RWMutex
}

func newFakeDater(day int32, month int32, year int32) *fakeDater {
	return &fakeDater{NewDate(day, month, year), &sync.RWMutex{}}
}

func (f *fakeDater) Now() *Date {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return copyDate(f.curDate)
}

func (f *fakeDater) Set(day int32, month int32, year int32) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.curDate = NewDate(day, month, year)
}

func copyDate(date *Date) *Date {
	c := *date
	return &c
}

package disk

import "time"

type DiskUsage struct {
	UsedBytes uint64
	FreeBytes uint64
}

// Monitor monitors disk usage for a given filesystem identified by a path.
// The path only has to reside in the file system; it doesn't literally have to
// be the mount point.
type Monitor interface {
	// Get the current disk usage
	CurrentUsage() DiskUsage
	// Get the path that identifies the FS that the monitor is monitoring
	Path() string
	// Returns a channel that pulses disk usage
	Pulse(period time.Duration) chan DiskUsage
}

type monitor struct {
	path string
}

func NewMonitor(path string) Monitor {
	return &monitor{
		path: path,
	}
}

func (m *monitor) CurrentUsage() DiskUsage {

}

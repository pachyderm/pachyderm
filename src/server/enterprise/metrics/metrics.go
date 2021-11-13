package metrics

import "sync/atomic"

var (
	enterpriseFailures int64
)

// IncEnterpriseFailures records that a command failed due to enterprise not being activated.
func IncEnterpriseFailures() {
	atomic.AddInt64(&enterpriseFailures, 1)
}

// GetEnterpriseFailures returns a count of how many enterprise failures there have been.
func GetEnterpriseFailures() int64 {
	return atomic.LoadInt64(&enterpriseFailures)
}

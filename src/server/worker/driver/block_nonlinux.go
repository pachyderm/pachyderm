//go:build !linux

package driver

// On OSes other than Linux, waitid(2) is too broken to use, or not implemented at all.
func blockUntilWaitable(pid int) (bool, error) {
	return false, nil
}

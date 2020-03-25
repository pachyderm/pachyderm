package assets

import "errors"

// This is an empty placeholder file. At build time, go-bindata will package
// cert data from /etc/ssl/certs here.

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	// don't panic here because the linter is too damn smart, and figures out
	// that downstream functions are unused when a panic occurs here
	return errors.New("RestoreAssets is not implemented because it is only populated at build-time")
}

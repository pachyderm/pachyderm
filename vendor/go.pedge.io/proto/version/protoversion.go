package protoversion

import "fmt"

// NewAPIServer creates a new APIServer for the given Version.
func NewAPIServer(version *Version) APIServer {
	return newAPIServer(version)
}

// VersionString returns a string representation of the Version.
func (v *Version) VersionString() string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Micro, v.Additional)
}

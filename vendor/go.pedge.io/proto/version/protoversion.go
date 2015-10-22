package protoversion

import (
	"fmt"

	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewAPIServer creates a new APIServer for the given Version.
func NewAPIServer(version *Version) APIServer {
	return newAPIServer(version)
}

// GetServerVersion gets the server *Version given the *grpc.ClientConn.
func GetServerVersion(clientConn *grpc.ClientConn) (*Version, error) {
	return NewAPIClient(clientConn).GetVersion(
		context.Background(),
		&google_protobuf.Empty{},
	)
}

// VersionString returns a string representation of the Version.
func (v *Version) VersionString() string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Micro, v.Additional)
}

// Println prints the VersionString() value with fmt.Println(...)
func (v *Version) Println() {
	fmt.Println(v.VersionString())
}

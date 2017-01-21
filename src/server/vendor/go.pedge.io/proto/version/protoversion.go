package protoversion // import "go.pedge.io/proto/version"

import (
	"fmt"

	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// APIServerOptions are options when creating a new APIServer.
type APIServerOptions struct {
	DisableLogging bool
}

// NewAPIServer creates a new APIServer for the given Version.
func NewAPIServer(version *Version, options APIServerOptions) APIServer {
	return newAPIServer(version, options)
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

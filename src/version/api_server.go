package version

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	pb "github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type apiServer struct {
	pb.UnimplementedAPIServer

	version *pb.Version
	options APIServerOptions
}

func newAPIServer(version *pb.Version, options APIServerOptions) *apiServer {
	return &apiServer{version: version, options: options}
}

func (a *apiServer) GetVersion(ctx context.Context, request *emptypb.Empty) (response *pb.Version, err error) {
	return a.version, nil
}

// APIServerOptions are options when creating a new APIServer.
type APIServerOptions struct {
	DisableLogging bool
}

// NewAPIServer creates a new APIServer for the given Version.
func NewAPIServer(version *pb.Version, options APIServerOptions) pb.APIServer {
	return newAPIServer(version, options)
}

// GetServerVersion gets the server *Version given the *grpc.ClientConn.
func GetServerVersion(clientConn *grpc.ClientConn) (*pb.Version, error) {
	res, err := pb.NewAPIClient(clientConn).GetVersion(
		context.Background(),
		&emptypb.Empty{},
	)
	return res, errors.EnsureStack(err)
}

// String returns a string representation of the Version.
func String(v *pb.Version) string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Micro, v.Additional)
}

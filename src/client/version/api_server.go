package version

import (
	"fmt"

	pb "github.com/pachyderm/pachyderm/src/client/version/versionpb"

	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type apiServer struct {
	version     *pb.Version
	clusterInfo *pb.ClusterInfo
	options     APIServerOptions
}

func newAPIServer(version *pb.Version, clusterInfo *pb.ClusterInfo, options APIServerOptions) *apiServer {
	return &apiServer{version, clusterInfo, options}
}

func (a *apiServer) GetVersion(ctx context.Context, request *types.Empty) (response *pb.Version, err error) {
	return a.version, nil
}

func (a *apiServer) GetClusterInfo(ctx context.Context, request *types.Empty) (response *pb.ClusterInfo, err error) {
	return a.clusterInfo, nil
}

// APIServerOptions are options when creating a new APIServer.
type APIServerOptions struct {
	DisableLogging bool
}

// NewAPIServer creates a new APIServer for the given Version.
func NewAPIServer(version *pb.Version, clusterInfo *pb.ClusterInfo, options APIServerOptions) pb.APIServer {
	return newAPIServer(version, clusterInfo, options)
}

// GetServerVersion gets the server *Version given the *grpc.ClientConn.
func GetServerVersion(clientConn *grpc.ClientConn) (*pb.Version, error) {
	return pb.NewAPIClient(clientConn).GetVersion(
		context.Background(),
		&types.Empty{},
	)
}

// String returns a string representation of the Version.
func String(v *pb.Version) string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Micro, v.Additional)
}

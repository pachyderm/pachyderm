package pps

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type FilterFunc func([]*pfs.FileInfo) bool

// GRPCClient is an implementation of KV that talks over RPC.
type FilterGRPCClient struct{ client FilterClient }

func (c *FilterGRPCClient) Filter(fis []*pfs.FileInfo) (bool, error) {
	resp, err := c.client.Filter(context.Background(), &FilterRequest{Files: fis})
	if err != nil {
		return false, err
	}
	return resp.Filter, nil
}

// Here is the gRPC server that GRPCClient talks to.
type FilterGRPCServer struct {
	// This is the real implementation
	impl FilterFunc
}

func (s *FilterGRPCServer) Filter(ctx context.Context, req *FilterRequest) (*FilterResponse, error) {
	return &FilterResponse{Filter: s.impl(req.Files)}, nil
}

// Handshake is a common handshake that is shared by plugin and host.
var FilterHandshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "FILTER",
	MagicCookieValue: "FILTER",
}

// PluginMap is the map of plugins we can dispense.
var FilterPluginMap = map[string]plugin.Plugin{
	"filter_grpc": &FilterGRPCPlugin{},
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type FilterGRPCPlugin struct {
	plugin.Plugin
	Impl FilterFunc
}

func (p *FilterGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterFilterServer(s, &FilterGRPCServer{impl: p.Impl})
	return nil
}

func (p *FilterGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &FilterGRPCClient{client: NewFilterClient(c)}, nil
}

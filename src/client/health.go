package client

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"

	"github.com/gogo/protobuf/types"
)

func (c APIClient) Health() error {
	_, err := c.healthClient.Health(c.Ctx(), &types.Empty{})
	return grpcutil.ScrubGRPC(err)
}

package client

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

// Version returns the version of pachd as a string.
func (c APIClient) Version() (string, error) {
	v, err := c.VersionAPIClient.GetVersion(c.Ctx(), &types.Empty{})
	if err != nil {
		return "", grpcutil.ScrubGRPC(err)
	}
	return version.PrettyPrintVersion(v), nil
}

package pachyderm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"

	"google.golang.org/grpc"
)

func (b *backend) versionPath() *framework.Path {
	return &framework.Path{
		// Pattern uses modified version of framework.GenericNameRegex which
		// requires a single colon
		Pattern: "version",
		Fields: map[string]*framework.FieldSchema{
			"client_only": &framework.FieldSchema{
				Type:    framework.TypeBool,
				Default: false,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathVersion,
		},
	}
}

func (b *backend) pathVersion(ctx context.Context, req *logical.Request, d *framework.FieldData) (resp *logical.Response, retErr error) {
	b.Logger().Debug(fmt.Sprintf("(%s) %s received at %s", req.ID, req.Operation, req.Path))
	defer func() {
		b.Logger().Debug(fmt.Sprintf("(%s) %s finished at %s (success=%t)", req.ID, req.Operation, req.Path, retErr == nil && !resp.IsError()))
	}()

	// Determine if caller only wants client version
	clientOnlyIface, ok, err := d.GetOkErr("client_only")
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("%v: could not extract 'client_only' from request", err)), nil
	}
	if !ok {
		return errMissingField("client_only"), nil
	}
	clientOnly, ok := clientOnlyIface.(bool)
	if !ok {
		return logical.ErrorResponse(fmt.Sprintf("invalid type for param 'client_only' (expected bool but got %T)", clientOnlyIface)), nil
	}

	// Respond with client version if that was all that was requested
	response := &logical.Response{
		Data: map[string]interface{}{
			"client-version": version.PrettyPrintVersion(version.Version),
		},
	}
	if clientOnly {
		return response, nil
	}

	// Get Pachd address from config
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if len(config.PachdAddress) == 0 {
		return nil, errors.New("plugin is missing pachd_address")
	}

	// Create version API client
	clientConn, err := grpc.Dial(config.PachdAddress, client.PachDialOptions()...)
	if err != nil {
		return nil, err
	}
	versionClient := versionpb.NewAPIClient(clientConn)
	if err != nil {
		return nil, err
	}

	// Retrieve server version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serverVersion, err := versionClient.GetVersion(ctx, &types.Empty{})
	if err != nil {
		return nil, err
	}
	response.Data["server-version"] = version.PrettyPrintVersion(serverVersion)
	return response, nil
}

package pachyderm

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/version"
)

func (b *backend) versionPath() *framework.Path {
	return &framework.Path{
		Pattern: "version(?P<clientonly>/client-only)?",
		Fields: map[string]*framework.FieldSchema{
			"clientonly": &framework.FieldSchema{
				Type: framework.TypeString,
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
	response := &logical.Response{
		Data: map[string]interface{}{
			"client-version": version.PrettyPrintVersion(version.Version),
		},
	}

	// Determine if caller only wants client version
	clientOnlyIface, ok, err := d.GetOkErr("clientonly")
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("%v: could not extract 'clientonly' from request", err)), nil
	}
	if ok {
		if clientOnly, ok := clientOnlyIface.(string); ok && clientOnly != "" {
			return response, nil
		}
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
	client, err := pclient.NewFromAddress(config.PachdAddress)
	if err != nil {
		return nil, err
	}
	defer client.Close() // avoid leaking connections

	// Retrieve server version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serverVersion, err := client.GetVersion(ctx, &types.Empty{})
	if err != nil {
		return nil, err
	}
	response.Data["server-version"] = version.PrettyPrintVersion(serverVersion)
	return response, nil
}

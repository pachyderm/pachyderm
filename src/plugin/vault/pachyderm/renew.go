package pachyderm

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

	// Grab the token
	adminTokenRaw, ok := req.Auth.InternalData["pachyderm_admin_token"]
	if !ok {
		return nil, errors.New("no internal token found in the store")
	}
	adminToken, ok := adminTokenRaw.(string)
	if !ok {
		return nil, errors.New("stored admin token is not a string")
	}

	// Grab the token
	usernameRaw, ok := req.Auth.InternalData["username"]
	if !ok {
		return nil, errors.New("no internal token found in the store")
	}
	username, ok := usernameRaw.(string)
	if !ok {
		return nil, errors.New("stored admin token is not a string")
	}

	// Use the admin token to perform an action
	// for testing, hardcoding username to something else so that I can validate
	// renew has an effect:
	username = "tweetybird"
	userToken, err := b.generateUserCredentials(ctx, username, adminToken)
	if err != nil {
		return nil, err
	}
	fmt.Printf("generated new user token for (%v): %v\n", username, userToken)

	ttl, maxTTL, err := b.SanitizeTTLStr("30s", "1h")
	if err != nil {
		return nil, err
	}

	return framework.LeaseExtend(ttl, maxTTL, b.System())(ctx, req, d)
}

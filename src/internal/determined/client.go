package determined

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	det "github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/rbacv1"
	"github.com/determined-ai/determined/proto/pkg/userv1"
	"github.com/determined-ai/determined/proto/pkg/workspacev1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	mlc "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Config struct {
	MasterURL string
	Username  string
	Password  string
	TLS       bool
}

func NewClient(ctx context.Context, masterURL string, withTLS bool) (client det.DeterminedClient, close func() error, retErr error) {
	tlsOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
	if withTLS {
		tlsOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		}))
	}
	determinedURL, err := url.Parse(masterURL)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "parsing determined url %q", masterURL)
	}
	conn, err := grpc.DialContext(ctx, determinedURL.Host, tlsOpt, grpc.WithStreamInterceptor(mlc.LogStream), grpc.WithUnaryInterceptor(mlc.LogUnary))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "dialing determined at %q", determinedURL.Host)
	}
	dc := det.NewDeterminedClient(conn)
	return dc, conn.Close, nil
}

func GetRole(ctx context.Context, dc det.DeterminedClient, name string) (*rbacv1.Role, error) {
	resp, err := dc.ListRoles(ctx, &det.ListRolesRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "search determined roles")
	}
	for _, r := range resp.Roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, errors.Errorf("workspaceEditor role not found")
}

func GetWorkspaces(ctx context.Context, dc det.DeterminedClient, workspaces []string) ([]*workspacev1.Workspace, error) {
	workspacesResp, err := dc.GetWorkspaces(ctx, &det.GetWorkspacesRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "list determined workspaces")
	}
	workspaceSet := make(map[string]struct{})
	for _, w := range workspaces {
		workspaceSet[w] = struct{}{}
	}
	res := make([]*workspacev1.Workspace, 0)
	for _, dw := range workspacesResp.Workspaces {
		if _, ok := workspaceSet[dw.Name]; ok {
			res = append(res, dw)
			delete(workspaceSet, dw.Name)
		}
	}
	if len(workspaceSet) > 0 {
		var errWorkspaces []string
		for w := range workspaceSet {
			errWorkspaces = append(errWorkspaces, w)
		}
		return nil, errors.Errorf("requested workspaces not found: %v", errWorkspaces)
	}
	return res, nil
}

func CreateWorkspace(ctx context.Context, dc det.DeterminedClient, workspace string) (*workspacev1.Workspace, error) {
	w, err := dc.PostWorkspace(ctx, &det.PostWorkspaceRequest{Name: workspace})
	if err != nil {
		return nil, errors.Wrapf(err, "post determined workspace %q", workspace)
	}
	return w.Workspace, nil
}

func MintToken(ctx context.Context, dc det.DeterminedClient, username, password string) (string, error) {
	loginResp, err := dc.Login(ctx, &det.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", errors.Wrap(err, "login as determined user")
	}
	return loginResp.Token, nil
}

func WithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-user-token", fmt.Sprintf("Bearer %s", token))
}

func GetUsers(ctx context.Context, dc det.DeterminedClient) ([]*userv1.User, error) {
	usersResp, err := dc.GetUsers(ctx, &det.GetUsersRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "get determined users")
	}
	return usersResp.Users, nil
}

func AssignRole(ctx context.Context, dc det.DeterminedClient, userId int32, roleId int32, workspaces []*workspacev1.Workspace) error {
	var roleAssignments []*rbacv1.UserRoleAssignment
	for _, w := range workspaces {
		roleAssignments = append(roleAssignments,
			&rbacv1.UserRoleAssignment{
				UserId: userId,
				RoleAssignment: &rbacv1.RoleAssignment{
					Role: &rbacv1.Role{
						RoleId: roleId,
					},
					ScopeWorkspaceId: &w.Id,
				},
			},
		)
	}
	if _, err := dc.AssignRoles(ctx, &det.AssignRolesRequest{
		UserRoleAssignments: roleAssignments,
	}); err != nil && !strings.Contains(err.Error(), "already exists") {
		return errors.Wrap(err, "assign pipeline's determined user editor role")
	}
	return nil
}

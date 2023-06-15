package server

import (
	"context"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	det "github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/rbacv1"
	"github.com/determined-ai/determined/proto/pkg/userv1"
	"github.com/determined-ai/determined/proto/pkg/workspacev1"
)

func (a *apiServer) hookDeterminedPipeline(ctx context.Context, pi *pps.PipelineInfo) error {
	var cf context.CancelFunc
	ctx, cf = context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	return backoff.RetryUntilCancel(ctx, func() error {
		var opts []grpc.DialOption
		conn, err := grpc.Dial(a.env.Config.DeterminedURL, opts)
		if err != nil {
			return err
		}
		defer conn.Close()
		dc := det.NewDeterminedClient(conn)
		tok, err := a.mintDeterminedToken(ctx, dc)
		if err != nil {
			return errors.Wrap(err, "mint determined token")
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "x-user-token", fmt.Sprintf("Bearer %s", tok))
		detWorkspaces, err := resolveDeterminedWorkspaces(ctx, dc, pi.Details.Determined.Workspaces)
		if err != nil {
			return err
		}
		if err := validateWorkspacePermissions(ctx, dc, detWorkspaces); err != nil {
			return err
		}
		if pi.Details.Determined.Password == "" {
			pi.Details.Determined.Password = uuid.NewWithoutDashes()
		}
		// TODO: if user already exists, do we want to still run role assignment for them?
		// this would involve a querying for the user's id
		userId, err := provisionDeterminedPipelineUser(ctx, dc, pi)
		if err != nil {
			return err
		}
		roleId, err := workspaceEditorRoleId(ctx, dc)
		if err != nil {
			return err
		}
		if err := assignDeterminedPipelineRole(ctx, dc, userId, roleId, detWorkspaces); err != nil {
			return err
		}
		return nil
	}, &backoff.ZeroBackOff{}, nil)
}

func workspaceEditorRoleId(ctx context.Context, dc det.DeterminedClient) (int32, error) {
	resp, err := dc.ListRoles(ctx, &det.ListRolesRequest{})
	if err != nil {
		return 0, errors.Wrap(err, "search determined roles")
	}
	for _, r := range resp.Roles {
		if r.Name == "Editor" {
			return r.RoleId, nil
		}
	}
	return 0, errors.Errorf("workspaceEditor role not found")
}

func resolveDeterminedWorkspaces(ctx context.Context, dc det.DeterminedClient, workspaces []string) ([]*workspacev1.Workspace, error) {
	workspacesResp, err := dc.GetWorkspaces(ctx, &det.GetWorkspacesRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "list determined workspaces")
	}
	workspaceSet := make(map[string]struct{})
	for _, w := range workspacesResp.Workspaces {
		workspaceSet[w.Name] = struct{}{}
	}
	res := make([]*workspacev1.Workspace, len(workspaces))
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

func (a *apiServer) mintDeterminedToken(ctx context.Context, dc det.DeterminedClient) (string, error) {
	loginResp, err := dc.Login(ctx, &det.LoginRequest{
		Username: a.env.Config.DeterminedUsername,
		Password: a.env.Config.DeterminedPassword,
	})
	if err != nil {
		return "", errors.Wrap(err, "login as determined user")
	}
	return loginResp.Token, nil
}

func validateWorkspacePermissions(ctx context.Context, dc det.DeterminedClient, dws []*workspacev1.Workspace) error {
	resp, err := dc.GetPermissionsSummary(ctx, &det.GetPermissionsSummaryRequest{})
	if err != nil {
		return errors.Wrap(err, "list determined user permissions")
	}
	assignRoleIds := make(map[int32]struct{})
	for _, r := range resp.Roles {
		for _, p := range r.Permissions {
			if p.Id == rbacv1.PermissionType_PERMISSION_TYPE_ASSIGN_ROLES {
				assignRoleIds[r.RoleId] = struct{}{}
				break
			}
		}
	}
	dwMap := make(map[int32]*workspacev1.Workspace)
	for _, dw := range dws {
		dwMap[dw.Id] = dw
	}
	for _, a := range resp.Assignments {
		if _, ok := assignRoleIds[a.RoleId]; ok {
			if a.ScopeCluster {
				return nil
			}
			for _, id := range a.ScopeWorkspaceIds {
				delete(dwMap, id)
			}
		}
	}
	if len(dwMap) > 0 {
		var workspaces []string
		for _, dw := range dwMap {
			workspaces = append(workspaces, dw.Name)
		}
		return errors.Errorf("requested workspaces don't exist: %v", workspaces)
	}
	return nil
}

func provisionDeterminedPipelineUser(ctx context.Context, dc det.DeterminedClient, pi *pps.PipelineInfo) (int32, error) {
	resp, err := dc.PostUser(ctx, &det.PostUserRequest{
		User:     &userv1.User{Username: pipelineUserName(pi.Pipeline)},
		Password: pi.Details.Determined.Password,
	})
	// TODO: handle existing existing user
	if err != nil {
		return 0, errors.Wrap(err, "provision determined user")
	}
	return resp.User.Id, nil
}

func assignDeterminedPipelineRole(ctx context.Context, dc det.DeterminedClient, userId int32, roleId int32, workspaces []*workspacev1.Workspace) error {
	var roleAssignments []*rbacv1.UserRoleAssignment
	for _, w := range workspaces {
		roleAssignments = append(roleAssignments,
			&*rbacv1.UserRoleAssignment{
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
	}); err != nil {
		return errors.Wrap(err, "assign pipeline's determined user editor role")
	}
	return nil
}

func pipelineUserName(p *pps.Pipeline) string {
	return p.String()
}

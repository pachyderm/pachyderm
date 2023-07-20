package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	det "github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/rbacv1"
	"github.com/determined-ai/determined/proto/pkg/userv1"
	"github.com/determined-ai/determined/proto/pkg/workspacev1"
	mlc "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
)

func (a *apiServer) hookDeterminedPipeline(ctx context.Context, p *pps.Pipeline, workspaces []string, pipPassword string, whoami string) error {
	var cf context.CancelFunc
	ctx, cf = context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	errCnt := 0
	// right now the entire integration is specifc to auth, so first check that auth is active
	if err := backoff.RetryUntilCancel(ctx, func() error {
		tlsOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
		if a.env.Config.DeterminedTLS {
			tlsOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			}))
		}
		determinedURL, err := url.Parse(a.env.Config.DeterminedURL)
		if err != nil {
			return errors.Wrapf(err, "parsing determined url %q", a.env.Config.DeterminedURL)
		}
		conn, err := grpc.DialContext(ctx, determinedURL.Host, tlsOpt, grpc.WithStreamInterceptor(mlc.LogStream), grpc.WithUnaryInterceptor(mlc.LogUnary))
		if err != nil {
			return errors.Wrapf(err, "dialing determined at %q", determinedURL.Host)
		}
		defer conn.Close()
		dc := det.NewDeterminedClient(conn)
		tok, err := mintDeterminedToken(ctx, dc, a.env.Config.DeterminedUsername, a.env.Config.DeterminedPassword)
		if err != nil {
			return err
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "x-user-token", fmt.Sprintf("Bearer %s", tok))
		detWorkspaces, err := resolveDeterminedWorkspaces(ctx, dc, workspaces)
		if err != nil {
			return err
		}
		if err := validateWorkspacePermissions(ctx, dc, detWorkspaces); err != nil {
			return err
		}
		if err := validateWorkspacePermsByUser(ctx, dc, whoami, detWorkspaces); err != nil {
			return err
		}
		userId, err := provisionDeterminedPipelineUser(ctx, dc, p, pipPassword)
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
	}, &backoff.ZeroBackOff{}, func(err error, _ time.Duration) error {
		log.Info(ctx, "determined hook error", zap.Error(err))
		errCnt++
		if errCnt >= 3 {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
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

func mintDeterminedToken(ctx context.Context, dc det.DeterminedClient, username, password string) (string, error) {
	loginResp, err := dc.Login(ctx, &det.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", errors.Wrap(err, "login as determined user")
	}
	return loginResp.Token, nil
}

func validateWorkspacePermsByUser(ctx context.Context, dc det.DeterminedClient, user string, ws []*workspacev1.Workspace) error {
	detUser, err := dc.GetUserByUsername(ctx, &det.GetUserByUsernameRequest{Username: user})
	if err != nil {
		return errors.Wrapf(err, "get determined user %q", user)
	}
	resp, err := dc.GetRolesAssignedToUser(ctx, &det.GetRolesAssignedToUserRequest{UserId: detUser.User.Id})
	if err != nil {
		return errors.Wrapf(err, "get roles assigned to user %q", user)
	}
	workspaceMap := make(map[int32]*workspacev1.Workspace)
	for _, w := range ws {
		workspaceMap[w.Id] = w
	}
	for _, r := range resp.Roles {
		relevant := false
		for _, p := range r.Role.Permissions {
			if p.Id == rbacv1.PermissionType_PERMISSION_TYPE_EDIT_MODEL_REGISTRY {
				relevant = true
			}
		}
		if !relevant {
			continue
		}
		for _, a := range r.UserRoleAssignments {
			// if user is editor of the cluster, exit early
			if a.RoleAssignment.ScopeCluster {
				return nil
			}
			delete(workspaceMap, *a.RoleAssignment.ScopeWorkspaceId)
		}
	}
	if len(workspaceMap) > 0 {
		var workspaces []string
		for _, w := range workspaceMap {
			workspaces = append(workspaces, w.Name)
		}
		return errors.Errorf("user %q doesn't have editor access for determined workspaces: %v", user, workspaces)
	}
	return nil
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
		return errors.Errorf("access required for determined workspaces: %v", workspaces)
	}
	return nil
}

func provisionDeterminedPipelineUser(ctx context.Context, dc det.DeterminedClient, p *pps.Pipeline, password string) (int32, error) {
	resp, err := dc.PostUser(ctx, &det.PostUserRequest{
		User:     &userv1.User{Username: pipelineUserName(p), Active: true},
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument && strings.Contains(err.Error(), "user already exists") {
			usersResp, err := dc.GetUsers(ctx, &det.GetUsersRequest{Name: pipelineUserName(p)})
			if err != nil {
				return 0, errors.Wrapf(err, "get determined user %q", pipelineUserName(p))
			}
			if len(usersResp.Users) == 0 {
				return 0, errors.Wrapf(err, "no determined users return for user %q", pipelineUserName(p))
			}
			if _, err := dc.SetUserPassword(ctx, &det.SetUserPasswordRequest{
				UserId:   usersResp.Users[0].Id,
				Password: password,
			}); err != nil {
				return 0, errors.Wrapf(err, "set password for user %q", pipelineUserName(p))
			}
			return usersResp.Users[0].Id, nil
		}
		return 0, errors.Wrap(err, "provision determined user")
	}
	return resp.User.Id, nil
}

func assignDeterminedPipelineRole(ctx context.Context, dc det.DeterminedClient, userId int32, roleId int32, workspaces []*workspacev1.Workspace) error {
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

func pipelineUserName(p *pps.Pipeline) string {
	// users with names containing '/' don't work correctly in determined.
	return strings.ReplaceAll(p.String(), "/", "_")
}

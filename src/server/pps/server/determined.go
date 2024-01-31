package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"

	det "github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/rbacv1"
	"github.com/determined-ai/determined/proto/pkg/userv1"
	"github.com/determined-ai/determined/proto/pkg/workspacev1"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	mlc "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type detConfig struct {
	MasterURL string
	Username  string
	Password  string
	TLS       bool
}

func (a *apiServer) getDetConfig() detConfig {
	return detConfig{
		MasterURL: a.env.Config.DeterminedURL,
		Username:  a.env.Config.DeterminedUsername,
		Password:  a.env.Config.DeterminedPassword,
		TLS:       a.env.Config.DeterminedTLS,
	}
}

func newInClusterDetClient(ctx context.Context, cfg detConfig) (det.DeterminedClient, context.Context, context.CancelFunc, error) {
	tlsOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
	if cfg.TLS {
		tlsOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		}))
	}
	determinedURL, err := url.Parse(cfg.MasterURL)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "parsing determined url %q", cfg.MasterURL)
	}
	conn, err := grpc.DialContext(ctx, determinedURL.Host, tlsOpt, grpc.WithStreamInterceptor(mlc.LogStream), grpc.WithUnaryInterceptor(mlc.LogUnary))
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "dialing determined at %q", determinedURL.Host)
	}
	dc := det.NewDeterminedClient(conn)
	tok, err := mintDeterminedToken(ctx, dc, cfg.Username, cfg.Password)
	if err != nil {
		return nil, nil, nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-user-token", fmt.Sprintf("Bearer %s", tok))
	return dc, ctx, func() { conn.Close() }, nil
}

func (a *apiServer) hookDeterminedPipeline(ctx context.Context, p *pps.Pipeline, workspaces []string, pipPassword string, whoami string) error {
	config := detConfig{
		MasterURL: a.env.Config.DeterminedURL,
		Username:  a.env.Config.DeterminedUsername,
		Password:  a.env.Config.DeterminedPassword,
		TLS:       a.env.Config.DeterminedTLS,
	}
	var cf context.CancelFunc
	ctx, cf = context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	errCnt := 0
	// right now the entire integration is specifc to auth, so first check that auth is active
	if err := backoff.RetryUntilCancel(ctx, func() error {
		dc, ctx, cf, err := newInClusterDetClient(ctx, config)
		if err != nil {
			return errors.Wrap(err, "set up in cluster determined client")
		}
		defer cf()
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

type pipelineGetter interface {
	ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error
}

func gcDetUsers(ctx context.Context, config detConfig, period time.Duration, pipGetter pipelineGetter, secrets corev1.SecretInterface) {
	if config.MasterURL == "" {
		log.Info(ctx, "Determined not configured. Skipping Determined user garbage collection.")
		return
	}
	log.Info(ctx, "Starting Determined user garbage collection.")
	err := backoff.RetryUntilCancel(ctx, func() error {
		dc, detCtx, cf, err := newInClusterDetClient(ctx, config)
		if err != nil {
			return errors.Wrap(err, "setup in cluster determined client for garbage collection")
		}
		defer cf()
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for {
			pipelines := make(map[string]struct{})
			if err := pipGetter.ListPipelineInfo(ctx, func(pi *pps.PipelineInfo) error {
				pipelines[pi.Pipeline.String()] = struct{}{}
				return nil
			}); err != nil {
				return errors.Wrap(err, "list pipelines for determined user garbage collection")
			}
			ss, err := secrets.List(ctx, v1.ListOptions{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ListOptions",
					APIVersion: "v1",
				},
				LabelSelector: "suite=pachyderm,determined=true"})
			if err != nil {
				return errors.Wrap(err, "list determined pipelien user secrets")
			}
			for _, s := range ss.Items {
				p := &pps.Pipeline{
					Name:    s.Labels["pipelineName"],
					Project: &pfs.Project{Name: s.Labels["project"]},
				}
				// make sure the secret was created long enough ago so that it can be assumed it's not in the middle of being created
				if s.CreationTimestamp.Add(time.Minute).Before(time.Now()) {
					continue
				}
				if _, ok := pipelines[p.String()]; !ok {
					ctx, end := log.SpanContextL(ctx, "", log.DebugLevel, zap.String("pipeline", p.String()))
					if err := func() error {
						u, err := getDetPipelineUser(detCtx, dc, p)
						if err != nil {
							return err
						}
						if _, err := dc.PatchUser(detCtx, &det.PatchUserRequest{
							UserId: u.Id,
							User: &userv1.PatchUser{
								Active: &wrapperspb.BoolValue{Value: false},
							}}); err != nil {
							return errors.Wrapf(err, "inactivate user")
						}
						if err := secrets.Delete(ctx, detUserSecretName(p), v1.DeleteOptions{}); err != nil {
							return errors.Wrapf(err, "delete determined pipeline user secret %q", detUserSecretName(p))
						}
						return nil
					}(); err != nil {
						end(log.Errorp(&err))
					}
				}
			}
			select {
			case <-ctx.Done():
				return errors.Wrap(context.Cause(ctx), "determined user garbage collection")
			case <-ticker.C:
			}
		}
	}, backoff.RetryEvery(5*time.Minute), nil)
	log.Info(ctx, "determined user GC context cancelled", zap.Error(err))
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
	// Try to get the user first
	u, err := getDetPipelineUser(ctx, dc, p)
	if err != nil {
		// Temporary thing, probably not the best to use strings.Contains here
		if strings.Contains(err.Error(), "no determined users return") {
			u = nil
		} else {
			return 0, err
		}
	}

	// If the user exists, check and update if needed
	if u != nil {
		if !u.Active {
			if _, err := dc.PatchUser(ctx, &det.PatchUserRequest{
				UserId: u.Id,
				User: &userv1.PatchUser{
					Active: &wrapperspb.BoolValue{Value: true},
				},
			}); err != nil {
				return 0, errors.Wrapf(err, "reactivate user %q", pipelineUserName(p))
			}
		}
		if _, err := dc.SetUserPassword(ctx, &det.SetUserPasswordRequest{
			UserId:   u.Id,
			Password: password,
		}); err != nil {
			return 0, errors.Wrapf(err, "set password for user %q", pipelineUserName(p))
		}
		return u.Id, nil
	}

	// If no user exists, proceed with user creation
	resp, err := dc.PostUser(ctx, &det.PostUserRequest{
		User:     &userv1.User{Username: pipelineUserName(p), Active: true},
		Password: password,
	})
	if err != nil {
		// Handle the error here, API call issue should be thrown as an error
		return 0, errors.Wrap(err, "provision determined user")
	}

	return resp.User.Id, nil
}

func getDetPipelineUser(ctx context.Context, dc det.DeterminedClient, p *pps.Pipeline) (*userv1.User, error) {
	usersResp, err := dc.GetUsers(ctx, &det.GetUsersRequest{Name: pipelineUserName(p)})
	if err != nil {
		return nil, errors.Wrapf(err, "get determined user %q", pipelineUserName(p))
	}
	if len(usersResp.Users) == 0 {
		return nil, errors.Wrapf(err, "no determined users return for user %q", pipelineUserName(p))
	}
	return usersResp.Users[0], nil
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

func detUserSecretName(p *pps.Pipeline) string {
	return p.Project.Name + "-" + p.Name + "-det"
}

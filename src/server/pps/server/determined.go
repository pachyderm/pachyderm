package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type detLoginResponse struct {
	Token string `json:"token"`
}

type detPostUserResponse struct {
	User detUser `json:"user"`
}

type detUser struct {
	Id int `json:"id"`
}

type detWorkspace struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type detRole struct {
	Id          int             `json:"id"`
	Permissions []detPermission `json:"permissions"`
}

type detPermission struct {
	Id string `json:"id"`
}

type detPermissionsSummaryResponse struct {
	Roles       []detRole       `json:"roles"`
	Assignments []detAssignment `json:"assignments"`
}

type detAssignment struct {
	RoleId            int   `json:"roleId"`
	ScopeWorkspaceIds []int `json:"scopeWorkspaceIds"`
	ScopeCluster      bool  `json:"scopeCluster"`
}

type detUserRoleAssignmentsRequest struct {
	UserRoleAssignments []detUserRoleAssignment `json:"userRoleAssignments"`
}

type detUserRoleAssignment struct {
	UserId         int               `json:"userId"`
	RoleAssignment detRoleAssignment `json:"roleAssignment"`
}

type detRoleAssignment struct {
	Role             detRoleAssignmentRole `json:"role"`
	ScopeWorkspaceId int                   `json:"scopeWorkspaceId"`
}

type detRoleAssignmentRole struct {
	RoleId int `json:"roleId"`
}

type detSearchRoleResponse struct {
	Roles []detSearchRole `json:"roles"`
}

type detSearchRole struct {
	RoleId int    `json:"roleId"`
	Name   string `json:"name"`
}

func (a *apiServer) hookDeterminedPipeline(ctx context.Context, pi *pps.PipelineInfo) error {
	tok, err := a.mintDeterminedToken(ctx)
	if err != nil {
		return errors.Wrap(err, "mint determined token")
	}
	detWorkspaces, err := a.resolveDeterminedWorkspaces(ctx, tok, pi.Details.Determined.Workspaces)
	if err != nil {
		return err
	}
	if err := a.validateWorkspacePermissions(ctx, tok, detWorkspaces); err != nil {
		return err
	}
	if pi.Details.Determined.Password == "" {
		pi.Details.Determined.Password = uuid.NewWithoutDashes()
	}
	// TODO: if user already exists, do we want to still run role assignment for them?
	// this would involve a querying for the user's id
	userId, err := a.provisionDeterminedPipelineUser(ctx, pi, tok)
	if err != nil {
		return err
	}
	roleId, err := a.workspaceEditorRoleId(ctx, tok)
	if err != nil {
		return err
	}
	if err := a.assignDeterminedPipelineRole(ctx, tok, userId, roleId, detWorkspaces); err != nil {
		return err
	}
	return nil
}

func (a *apiServer) workspaceEditorRoleId(ctx context.Context, token string) (int, error) {
	data, err := a.determinedHttpRequest(ctx, "POST", "/api/v1/roles/search", token, map[string]interface{}{
		"offset": 0,
		"limit":  0,
	})
	if err != nil {
		return 0, errors.Wrap(err, "search determined roles")
	}
	var roleResponse detSearchRoleResponse
	if err := json.Unmarshal(data, &roleResponse); err != nil {
		return 0, errors.Wrap(err, "unmarshal search determined roles response")
	}
	for _, r := range roleResponse.Roles {
		if r.Name == "Editor" {
			return r.RoleId, nil
		}
	}
	return 0, errors.Errorf("workspaceEditor role not found")
}

func (a *apiServer) resolveDeterminedWorkspaces(ctx context.Context, token string, workspaces []string) ([]detWorkspace, error) {
	data, err := a.determinedHttpRequest(ctx, "GET", "/api/v1/workspaces", token, nil)
	if err != nil {
		return nil, errors.Wrap(err, "list determined workspaces")
	}
	var detWorkspaces []detWorkspace
	if err := json.Unmarshal(data, &detWorkspaces); err != nil {
		return nil, errors.Wrap(err, "unmarshal determined workspaces")
	}
	workspaceSet := make(map[string]struct{})
	for _, w := range workspaces {
		workspaceSet[w] = struct{}{}
	}
	res := make([]detWorkspace, len(workspaces))
	for _, dw := range detWorkspaces {
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

func (a *apiServer) mintDeterminedToken(ctx context.Context) (string, error) {
	u, p, err := a.determinedCredentials(ctx)
	if err != nil {
		return "", err
	}
	m := map[string]interface{}{
		"username": u,
		"password": p,
	}
	data, err := a.determinedHttpRequest(ctx, "POST", "/api/v1/auth/login", "", m)
	if err != nil {
		return "", errors.Wrap(err, "login as determined user")
	}
	var respObj detLoginResponse
	if err := json.Unmarshal(data, &respObj); err != nil {
		return "", errors.Wrap(err, "unmarshal determined token")
	}
	return respObj.Token, nil
}

func (a *apiServer) determinedCredentials(ctx context.Context) (string, string, error) {
	secretName := a.env.Config.DeterminedCredentialsSecret
	secret, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return "", "", errors.Errorf("missing Kubernetes secret %s", secretName)
		}
		return "", "", errors.Wrapf(err, "could not get Kubernetes secret %s", secretName)
	}
	var username string
	var password string
	var ok bool
	if username, ok = secret.StringData["username"]; !ok {
		return "", "", errors.Errorf("username missing in Kubernetes secret %q", secretName)
	}
	if password, ok = secret.StringData["password"]; !ok {
		return "", "", errors.Errorf("password missing in Kubernetes secret %q", secretName)
	}
	return username, password, nil
}

func (a *apiServer) validateWorkspacePermissions(ctx context.Context, token string, dws []detWorkspace) error {
	data, err := a.determinedHttpRequest(ctx, "GET", "/api/v1/permissions/summary", token, nil)
	if err != nil {
		return errors.Wrap(err, "list determined user permissions")
	}
	var resp detPermissionsSummaryResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return errors.Wrap(err, "unmarshal determined permissions summary response")
	}
	assignRoleIds := make(map[int]struct{})
	for _, r := range resp.Roles {
		for _, p := range r.Permissions {
			if p.Id == "PERMISSION_TYPE_ASSIGN_ROLES" {
				assignRoleIds[r.Id] = struct{}{}
				break
			}
		}
	}
	dwMap := make(map[int]detWorkspace)
	for _, dw := range dwMap {
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

func (a *apiServer) provisionDeterminedPipelineUser(ctx context.Context, pi *pps.PipelineInfo, token string) (int, error) {
	m := map[string]interface{}{
		"username": pipelineUserName(pi.Pipeline),
		"password": pi.Details.Determined.Password,
	}
	// TODO: handle existing existing user
	data, err := a.determinedHttpRequest(ctx, "POST", "/api/v1/users", token, m)
	if err != nil {
		return 0, errors.Wrap(err, "provision determined user")
	}
	var resp detPostUserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, errors.Wrap(err, "unmarshal determined provision user response")
	}
	return resp.User.Id, nil
}

func (a *apiServer) assignDeterminedPipelineRole(ctx context.Context, token string, userId int, roleId int, workspaces []detWorkspace) error {
	var roleAssignments []detUserRoleAssignment
	for _, w := range workspaces {
		roleAssignments = append(roleAssignments,
			detUserRoleAssignment{
				UserId: userId,
				RoleAssignment: detRoleAssignment{
					Role: detRoleAssignmentRole{
						RoleId: roleId,
					},
					ScopeWorkspaceId: w.Id,
				},
			},
		)
	}
	body := detUserRoleAssignmentsRequest{UserRoleAssignments: roleAssignments}
	// craft request with workspace IDs
	_, err := a.determinedHttpRequest(ctx, "POST", "/api/v1/roles/add-assignments", token, body)
	if err != nil {
		return errors.Wrap(err, "assign pipeline's determined user editor role")
	}
	return nil
}

func (a *apiServer) determinedHttpRequest(ctx context.Context, method, endpoint, token string, body any) ([]byte, error) {
	var b *bytes.Buffer
	if body != nil {
		b = new(bytes.Buffer)
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return nil, errors.Wrapf(err, "encode to json: %v", body)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, a.env.Config.DeterminedURL+endpoint, b)
	if err != nil {
		return nil, errors.Wrap(err, "create determined http request")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "execute determined http request")
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "read determined http response body")
		}
		return data, nil
	}
	return nil, errors.New(resp.Status)
}

func pipelineUserName(p *pps.Pipeline) string {
	return p.String()
}

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

func (a *apiServer) hookDeterminedPipeline(ctx context.Context, pi *pps.PipelineInfo) error {
	var errs error
	for _, w := range pi.Details.Determined.Workspaces {
		if err := a.validateWorkspacePermissions(ctx, w); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "validate permissions on workspace %q", w))
		}
	}
	if errs != nil {
		return errs
	}
	if pi.Details.Determined.Password == "" {
		pi.Details.Determined.Password = uuid.NewWithoutDashes()
	}
	tok, err := a.mintDeterminedToken(ctx)
	if err != nil {
		return errors.Wrap(err, "mint determined token")
	}
	userId, err := a.provisionDeterminedPipelineUser(ctx, pi, tok)
	if err != nil {
		return err
	}
	if err := a.assignDeterminedPipelineRole(ctx, pi, tok, userId); err != nil {
		return err
	}
	return nil
}

func (a *apiServer) mintDeterminedToken(ctx context.Context) (string, error) {
	u, p, err := a.determinedCredentials(ctx)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	m := map[string]interface{}{
		"username": u,
		"password": p,
	}
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(m); err != nil {
		return "", errors.Wrapf(err, "encode to json: %v", m)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", a.env.Config.DeterminedURL+"/api/v1/auth/login", body)
	if err != nil {
		return "", errors.Wrap(err, "create request to login as determined user")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "login as determined user")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read login response body")
	}
	var respObj detLoginResponse
	if err := json.Unmarshal(data, &respObj); err != nil {
		return "", errors.Wrap(err, "unmarshal determined token")
	}
	return respObj.Token, nil
}

type detLoginResponse struct {
	Token string `json:"token"`
}

type detPostUserResponse struct {
	User detUser `json:"user"`
}

type detUser struct {
	Id int `json:"id"`
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

// validate the system determined user has Editor Permissions on all of the workspaces
func (a *apiServer) validateWorkspacePermissions(ctx context.Context, workspaces string) error {
	return nil
}

// TODO: grant user workspace editor role
// returns user ID
func (a *apiServer) provisionDeterminedPipelineUser(ctx context.Context, pi *pps.PipelineInfo, bearerToken string) (int, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	body, err := provisionUserRequestBody(pi)
	if err != nil {
		return 0, err
	}
	determinedURI := a.env.Config.DeterminedURL
	req, err := http.NewRequestWithContext(ctx, "POST", determinedURI+"/api/v1/users", body)
	// TODO: handle existing existing user
	if err != nil {
		return 0, errors.Wrap(err, "create request to provision determined user")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bearerToken)
	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "provision determined user")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "read provision response body")
	}
	var respObj detPostUserResponse
	if err := json.Unmarshal(data, &respObj); err != nil {
		return 0, errors.Wrap(err, "unmarshal determined provision user response")
	}
	return respObj.User.Id, nil
}

func (a *apiServer) assignDeterminedPipelineRole(ctx context.Context, pi *pps.PipelineInfo, bearerToken string, userId int) error {
	client := &http.Client{Timeout: 10 * time.Second}
	determinedURI := a.env.Config.DeterminedURL
	req, err := http.NewRequestWithContext(ctx, "POST", determinedURI+"/api/v1/roles/add-assignments", nil)
	if err != nil {
		return errors.Wrap(err, "create request to add determined role assignment")
	}
	if _, err := client.Do(req); err != nil {
		return errors.Wrap(err, "add determined role assignment")
	}
	return nil
}

func provisionUserRequestBody(pi *pps.PipelineInfo) (*bytes.Buffer, error) {
	m := map[string]interface{}{
		"username": pipelineUserName(pi.Pipeline),
		"password": pi.Details.Determined.Password,
	}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(m); err != nil {
		return nil, errors.Wrapf(err, "encode to json: %v", m)
	}
	return b, nil
}

func pipelineUserName(p *pps.Pipeline) string {
	return p.String()
}

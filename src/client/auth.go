package client

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
)

// IsAuthActive returns whether auth is activated on the cluster
func (c APIClient) IsAuthActive() (bool, error) {
	_, err := c.GetRoleBinding(c.Ctx(), &auth.GetRoleBindingRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER},
	})
	switch {
	case err == nil:
		return true, nil
	case auth.IsErrNotAuthorized(err):
		return true, nil
	case auth.IsErrNotSignedIn(err):
		return true, nil
	case auth.IsErrNotActivated(err):
		return false, nil
	default:
		return false, grpcutil.ScrubGRPC(err)
	}
}

func (c APIClient) GetClusterRoleBinding() (*auth.RoleBinding, error) {
	resp, err := c.GetRoleBinding(c.Ctx(), &auth.GetRoleBindingRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER},
	})
	if err != nil {
		return nil, err
	}
	return resp.Binding, nil
}

func (c APIClient) ModifyClusterRoleBinding(principal string, roles []string) error {
	_, err := c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
		Resource:  &auth.Resource{Type: auth.ResourceType_CLUSTER},
		Principal: principal,
		Roles:     roles,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c APIClient) GetProjectRoleBinding(project string) (*auth.RoleBinding, error) {
	resp, err := c.GetRoleBinding(c.Ctx(), &auth.GetRoleBindingRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project},
	})
	if err != nil {
		return nil, err
	}
	return resp.Binding, nil
}

// Deprecated: use GetProjectRepoRoleBinding instead.
func (c APIClient) GetRepoRoleBinding(repoName string) (*auth.RoleBinding, error) {
	return c.GetProjectRepoRoleBinding(pfs.DefaultProjectName, repoName)
}

// Return the roles bound to a repo within a project.
func (c APIClient) GetProjectRepoRoleBinding(projectName, repoName string) (*auth.RoleBinding, error) {
	resp, err := c.GetRoleBinding(c.Ctx(), &auth.GetRoleBindingRequest{
		Resource: NewProjectRepo(projectName, repoName).AuthResource(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Binding, nil
}

// Deprecated: use ModifyProjectRepoRoleBinding instead.
func (c APIClient) ModifyRepoRoleBinding(repoName, principal string, roles []string) error {
	return c.ModifyProjectRepoRoleBinding(pfs.DefaultProjectName, repoName, principal, roles)
}

// Update the roles bound to a repo within a project.
func (c APIClient) ModifyProjectRepoRoleBinding(projectName, repoName, principal string, roles []string) error {
	_, err := c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
		Resource:  NewProjectRepo(projectName, repoName).AuthResource(),
		Principal: principal,
		Roles:     roles,
	})
	if err != nil {
		return err
	}
	return nil
}

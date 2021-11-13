package client

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
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

func (c APIClient) GetRepoRoleBinding(repo string) (*auth.RoleBinding, error) {
	resp, err := c.GetRoleBinding(c.Ctx(), &auth.GetRoleBindingRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
	})
	if err != nil {
		return nil, err
	}
	return resp.Binding, nil
}

func (c APIClient) ModifyRepoRoleBinding(repo, principal string, roles []string) error {
	_, err := c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
		Resource:  &auth.Resource{Type: auth.ResourceType_REPO, Name: repo},
		Principal: principal,
		Roles:     roles,
	})
	if err != nil {
		return err
	}
	return nil
}

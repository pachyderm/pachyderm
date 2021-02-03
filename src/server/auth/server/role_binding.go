package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
)

type groupLookupFn func(ctx context.Context, subject string) ([]string, error)

type authorizeRequest struct {
	subject          string
	permissions      map[auth.Permission]interface{}
	groupsForSubject groupLookupFn
	groups           []string
}

func newAuthorizeRequest(subject string, permissions []auth.Permission, groupsForSubject groupLookupFn) *authorizeRequest {
	permissionMap := make(map[auth.Permission]interface{})
	for _, p := range permissions {
		permissionMap[p] = true
	}

	return &authorizeRequest{
		subject:          subject,
		permissions:      permissionMap,
		groupsForSubject: groupsForSubject,
	}
}

// satisfied returns true if no permissions remain
func (r *authorizeRequest) satisfied() bool {
	return len(r.permission) == 0
}

// evaluateRoleBinding removes permissions that are satisfied by the role binding from the
// set of desired permissions.
func (r *authorizeRequest) evaluateRoleBinding(binding *auth.RoleBinding) error {
	if err := binding.evaluateRoleBindingForSubject(r.subject, binding); err != nil {
		return err
	}

	if len(r.permissions) == 0 {
		return nil
	}

	if r.groups == nil {
		var err error
		r.groups, err = r.groupsForSubject(r.subject)
		if err != nil {
			return err
		}
	}

	for _, g := range r.groups {
		if err := binding.evaluateBindingForSubject(g, binding); err != nil {
			return err
		}
	}

	return nil
}

func (r *authorizeRequest) evaluateRoleBindingForSubject(subject string, binding *auth.RoleBinding) error {
	if binding.Entries == nil {
		return
	}

	if entry, ok := bindings.Entries[subject]; ok {
		for role := range entry {
			permissions, err := permissionsForRole(role)
			if err != nil {
				return err
			}

			for _, permission := range permissions {
				delete(r.permissions, permission)
			}
		}
	}
	return nil
}

// permissionsForRole returns the set of permissions associated with a role.
// For now this is a hard-coded list but it may be extended to support user-defined roles.
func permissionsForRole(role string) ([]auth.Permission, error) {
	switch role {
	case auth.ClusterAdminRole:
		return []auth.Permission{
			auth.Permission_CLUSTER_ADMIN,
		}
	case auth.RepoOwnerRole:
		return []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_MODIFY_BINDINGS,
		}
	case auth.RepoWriterRole:
		return []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
		}
	case auth.RepoReaderRole:
		return []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
		}
	}
	return nil, fmt.Errorf("unknown role %q", role)
}

package server

import (
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"

	"github.com/pachyderm/pachyderm/v2/src/auth"
)

type groupLookupFn func(txnCtx *txncontext.TransactionContext, subject string) ([]string, error)

// authorizeRequest is a helper struct used to evaluate an incoming Authorize request.
// It's initialized with the subject and set of permissions required for an Operation,
// and it looks at role bindings to figure out which permissions are satisfied.
// This decouples evaluating role bindings from fetching the role bindings,
// so we can easily try the cheapest/most likely options and exit early if
// the permissions are all satisfied.
type authorizeRequest struct {
	subject              string
	permissions          map[auth.Permission]bool
	roleMap              map[string]*auth.Role
	satisfiedPermissions []auth.Permission
	groupsForSubject     groupLookupFn
	groups               []string
}

func newAuthorizeRequest(subject string, permissions map[auth.Permission]bool, groupsForSubject groupLookupFn) *authorizeRequest {
	return &authorizeRequest{
		subject:              subject,
		roleMap:              make(map[string]*auth.Role),
		permissions:          permissions,
		groupsForSubject:     groupsForSubject,
		satisfiedPermissions: make([]auth.Permission, 0),
	}
}

func (r *authorizeRequest) rolesForResourceType(rt auth.ResourceType) []string {
	roles := make([]string, 0, len(r.roleMap))
	for r, def := range r.roleMap {
		if roleReturnedForResource(def, rt) {
			roles = append(roles, r)
		}
	}
	sort.Strings(roles)
	return roles
}

func (r *authorizeRequest) satisfiedForResourceType(rt auth.ResourceType) []auth.Permission {
	roles := r.rolesForResourceType(rt)
	perms := make([]auth.Permission, 0)
	for _, role := range roles {
		perms = append(perms, r.roleMap[role].Permissions...)
	}
	return perms
}

// isSatisfied returns true if no permissions remain
func (r *authorizeRequest) isSatisfied() bool {
	return len(r.permissions) == 0
}

func (r *authorizeRequest) missing() []auth.Permission {
	missing := make([]auth.Permission, 0, len(r.permissions))
	for p := range r.permissions {
		missing = append(missing, p)
	}
	return missing
}

// evaluateRoleBinding removes permissions that are satisfied by the role binding from the
// set of desired permissions. A subject derives permissions from:
// - role bindings that refer to them by name
// - role bindings that refer to allClusterUsers
// - role bindings that refer to any group the subject belongs to
func (r *authorizeRequest) evaluateRoleBinding(txnCtx *txncontext.TransactionContext, binding *auth.RoleBinding) error {
	if err := r.evaluateRoleBindingForSubject(r.subject, binding); err != nil {
		return err
	}

	if len(r.permissions) == 0 {
		return nil
	}

	if err := r.evaluateRoleBindingForSubject(auth.AllClusterUsersSubject, binding); err != nil {
		return err
	}

	// If all permissions are satisfied without checking group membership, then return early
	if len(r.permissions) == 0 {
		return nil
	}

	// Cache the group membership after the first lookup, in case we need to evaluate several
	// bindings to cover the set of permissions.
	if r.groups == nil {
		var err error
		r.groups, err = r.groupsForSubject(txnCtx, r.subject)
		if err != nil {
			return err
		}
	}

	for _, g := range r.groups {
		if err := r.evaluateRoleBindingForSubject(g, binding); err != nil {
			return err
		}
	}

	return nil
}

func (r *authorizeRequest) evaluateRoleBindingForSubject(subject string, binding *auth.RoleBinding) error {
	if binding.Entries == nil {
		return nil
	}

	if entry, ok := binding.Entries[subject]; ok {
		for role := range entry.Roles {
			// Don't look up permissions for a role we already saw in another binding
			if _, ok := r.roleMap[role]; ok {
				continue
			}

			roleDefinition, err := getRole(role)
			if err != nil {
				return err
			}

			r.roleMap[role] = roleDefinition.role

			for _, permission := range roleDefinition.role.Permissions {
				if _, ok := r.permissions[permission]; ok {
					r.satisfiedPermissions = append(r.satisfiedPermissions, permission)
					delete(r.permissions, permission)
				}
			}
		}
	}
	return nil
}

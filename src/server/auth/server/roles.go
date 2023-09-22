package server

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type internalRole struct {
	role             *auth.Role
	permissionsIndex map[auth.Permission]bool
}

var roles = make(map[string]*internalRole)

func registerRole(r *auth.Role) *auth.Role {
	permissionsIndex := make(map[auth.Permission]bool)

	for _, permission := range r.Permissions {
		permissionsIndex[permission] = true
	}

	roles[r.Name] = &internalRole{
		role:             r,
		permissionsIndex: permissionsIndex,
	}
	return r
}

func getRole(name string) (*internalRole, error) {
	r, ok := roles[name]
	if !ok {
		return nil, errors.Errorf("unknown role %q", name)
	}
	return r, nil
}

func init() {
	// repoReader has the ability to view files, commits, branches
	// and create pipelines that read from a repo.
	repoReaderRole := registerRole(&auth.Role{
		Name:         auth.RepoReaderRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT, auth.ResourceType_REPO},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_REPO},
		Permissions: []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_PIPELINE_LIST_JOB,
		},
	})

	// repoWriter has the ability to create and delete commits,
	// write files to a repo and create pipelines that write to a repo,
	// plus all the permissions of repoReader.
	repoWriterRole := registerRole(&auth.Role{
		Name:         auth.RepoWriterRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT, auth.ResourceType_REPO},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_REPO},
		Permissions: combinePermissions(repoReaderRole.Permissions, []auth.Permission{
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
		}),
	})

	// repoOwner has the ability to modify the role bindings for
	// a repo and delete it, plus all the permissions of repoWriter.
	repoOwnerRole := registerRole(&auth.Role{
		Name:         auth.RepoOwnerRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT, auth.ResourceType_REPO},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_REPO},
		Permissions: combinePermissions(repoWriterRole.Permissions, []auth.Permission{
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_DELETE,
		}),
	})

	// oidcAppAdmin has the ability to create, update and
	// delete OIDC apps.
	oidcAppAdminRole := registerRole(&auth.Role{
		Name:         auth.OIDCAppAdminRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_IDENTITY_DELETE_OIDC_CLIENT,
			auth.Permission_CLUSTER_IDENTITY_CREATE_OIDC_CLIENT,
			auth.Permission_CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT,
			auth.Permission_CLUSTER_IDENTITY_LIST_OIDC_CLIENTS,
			auth.Permission_CLUSTER_IDENTITY_GET_OIDC_CLIENT,
		},
	})

	// idpAdmin has the ability to create, update and delete
	// identity providers.
	idpAdminRole := registerRole(&auth.Role{
		Name:         auth.IDPAdminRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_IDENTITY_CREATE_IDP,
			auth.Permission_CLUSTER_IDENTITY_UPDATE_IDP,
			auth.Permission_CLUSTER_IDENTITY_LIST_IDPS,
			auth.Permission_CLUSTER_IDENTITY_GET_IDP,
			auth.Permission_CLUSTER_IDENTITY_DELETE_IDP,
		},
	})

	// identityAdmin has the ability to modify the identity
	// server configuration
	identityAdminRole := registerRole(&auth.Role{
		Name:         auth.IdentityAdminRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_IDENTITY_SET_CONFIG,
			auth.Permission_CLUSTER_IDENTITY_GET_CONFIG,
		},
	})

	// debugger has the ability to produce debug dumps
	debuggerRole := registerRole(&auth.Role{
		Name:         auth.DebuggerRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_DEBUG_DUMP,
			auth.Permission_CLUSTER_GET_PACHD_LOGS,
			auth.Permission_CLUSTER_GET_LOKI_LOGS,
		},
	})

	// lokiLogReader has the ability to read logs from Loki
	lokiLogReaderRole := registerRole(&auth.Role{
		Name:         auth.LokiLogReaderRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_GET_LOKI_LOGS,
		},
	})

	// robotUser has the ability to create tokens for any robot user
	robotUserRole := registerRole(&auth.Role{
		Name:         auth.RobotUserRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_AUTH_GET_ROBOT_TOKEN,
		},
	})

	// licenseAdmin has the ability to update the enterprise license and manage clusters
	licenseAdminRole := registerRole(&auth.Role{
		Name:         auth.LicenseAdminRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_LICENSE_ACTIVATE,
			auth.Permission_CLUSTER_LICENSE_GET_CODE,
			auth.Permission_CLUSTER_LICENSE_ADD_CLUSTER,
			auth.Permission_CLUSTER_LICENSE_UPDATE_CLUSTER,
			auth.Permission_CLUSTER_LICENSE_DELETE_CLUSTER,
			auth.Permission_CLUSTER_LICENSE_LIST_CLUSTERS,
		},
	})

	// secretAdmin has the ability to list, create and delete secrets in k8s
	secretAdminRole := registerRole(&auth.Role{
		Name:         auth.SecretAdminRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_CREATE_SECRET,
			auth.Permission_CLUSTER_LIST_SECRETS,
			auth.Permission_SECRET_INSPECT,
			auth.Permission_SECRET_DELETE,
		},
	})

	pachdLogReaderRole := registerRole(&auth.Role{
		Name:         auth.PachdLogReaderRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_CLUSTER_GET_PACHD_LOGS,
		},
	})

	// Project related roles
	projectViewerRole := registerRole(&auth.Role{
		Name:         auth.ProjectViewerRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_PROJECT},
		Permissions: []auth.Permission{
			auth.Permission_PROJECT_LIST_REPO,
		},
	})

	projectWriterRole := registerRole(&auth.Role{
		Name:         auth.ProjectWriterRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_PROJECT},
		Permissions: combinePermissions(projectViewerRole.Permissions, []auth.Permission{
			auth.Permission_PROJECT_CREATE_REPO,
		}),
	})

	projectOwnerRole := registerRole(&auth.Role{
		Name:         auth.ProjectOwnerRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_PROJECT, auth.ResourceType_REPO},
		Permissions: combinePermissions(repoOwnerRole.Permissions, projectWriterRole.Permissions, []auth.Permission{
			auth.Permission_PROJECT_DELETE,
			auth.Permission_PROJECT_MODIFY_BINDINGS,
		}),
	})

	projectCreatorRole := registerRole(&auth.Role{
		Name:         auth.ProjectCreatorRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER},
		Permissions: []auth.Permission{
			auth.Permission_PROJECT_CREATE,
		},
	})

	// clusterAdmin is a catch-all role that has every permission
	registerRole(&auth.Role{
		Name:         auth.ClusterAdminRole,
		CanBeBoundTo: []auth.ResourceType{auth.ResourceType_CLUSTER},
		ReturnedFor:  []auth.ResourceType{auth.ResourceType_CLUSTER, auth.ResourceType_PROJECT, auth.ResourceType_REPO},
		Permissions: combinePermissions(
			repoOwnerRole.Permissions,
			oidcAppAdminRole.Permissions,
			idpAdminRole.Permissions,
			identityAdminRole.Permissions,
			debuggerRole.Permissions,
			lokiLogReaderRole.Permissions,
			robotUserRole.Permissions,
			licenseAdminRole.Permissions,
			secretAdminRole.Permissions,
			pachdLogReaderRole.Permissions,
			projectOwnerRole.Permissions,
			projectCreatorRole.Permissions,
			[]auth.Permission{
				auth.Permission_CLUSTER_MODIFY_BINDINGS,
				auth.Permission_CLUSTER_GET_BINDINGS,
				auth.Permission_CLUSTER_AUTH_ACTIVATE,
				auth.Permission_CLUSTER_AUTH_DEACTIVATE,
				auth.Permission_CLUSTER_AUTH_GET_CONFIG,
				auth.Permission_CLUSTER_AUTH_SET_CONFIG,
				auth.Permission_CLUSTER_AUTH_MODIFY_GROUP_MEMBERS,
				auth.Permission_CLUSTER_AUTH_GET_GROUPS,
				auth.Permission_CLUSTER_AUTH_GET_GROUP_USERS,
				auth.Permission_CLUSTER_AUTH_EXTRACT_TOKENS,
				auth.Permission_CLUSTER_AUTH_RESTORE_TOKEN,
				auth.Permission_CLUSTER_AUTH_ROTATE_ROOT_TOKEN,
				auth.Permission_CLUSTER_AUTH_DELETE_EXPIRED_TOKENS,
				auth.Permission_CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL,
				auth.Permission_CLUSTER_AUTH_REVOKE_USER_TOKENS,
				auth.Permission_CLUSTER_ENTERPRISE_ACTIVATE,
				auth.Permission_CLUSTER_ENTERPRISE_HEARTBEAT,
				auth.Permission_CLUSTER_ENTERPRISE_GET_CODE,
				auth.Permission_CLUSTER_ENTERPRISE_DEACTIVATE,
				auth.Permission_CLUSTER_DELETE_ALL,
				auth.Permission_CLUSTER_ENTERPRISE_PAUSE,
				auth.Permission_CLUSTER_SET_DEFAULTS,
			}),
	})
}

func combinePermissions(permissionSets ...[]auth.Permission) []auth.Permission {
	seen := make(map[auth.Permission]bool)
	output := make([]auth.Permission, 0)
	for _, permissions := range permissionSets {
		for _, permission := range permissions {
			if seen[permission] {
				continue
			}
			seen[permission] = true
			output = append(output, permission)
		}
	}
	return output
}

func roleReturnedForResource(r *auth.Role, rt auth.ResourceType) bool {
	for _, t := range r.ReturnedFor {
		if t == rt {
			return true
		}
	}
	return false
}

func rolesForPermission(permission auth.Permission) []*auth.Role {
	resp := make([]*auth.Role, 0)
	for _, r := range roles {
		if _, ok := r.permissionsIndex[permission]; ok {
			resp = append(resp, r.role)
		}
	}
	return resp
}

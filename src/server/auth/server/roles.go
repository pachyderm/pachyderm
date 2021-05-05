package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/auth"
)

var (
	// repoReader has the ability to view files, commits, branches
	// and create pipelines that read from a repo.
	repoReaderRole = []auth.Permission{
		auth.Permission_REPO_READ,
		auth.Permission_REPO_INSPECT_COMMIT,
		auth.Permission_REPO_LIST_COMMIT,
		auth.Permission_REPO_LIST_BRANCH,
		auth.Permission_REPO_LIST_FILE,
		auth.Permission_REPO_INSPECT_FILE,
		auth.Permission_REPO_ADD_PIPELINE_READER,
		auth.Permission_REPO_REMOVE_PIPELINE_READER,
		auth.Permission_PIPELINE_LIST_JOB,
	}

	// repoWriter has the ability to create and delete commits,
	// write files to a repo and create pipelines that write to a repo,
	// plus all the permissions of repoReader.
	repoWriterRole = combinePermissions(repoReaderRole, []auth.Permission{
		auth.Permission_REPO_WRITE,
		auth.Permission_REPO_DELETE_COMMIT,
		auth.Permission_REPO_CREATE_BRANCH,
		auth.Permission_REPO_DELETE_BRANCH,
		auth.Permission_REPO_ADD_PIPELINE_WRITER,
	})

	// repoWriter has the ability to modify the role bindings for
	// a repo and delete it, plus all the permissions of repoWriter.
	repoOwnerRole = combinePermissions(repoWriterRole, []auth.Permission{
		auth.Permission_REPO_MODIFY_BINDINGS,
		auth.Permission_REPO_DELETE,
	})

	// oidcAppAdmin has the ability to create, update and
	// delete OIDC apps.
	oidcAppAdminRole = []auth.Permission{
		auth.Permission_CLUSTER_IDENTITY_DELETE_OIDC_CLIENT,
		auth.Permission_CLUSTER_IDENTITY_CREATE_OIDC_CLIENT,
		auth.Permission_CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT,
		auth.Permission_CLUSTER_IDENTITY_LIST_OIDC_CLIENTS,
		auth.Permission_CLUSTER_IDENTITY_GET_OIDC_CLIENT,
	}

	// idpAdmin has the ability to create, update and delete
	// identity providers.
	idpAdminRole = []auth.Permission{
		auth.Permission_CLUSTER_IDENTITY_CREATE_IDP,
		auth.Permission_CLUSTER_IDENTITY_UPDATE_IDP,
		auth.Permission_CLUSTER_IDENTITY_LIST_IDPS,
		auth.Permission_CLUSTER_IDENTITY_GET_IDP,
		auth.Permission_CLUSTER_IDENTITY_DELETE_IDP,
	}

	// identityAdmin has the ability to modify the identity
	// server configuration
	identityAdminRole = []auth.Permission{
		auth.Permission_CLUSTER_IDENTITY_SET_CONFIG,
		auth.Permission_CLUSTER_IDENTITY_GET_CONFIG,
	}

	// debugger has the ability to produce debug dumps
	debuggerRole = []auth.Permission{
		auth.Permission_CLUSTER_DEBUG_DUMP,
	}

	// robotUser has the ability to create tokens for any robot user
	robotUserRole = []auth.Permission{
		auth.Permission_CLUSTER_AUTH_GET_ROBOT_TOKEN,
	}

	// licenseAdmin has the ability to update the enterprise license and manage clusters
	licenseAdminRole = []auth.Permission{
		auth.Permission_CLUSTER_LICENSE_ACTIVATE,
		auth.Permission_CLUSTER_LICENSE_GET_CODE,
		auth.Permission_CLUSTER_LICENSE_ADD_CLUSTER,
		auth.Permission_CLUSTER_LICENSE_UPDATE_CLUSTER,
		auth.Permission_CLUSTER_LICENSE_DELETE_CLUSTER,
		auth.Permission_CLUSTER_LICENSE_LIST_CLUSTERS,
	}

	// secretAdmin has the ability to list, create and delete secrets in k8s
	secretAdminRole = []auth.Permission{
		auth.Permission_CLUSTER_CREATE_SECRET,
		auth.Permission_CLUSTER_LIST_SECRETS,
		auth.Permission_SECRET_INSPECT,
		auth.Permission_SECRET_DELETE,
	}

	// clusterAdmin is a catch-all role that has every permission
	clusterAdminRole = combinePermissions(
		repoOwnerRole,
		oidcAppAdminRole,
		idpAdminRole,
		identityAdminRole,
		debuggerRole,
		robotUserRole,
		licenseAdminRole,
		secretAdminRole,
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
		})
)

// permissionsForRole returns the set of permissions associated with a role.
// For now this is a hard-coded list but it may be extended to support user-defined roles.
func permissionsForRole(role string) ([]auth.Permission, error) {
	switch role {
	case auth.ClusterAdminRole:
		return clusterAdminRole, nil
	case auth.RepoOwnerRole:
		return repoOwnerRole, nil
	case auth.RepoWriterRole:
		return repoWriterRole, nil
	case auth.RepoReaderRole:
		return repoReaderRole, nil
	case auth.OIDCAppAdminRole:
		return oidcAppAdminRole, nil
	case auth.IDPAdminRole:
		return idpAdminRole, nil
	case auth.IdentityAdminRole:
		return identityAdminRole, nil
	case auth.DebuggerRole:
		return debuggerRole, nil
	case auth.RobotUserRole:
		return robotUserRole, nil
	case auth.LicenseAdminRole:
		return licenseAdminRole, nil
	}
	return nil, fmt.Errorf("unknown role %q", role)
}

func combinePermissions(permissions ...[]auth.Permission) []auth.Permission {
	output := make([]auth.Permission, 0)
	for _, p := range permissions {
		output = append(output, p...)
	}
	return output
}

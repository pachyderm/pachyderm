package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/auth"
)

// permissionsForRole returns the set of permissions associated with a role.
// For now this is a hard-coded list but it may be extended to support user-defined roles.
func permissionsForRole(role string) ([]auth.Permission, error) {
	switch role {
	case auth.ClusterAdminRole:
		return []auth.Permission{
			auth.Permission_CLUSTER_MODIFY_BINDINGS,
			auth.Permission_CLUSTER_GET_BINDINGS,
			auth.Permission_CLUSTER_AUTH_ACTIVATE,
			auth.Permission_CLUSTER_AUTH_DEACTIVATE,
			auth.Permission_CLUSTER_AUTH_GET_CONFIG,
			auth.Permission_CLUSTER_AUTH_SET_CONFIG,
			auth.Permission_CLUSTER_AUTH_GET_ROBOT_TOKEN,
			auth.Permission_CLUSTER_AUTH_MODIFY_GROUP_MEMBERS,
			auth.Permission_CLUSTER_AUTH_GET_GROUPS,
			auth.Permission_CLUSTER_AUTH_GET_GROUP_USERS,
			auth.Permission_CLUSTER_AUTH_EXTRACT_TOKENS,
			auth.Permission_CLUSTER_AUTH_RESTORE_TOKEN,
			auth.Permission_CLUSTER_ENTERPRISE_ACTIVATE,
			auth.Permission_CLUSTER_ENTERPRISE_HEARTBEAT,
			auth.Permission_CLUSTER_ENTERPRISE_GET_CODE,
			auth.Permission_CLUSTER_ENTERPRISE_DEACTIVATE,
			auth.Permission_CLUSTER_IDENTITY_SET_CONFIG,
			auth.Permission_CLUSTER_IDENTITY_GET_CONFIG,
			auth.Permission_CLUSTER_IDENTITY_CREATE_IDP,
			auth.Permission_CLUSTER_IDENTITY_UPDATE_IDP,
			auth.Permission_CLUSTER_IDENTITY_LIST_IDPS,
			auth.Permission_CLUSTER_IDENTITY_GET_IDP,
			auth.Permission_CLUSTER_IDENTITY_DELETE_IDP,
			auth.Permission_CLUSTER_IDENTITY_DELETE_OIDC_CLIENT,
			auth.Permission_CLUSTER_IDENTITY_CREATE_OIDC_CLIENT,
			auth.Permission_CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT,
			auth.Permission_CLUSTER_IDENTITY_LIST_OIDC_CLIENTS,
			auth.Permission_CLUSTER_IDENTITY_GET_OIDC_CLIENT,
			auth.Permission_CLUSTER_DEBUG_DUMP,
			auth.Permission_CLUSTER_LICENSE_ACTIVATE,
			auth.Permission_CLUSTER_LICENSE_GET_CODE,
			auth.Permission_CLUSTER_LICENSE_ADD_CLUSTER,
			auth.Permission_CLUSTER_LICENSE_UPDATE_CLUSTER,
			auth.Permission_CLUSTER_LICENSE_DELETE_CLUSTER,
			auth.Permission_CLUSTER_LICENSE_LIST_CLUSTERS,
			auth.Permission_CLUSTER_DELETE_ALL,
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_DELETE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_PIPELINE_LIST_JOB,
		}, nil
	case auth.RepoOwnerRole:
		return []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_DELETE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_PIPELINE_LIST_JOB,
		}, nil
	case auth.RepoWriterRole:
		return []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
			auth.Permission_PIPELINE_LIST_JOB,
		}, nil
	case auth.RepoReaderRole:
		return []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_PIPELINE_LIST_JOB,
		}, nil
	}
	return nil, fmt.Errorf("unknown role %q", role)
}

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
			auth.Permission_CLUSTER_ADMIN,
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

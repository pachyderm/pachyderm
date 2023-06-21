import {Permission, ResourceType} from '@graphqlTypes';

import useCommit from '@dash-frontend/hooks/useCommit';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';

const useRepoDetails = () => {
  const {repoId, projectId, commitId} = useUrlState();
  const {loading: repoLoading, repo, error: repoError} = useCurrentRepo();

  const {
    commit,
    loading: commitLoading,
    error: commitError,
  } = useCommit({
    args: {
      projectId,
      repoName: repoId,
      id: commitId ? commitId : '',
      withDiff: true,
    },
  });

  const {isAuthorizedAction: editRolesPermission} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_MODIFY_BINDINGS],
    resource: {type: ResourceType.REPO, name: `${projectId}/${repoId}`},
  });

  const currentRepoLoading =
    repoLoading || commitLoading || repoId !== repo?.id;

  return {
    projectId,
    repoId,
    repo,
    commit,
    currentRepoLoading,
    commitError,
    repoError,
    editRolesPermission,
  };
};

export default useRepoDetails;

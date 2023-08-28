import {Permission, ResourceType} from '@graphqlTypes';

import {useCommitDiffQuery} from '@dash-frontend/generated/hooks';
import useCommit from '@dash-frontend/hooks/useCommit';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';

const useRepoDetails = () => {
  const {repoId, projectId, commitId} = useUrlState();
  const {loading: repoLoading, repo, error: repoError} = useCurrentRepo();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const {
    commit,
    loading: commitLoading,
    error: commitError,
  } = useCommit({
    args: {
      projectId,
      repoName: repoId,
      id: commitId ? commitId : '',
    },
  });
  const {
    data: commitDiff,
    loading: diffLoading,
    error: diffError,
  } = useCommitDiffQuery({
    variables: {
      args: {
        projectId,
        commitId: commit?.id,
        branchName: commit?.branch?.name,
        repoName: commit?.repoName,
      },
    },
    skip: !commit || commit.finished === -1,
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
    commitDiff,
    diffLoading,
    diffError,
    repoError,
    editRolesPermission,
    getPathToFileBrowser,
  };
};

export default useRepoDetails;

import {RepoInfo} from '@dash-frontend/api/pfs';
import {
  Permission,
  ResourceType,
  useAuthorizeLazy,
} from '@dash-frontend/hooks/useAuthorize';
import {useCommits} from '@dash-frontend/hooks/useCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useModal} from '@pachyderm/components';

const useRepoListRow = (repo: RepoInfo) => {
  const {projectId} = useUrlState();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  const {commits, loading: lastCommitLoading} = useCommits({
    projectName: projectId,
    repoName: repo.repo?.name || '',
    args: {
      number: 1,
    },
  });

  const lastCommit = commits && commits[0];

  const {checkPermissions, hasAllPermissions: hasRepoEditRoles} =
    useAuthorizeLazy({
      permissions: [Permission.REPO_MODIFY_BINDINGS],
      resource: {
        type: ResourceType.REPO,
        name: `${projectId}/${repo?.repo?.name}`,
      },
    });

  return {
    projectId,
    rolesModalOpen,
    openRolesModal,
    closeRolesModal,
    hasRepoEditRoles,
    lastCommit,
    lastCommitLoading,
    checkPermissions,
  };
};

export default useRepoListRow;

import {
  ReposWithCommitQuery,
  Commit,
  Permission,
  ResourceType,
} from '@graphqlTypes';
import {useHistory} from 'react-router';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorizationLazy} from '@dash-frontend/hooks/useVerifiedAuthorizationLazy';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem, useModal} from '@pachyderm/components';

const useRepoListRow = (repoId: string) => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  const {
    checkRolesPermission,
    isAuthorizedAction: editRolesPermission,
    isAuthActive,
  } = useVerifiedAuthorizationLazy({
    permissionsList: [Permission.REPO_MODIFY_BINDINGS],
    resource: {type: ResourceType.REPO, name: `${projectId}/${repoId}`},
  });

  const inspectCommitRedirect = (commit: Commit | null | undefined) => {
    if (commit) {
      const fileBrowserLink = getPathToFileBrowser(
        {
          projectId,
          repoId: commit.repoName,
          commitId: commit.id,
          branchId: commit.branch?.name || 'default',
        },
        true,
      );
      browserHistory.push(fileBrowserLink);
    }
  };

  const onOverflowMenuSelect =
    (repo: ReposWithCommitQuery['repos'][number]) => (id: string) => {
      switch (id) {
        case 'dag':
          return browserHistory.push(
            repoRoute({projectId, repoId: repo?.id || ''}),
          );
        case 'inspect-commits':
          return inspectCommitRedirect(repo?.lastCommit);
        case 'set-roles':
          return openRolesModal();
        default:
          return null;
      }
    };

  const generateIconItems = (commitId?: string): DropdownItem[] => {
    const items: DropdownItem[] = [
      {
        id: 'inspect-commits',
        content: 'Inspect commits',
        closeOnClick: true,
        disabled: commitId === undefined,
      },
      {
        id: 'dag',
        content: 'View in DAG',
        closeOnClick: true,
      },
    ];

    if (isAuthActive) {
      items.push({
        id: 'set-roles',
        content: editRolesPermission ? 'Set Roles' : 'See All Roles',
        closeOnClick: true,
        topBorder: true,
      });
    }
    return items;
  };

  return {
    checkRolesPermission,
    projectId,
    generateIconItems,
    onOverflowMenuSelect,
    rolesModalOpen,
    closeRolesModal,
    editRolesPermission,
  };
};

export default useRepoListRow;

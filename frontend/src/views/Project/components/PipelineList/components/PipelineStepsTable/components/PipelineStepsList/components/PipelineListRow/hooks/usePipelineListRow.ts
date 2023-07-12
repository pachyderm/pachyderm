import {Permission, ResourceType} from '@graphqlTypes';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorizationLazy} from '@dash-frontend/hooks/useVerifiedAuthorizationLazy';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem, useModal} from '@pachyderm/components';

const usePipelineListRow = (pipelineName: string) => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
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
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineName}`},
  });

  const onOverflowMenuSelect = (pipelineId: string) => (id: string) => {
    switch (id) {
      case 'dag':
        return browserHistory.push(pipelineRoute({projectId, pipelineId}));
      case 'set-roles':
        return openRolesModal();
      default:
        return null;
    }
  };

  const iconItems: DropdownItem[] = [
    {
      id: 'dag',
      content: 'View in DAG',
      closeOnClick: true,
    },
  ];

  if (isAuthActive) {
    iconItems.push({
      id: 'set-roles',
      content: editRolesPermission
        ? 'Set Roles via Repo'
        : 'See All Roles via Repo',
      closeOnClick: true,
      topBorder: true,
    });
  }

  return {
    checkRolesPermission,
    projectId,
    iconItems,
    onOverflowMenuSelect,
    rolesModalOpen,
    closeRolesModal,
    editRolesPermission,
    isAuthActive,
  };
};

export default usePipelineListRow;

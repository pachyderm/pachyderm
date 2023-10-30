import {Permission, PipelineType, ResourceType} from '@graphqlTypes';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorizationLazy} from '@dash-frontend/hooks/useVerifiedAuthorizationLazy';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem, useModal} from '@pachyderm/components';

const usePipelineListRow = (pipelineName: string, type?: PipelineType) => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  const {
    openModal: openRerunPipelineModal,
    closeModal: closeRerunPipelineModal,
    isOpen: rerunPipelineModalOpen,
  } = useModal(false);

  const {checkRolesPermission, satisfiedList, isAuthActive} =
    useVerifiedAuthorizationLazy({
      permissionsList: [Permission.REPO_MODIFY_BINDINGS, Permission.REPO_WRITE], // I don't think there is a pipeline edit permission directly TODO
      resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineName}`},
    });

  const hasRepoWrite =
    satisfiedList && satisfiedList.indexOf(Permission.REPO_WRITE) !== -1;

  const editRolesPermission =
    satisfiedList &&
    satisfiedList.indexOf(Permission.REPO_MODIFY_BINDINGS) !== -1;

  const onOverflowMenuSelect = (pipelineId: string) => (id: string) => {
    switch (id) {
      case 'dag':
        return browserHistory.push(pipelineRoute({projectId, pipelineId}));
      case 'set-roles':
        return openRolesModal();
      case 'rerun-pipeline':
        return openRerunPipelineModal();
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
    {
      id: 'rerun-pipeline',
      content: 'Rerun Pipeline',
      closeOnClick: true,
      disabled: type === PipelineType.SPOUT || (isAuthActive && !hasRepoWrite),
    },
  ];

  if (isAuthActive) {
    iconItems.push({
      id: 'set-roles',
      content: editRolesPermission
        ? 'Set Roles via Repo'
        : 'See All Roles via Repo',
      closeOnClick: true,
    });
  }

  return {
    checkRolesPermission,
    projectId,
    iconItems,
    onOverflowMenuSelect,
    rolesModalOpen,
    closeRolesModal,
    closeRerunPipelineModal,
    rerunPipelineModalOpen,
    editRolesPermission,
    isAuthActive,
  };
};

export default usePipelineListRow;

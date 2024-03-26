import {ResourceType} from '@dash-frontend/api/auth';
import {Permission, useAuthorizeLazy} from '@dash-frontend/hooks/useAuthorize';
import {useRepo} from '@dash-frontend/hooks/useRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useModal} from '@pachyderm/components';

const usePipelineListRow = (pipelineName: string) => {
  const {projectId} = useUrlState();
  const {
    openModal: openRolesModal,
    closeModal: closeRolesModal,
    isOpen: rolesModalOpen,
  } = useModal(false);

  const {repo, loading: repoLoading} = useRepo({
    repo: {
      name: pipelineName,
      project: {
        name: projectId,
      },
    },
  });

  const {checkPermissions, hasRepoModifyBindings: hasRepoEditRoles} =
    useAuthorizeLazy({
      permissions: [Permission.REPO_MODIFY_BINDINGS],
      resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineName}`},
    });

  return {
    projectId,
    rolesModalOpen,
    openRolesModal,
    closeRolesModal,
    hasRepoEditRoles,
    checkPermissions,
    repo,
    repoLoading,
  };
};

export default usePipelineListRow;

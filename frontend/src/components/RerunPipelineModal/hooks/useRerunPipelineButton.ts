import {Permission, ResourceType} from '@graphqlTypes';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';
import {useModal} from '@pachyderm/components';

const useRerunPipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const {openModal, closeModal, isOpen} = useModal(false);
  const {isSpout} = useCurrentPipeline();

  const {isAuthorizedAction: hasPermission} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_WRITE], // I don't think there is a pipeline edit permission directly TODO
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const tooltipText = () => {
    if (isSpout) {
      return 'Spout pipelines cannot be rerun.';
    } else if (!hasPermission) {
      return 'You need at least repoWriter to rerun this pipeline.';
    }
    return 'Rerun Pipeline';
  };

  return {
    pipelineId,
    openModal,
    closeModal,
    isOpen,
    disable: !hasPermission || isSpout,
    tooltipText,
  };
};

export default useRerunPipelineButton;

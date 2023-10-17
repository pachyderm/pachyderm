import {Permission, ResourceType} from '@graphqlTypes';
import {useCallback} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';
import {updatePipelineRoute} from '@dash-frontend/views/Project/utils/routes';

const useUpdatePipelineButton = () => {
  const browserHistory = useHistory();
  const {projectId, pipelineId} = useUrlState();
  const {isAuthorizedAction: hasPipelineUpdate} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_WRITE], // I don't think there is a pipeline edit permission directly TODO
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const disableButton = !hasPipelineUpdate;

  const tooltipText = () => {
    if (!hasPipelineUpdate)
      return 'You need at least repoWriter to update this.';
    return 'Update Pipeline';
  };

  const handleClick = useCallback(() => {
    browserHistory.push(updatePipelineRoute({projectId, pipelineId}));
  }, [browserHistory, pipelineId, projectId]);

  return {
    disableButton,
    tooltipText: tooltipText(),
    handleClick,
  };
};

export default useUpdatePipelineButton;

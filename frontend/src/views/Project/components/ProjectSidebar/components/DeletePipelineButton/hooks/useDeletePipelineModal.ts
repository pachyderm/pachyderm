import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import {useDeletePipeline} from '@dash-frontend/hooks/useDeletePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {PROJECT_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  lineageRoute,
  projectPipelinesRoute,
} from '@dash-frontend/views/Project/utils/routes';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const browserHistory = useHistory();
  const projectMatch = !!useRouteMatch(PROJECT_PATH);

  const onCompleted = () => {
    if (projectMatch) {
      browserHistory.push(projectPipelinesRoute({projectId}));
    } else {
      browserHistory.push(lineageRoute({projectId}));
    }
  };

  const {
    deletePipeline,
    loading: updating,
    error,
  } = useDeletePipeline(pipelineId, onCompleted);

  const onDelete = useCallback(() => {
    if (pipelineId) {
      deletePipeline({
        variables: {
          args: {
            name: pipelineId,
            projectId,
          },
        },
      });
    }
  }, [pipelineId, projectId, deletePipeline]);

  return {
    onDelete,
    updating,
    error,
  };
};

export default useDeletePipelineButton;

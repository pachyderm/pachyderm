import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useDeletePipeline from '@dash-frontend/hooks/useDeletePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {PROJECT_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  lineageRoute,
  projectPipelinesRoute,
} from '@dash-frontend/views/Project/utils/routes';

const useDeletePipelineButton = (pipelineId: string) => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const projectMatch = !!useRouteMatch(PROJECT_PATH);

  const {
    deletePipeline,
    loading: updating,
    error,
  } = useDeletePipeline({
    onSuccess: () => {
      if (projectMatch) {
        browserHistory.push(projectPipelinesRoute({projectId}));
      } else {
        browserHistory.push(lineageRoute({projectId}, false));
      }
    },
  });

  const onDelete = useCallback(async () => {
    if (pipelineId) {
      deletePipeline({
        pipeline: {
          project: {
            name: projectId,
          },
          name: pipelineId,
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

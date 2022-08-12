import {useCallback} from 'react';

import {useDeletePipeline} from '@dash-frontend/hooks/useDeletePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const {
    deletePipeline,
    loading: updating,
    error,
  } = useDeletePipeline(pipelineId);

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

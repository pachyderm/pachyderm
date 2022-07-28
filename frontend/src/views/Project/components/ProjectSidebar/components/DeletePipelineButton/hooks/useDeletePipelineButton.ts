import {useCallback, useMemo, useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import {useDeletePipeline} from '@dash-frontend/hooks/useDeletePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});
  const {
    deletePipeline,
    loading: updating,
    error,
  } = useDeletePipeline(pipelineId, () => setModalOpen(false));

  const canDelete = useMemo(() => {
    return (
      dagData &&
      !dagData?.dag?.some(
        ({parents, type}) =>
          type === 'PIPELINE' &&
          parents.some((parent) => parent === pipelineId),
      )
    );
  }, [pipelineId, dagData]);

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
    canDelete,
    modalOpen,
    setModalOpen,
    onDelete,
    updating,
    error,
  };
};

export default useDeletePipelineButton;

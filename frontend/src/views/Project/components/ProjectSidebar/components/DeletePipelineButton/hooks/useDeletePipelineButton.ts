import {useMemo, useState} from 'react';

import {
  useGetDagQuery,
  useDeletePipelineMutation,
} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});
  const [deletePipeline, {loading: updating}] = useDeletePipelineMutation({
    onCompleted: () => setModalOpen(false),
  });

  const [modalOpen, setModalOpen] = useState(false);

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

  const onDelete = () => {
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
  };

  return {
    canDelete,
    modalOpen,
    setModalOpen,
    onDelete,
    updating,
  };
};

export default useDeletePipelineButton;

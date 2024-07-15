import {useMemo, useState} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerticesLazy} from '@dash-frontend/hooks/useVertices';
import {NodeType} from '@dash-frontend/lib/types';

const useDeletePipelineButton = (pipelineId: string) => {
  const {projectId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {getVertices, vertices, loading: verticesLoading} = useVerticesLazy();

  const canDelete = useMemo(() => {
    return !vertices?.some(
      ({parents, type}) =>
        type === NodeType.PIPELINE &&
        parents.some(
          ({project, name}) => project === projectId && name === pipelineId,
        ),
    );
  }, [vertices, projectId, pipelineId]);

  return {
    getVertices,
    modalOpen,
    setModalOpen,
    verticesLoading,
    canDelete,
  };
};

export default useDeletePipelineButton;

import {Vertex} from '@graphqlTypes';
import {useCallback, useMemo, useState} from 'react';

import {
  useGetDagQuery,
  useDeletePipelineMutation,
} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});
  const [deletePipeline, {loading: updating}] = useDeletePipelineMutation({
    refetchQueries: [{query: GET_DAG_QUERY, variables: {args: {projectId}}}],
    awaitRefetchQueries: true,
    update(cache, {data}) {
      if (data?.deletePipeline) {
        cache.modify({
          fields: {
            dag(existingVertices: Vertex[]) {
              return existingVertices.filter(
                (vertex) =>
                  vertex.name !== pipelineId &&
                  vertex.name !== `${pipelineId}_repo`,
              );
            },
          },
        });
        setModalOpen(false);
      }
    },
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
  };
};

export default useDeletePipelineButton;

import {Vertex} from '@graphqlTypes';

import {useDeletePipelineMutation} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';

export const useDeletePipeline = (
  pipelineId: string,
  onCompleted?: () => void,
) => {
  const {projectId} = useUrlState();
  const [deletePipeline, {loading, error}] = useDeletePipelineMutation({
    refetchQueries: [{query: GET_DAG_QUERY, variables: {args: {projectId}}}],
    awaitRefetchQueries: true,
    onCompleted: onCompleted,
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
      }
    },
  });

  return {
    deletePipeline,
    error,
    loading,
  };
};

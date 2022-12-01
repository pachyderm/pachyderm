import {Vertex} from '@graphqlTypes';

import {useDeleteRepoMutation} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';

export const useDeleteRepo = (repoId: string, onCompleted?: () => void) => {
  const {projectId} = useUrlState();
  const [deleteRepo, {loading, error}] = useDeleteRepoMutation({
    refetchQueries: [{query: GET_DAG_QUERY, variables: {args: {projectId}}}],
    awaitRefetchQueries: true,
    onCompleted: onCompleted,
    update(cache, {data}) {
      if (data?.deleteRepo) {
        cache.modify({
          fields: {
            dag(existingVertices: Vertex[]) {
              return existingVertices.filter(
                (vertex) => vertex.name !== `${repoId}_repo`,
              );
            },
          },
        });
      }
    },
  });

  return {
    deleteRepo,
    error,
    loading,
  };
};

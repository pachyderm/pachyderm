import {Vertex} from '@graphqlTypes';
import {useMemo, useState} from 'react';

import {
  useGetDagQuery,
  useDeleteRepoMutation,
} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {GET_DAG_QUERY} from '@dash-frontend/queries/GetDagQuery';

const useDeleteRepoButton = () => {
  const {projectId, repoId} = useUrlState();
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});
  const [deleteRepo, {loading: updating}] = useDeleteRepoMutation({
    refetchQueries: [{query: GET_DAG_QUERY, variables: {args: {projectId}}}],
    awaitRefetchQueries: true,
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
        setModalOpen(false);
      }
    },
  });

  const [modalOpen, setModalOpen] = useState(false);

  const canDelete = useMemo(() => {
    return (
      dagData &&
      !dagData?.dag?.some(({parents}) =>
        parents.some((parent) => parent === repoId),
      )
    );
  }, [repoId, dagData]);

  const onDelete = () => {
    if (repoId) {
      deleteRepo({
        variables: {
          args: {
            repo: {
              name: repoId,
            },
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

export default useDeleteRepoButton;

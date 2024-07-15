import {useState, useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerticesLazy} from '@dash-frontend/hooks/useVertices';
import {NodeType} from '@dash-frontend/lib/types';

const useDeleteRepoButton = (repoId: string) => {
  const {projectId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {getVertices, vertices, loading: verticesLoading} = useVerticesLazy();
  const canDelete = useMemo(
    () =>
      !vertices?.some(({parents, project, name, type}) => {
        const isDescendant = parents.some(
          ({project: parentProject, name: parentName}) =>
            parentProject === projectId && parentName === repoId,
        );
        const isOutputRepo =
          project === projectId &&
          name === repoId &&
          type === NodeType.PIPELINE;

        return isDescendant || isOutputRepo;
      }),
    [projectId, repoId, vertices],
  );

  return {
    getVertices,
    modalOpen,
    setModalOpen,
    verticesLoading,
    canDelete,
  };
};

export default useDeleteRepoButton;

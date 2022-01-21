import {useMemo, useState, useEffect} from 'react';

import {
  useGetDagQuery,
  useDeleteRepoMutation,
} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeleteRepoButton = () => {
  const {projectId, repoId} = useUrlState();
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});
  const [deleteRepo, {loading: updating, data}] = useDeleteRepoMutation();

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

  useEffect(() => {
    if (data?.deleteRepo) {
      setModalOpen(false);
    }
  }, [data?.deleteRepo, setModalOpen]);

  return {
    canDelete,
    modalOpen,
    setModalOpen,
    onDelete,
    updating,
  };
};

export default useDeleteRepoButton;

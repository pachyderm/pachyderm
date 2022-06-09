import {useMemo, useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import {useDeleteRepo} from '@dash-frontend/hooks/useDeleteRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeleteRepoButton = () => {
  const {projectId, repoId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});
  const {deleteRepo, loading: updating} = useDeleteRepo(repoId, () =>
    setModalOpen(false),
  );

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

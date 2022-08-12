import {useMemo, useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeleteRepoButton = () => {
  const {projectId, repoId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});

  const canDelete = useMemo(() => {
    return (
      dagData &&
      !dagData?.dag?.some(({parents}) =>
        parents.some((parent) => parent === repoId),
      )
    );
  }, [repoId, dagData]);

  return {
    canDelete,
    modalOpen,
    setModalOpen,
  };
};

export default useDeleteRepoButton;

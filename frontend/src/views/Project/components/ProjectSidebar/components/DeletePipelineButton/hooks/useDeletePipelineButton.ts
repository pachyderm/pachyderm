import {useMemo, useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: dagData} = useGetDagQuery({variables: {args: {projectId}}});

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

  return {
    canDelete,
    modalOpen,
    setModalOpen,
  };
};

export default useDeletePipelineButton;

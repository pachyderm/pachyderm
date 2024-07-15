import React from 'react';

import {BasicModal} from '@pachyderm/components';

import useDeletePipelineModal from './hooks/useDeletePipelineModal';

const DeletePipelineButton: React.FC<{
  setModalOpen: React.Dispatch<React.SetStateAction<boolean>>;
  pipelineId: string;
}> = ({setModalOpen, pipelineId}) => {
  const {onDelete, updating, error} = useDeletePipelineModal(pipelineId);

  return (
    <BasicModal
      show={true}
      onHide={() => {
        setModalOpen(false);
      }}
      headerContent="Are you sure you want to delete this Pipeline?"
      actionable
      mode="Small"
      confirmText="Delete"
      onConfirm={onDelete}
      updating={updating}
      loading={false}
      disabled={updating}
      errorMessage={error}
    >
      Deleting this pipeline will erase all data inside it.
    </BasicModal>
  );
};

export default DeletePipelineButton;

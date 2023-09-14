import React from 'react';

import {BasicModal} from '@pachyderm/components';
import getServerErrorMessage from 'lib/errorHandling';

import useDeleteRepoModal from './hooks/useDeleteRepoModal';

const DeleteRepoModal: React.FC<{
  setModalOpen: React.Dispatch<React.SetStateAction<boolean>>;
}> = ({setModalOpen}) => {
  const {onDelete, updating, error} = useDeleteRepoModal();

  return (
    <BasicModal
      show={true}
      onHide={() => {
        setModalOpen(false);
      }}
      headerContent="Are you sure you want to delete this Repo?"
      actionable
      mode="Small"
      confirmText="Delete"
      onConfirm={onDelete}
      updating={updating}
      loading={false}
      disabled={updating}
      errorMessage={getServerErrorMessage(error)}
    >
      Deleting this repo will erase all data inside it.
    </BasicModal>
  );
};

export default DeleteRepoModal;

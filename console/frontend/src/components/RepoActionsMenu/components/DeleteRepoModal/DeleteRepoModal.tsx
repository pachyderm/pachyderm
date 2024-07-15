import React from 'react';

import {BasicModal} from '@pachyderm/components';

import useDeleteRepoModal from './hooks/useDeleteRepoModal';

const DeleteRepoModal: React.FC<{
  setModalOpen: React.Dispatch<React.SetStateAction<boolean>>;
  repoId: string;
}> = ({setModalOpen, repoId}) => {
  const {onDelete, updating, error} = useDeleteRepoModal(repoId);

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
      errorMessage={error}
    >
      Deleting this repo will erase all data inside it.
    </BasicModal>
  );
};

export default DeleteRepoModal;

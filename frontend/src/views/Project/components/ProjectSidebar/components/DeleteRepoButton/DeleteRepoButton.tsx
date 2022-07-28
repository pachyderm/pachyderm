import {TrashSVG, Tooltip, BasicModal, Button} from '@pachyderm/components';
import React from 'react';

import useDeleteRepoButton from './hooks/useDeleteRepoButton';

const DeleteRepoButton: React.FC = () => {
  const {canDelete, modalOpen, setModalOpen, onDelete, updating, error} =
    useDeleteRepoButton();

  return (
    <>
      {modalOpen && (
        <BasicModal
          show={true}
          onHide={() => {
            setModalOpen(false);
          }}
          headerContent="Are you sure you want to delete this Repo?"
          actionable
          small
          confirmText="Delete"
          onConfirm={onDelete}
          updating={updating}
          loading={false}
          disabled={updating}
          errorMessage={error && 'Error deleting repo. Please try again later.'}
        >
          Deleting this repo will erase all data inside it.
        </BasicModal>
      )}
      <Tooltip
        tooltipKey="Delete info"
        tooltipText={
          canDelete
            ? 'Delete Repo'
            : "This repo can't be deleted while it has associated pipelines."
        }
      >
        <Button
          disabled={!canDelete}
          onClick={() => setModalOpen(true)}
          data-testid="DeleteRepoButton__link"
          buttonType="ghost"
          IconSVG={TrashSVG}
          aria-label="Delete"
        />
      </Tooltip>
    </>
  );
};

export default DeleteRepoButton;

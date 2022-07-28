import {TrashSVG, Tooltip, BasicModal, Button} from '@pachyderm/components';
import React from 'react';

import useDeletePipelineButton from './hooks/useDeletePipelineButton';

const DeletePipelineButton: React.FC = () => {
  const {canDelete, modalOpen, setModalOpen, onDelete, updating, error} =
    useDeletePipelineButton();

  return (
    <>
      {modalOpen && (
        <BasicModal
          show={true}
          onHide={() => {
            setModalOpen(false);
          }}
          headerContent="Are you sure you want to delete this Pipeline?"
          actionable
          small
          confirmText="Delete"
          onConfirm={onDelete}
          updating={updating}
          loading={false}
          disabled={updating}
          errorMessage={
            error && 'Error deleting pipeline. Please try again later.'
          }
        >
          Deleting this pipeline will erase all data inside it.
        </BasicModal>
      )}
      <Tooltip
        tooltipKey="Delete info"
        tooltipText={
          canDelete
            ? 'Delete Pipeline'
            : "This pipeline can't be deleted while it has downstream pipelines."
        }
      >
        <Button
          IconSVG={TrashSVG}
          disabled={!canDelete}
          onClick={() => setModalOpen(true)}
          data-testid="DeletePipelineButton__link"
          buttonType="ghost"
          aria-label="Delete"
        />
      </Tooltip>
    </>
  );
};

export default DeletePipelineButton;

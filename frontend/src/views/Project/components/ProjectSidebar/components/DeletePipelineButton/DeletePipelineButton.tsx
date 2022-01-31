import {
  ButtonLink,
  TrashSVG,
  Icon,
  Tooltip,
  BasicModal,
} from '@pachyderm/components';
import React from 'react';

import useDeletePipelineButton from './hooks/useDeletePipelineButton';

const DeletePipelineButton: React.FC = () => {
  const {canDelete, modalOpen, setModalOpen, onDelete, updating} =
    useDeletePipelineButton();

  return (
    <>
      {modalOpen && (
        <BasicModal
          show={true}
          onHide={() => {
            setModalOpen(false);
          }}
          headerContent={
            <span>Are you sure you want to delete this Pipeline?</span>
          }
          actionable
          small
          confirmText="Delete"
          onConfirm={onDelete}
          updating={updating}
          loading={false}
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
        <ButtonLink
          disabled={!canDelete}
          onClick={() => setModalOpen(true)}
          data-testid="DeletePipelineButton__link"
        >
          <Icon color="plum">
            <TrashSVG aria-label="Delete" />
          </Icon>
        </ButtonLink>
      </Tooltip>
    </>
  );
};

export default DeletePipelineButton;

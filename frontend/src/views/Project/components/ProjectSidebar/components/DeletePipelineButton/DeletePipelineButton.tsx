import React from 'react';

import {TrashSVG, Tooltip, Button} from '@pachyderm/components';

import DeletePipelineModal from './DeletePipelineModal';
import useDeletePipelineButton from './hooks/useDeletePipelineButton';

const DeletePipelineButton: React.FC = () => {
  const {canDelete, modalOpen, setModalOpen} = useDeletePipelineButton();

  return (
    <>
      {modalOpen && <DeletePipelineModal setModalOpen={setModalOpen} />}
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

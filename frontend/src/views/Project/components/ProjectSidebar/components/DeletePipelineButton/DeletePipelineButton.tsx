import React from 'react';

import {TrashSVG, Tooltip, Button} from '@pachyderm/components';

import styles from './DeletePipelineButton.module.css';
import DeletePipelineModal from './DeletePipelineModal';
import useDeletePipelineButton from './hooks/useDeletePipelineButton';

const DeletePipelineButton: React.FC = () => {
  const {modalOpen, setModalOpen, tooltipText, disableButton} =
    useDeletePipelineButton();

  return (
    <>
      {modalOpen && <DeletePipelineModal setModalOpen={setModalOpen} />}
      <Tooltip tooltipKey="Delete info" tooltipText={tooltipText}>
        <span>
          <Button
            IconSVG={TrashSVG}
            disabled={disableButton}
            onClick={() => setModalOpen(true)}
            data-testid="DeletePipelineButton__link"
            buttonType="ghost"
            aria-label="Delete"
            className={disableButton && styles.pointerEventsNone}
          />
        </span>
      </Tooltip>
    </>
  );
};

export default DeletePipelineButton;

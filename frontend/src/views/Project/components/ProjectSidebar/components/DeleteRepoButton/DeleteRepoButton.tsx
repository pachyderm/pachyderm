import React from 'react';

import {TrashSVG, Tooltip, Button} from '@pachyderm/components';

import styles from './DeleteRepoButton.module.css';
import DeleteRepoModal from './DeleteRepoModal';
import useDeleteRepoButton from './hooks/useDeleteRepoButton';

const DeleteRepoButton: React.FC = () => {
  const {modalOpen, setModalOpen, disableButton, tooltipText} =
    useDeleteRepoButton();

  return (
    <>
      {modalOpen && <DeleteRepoModal setModalOpen={setModalOpen} />}
      <Tooltip tooltipKey="Delete info" tooltipText={tooltipText} size="large">
        <span>
          <Button
            disabled={disableButton}
            onClick={() => setModalOpen(true)}
            data-testid="DeleteRepoButton__link"
            buttonType="ghost"
            IconSVG={TrashSVG}
            aria-label="Delete"
            className={disableButton && styles.pointerEventsNone}
          />
        </span>
      </Tooltip>
    </>
  );
};

export default DeleteRepoButton;

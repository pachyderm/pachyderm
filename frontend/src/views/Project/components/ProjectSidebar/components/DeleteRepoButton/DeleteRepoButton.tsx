import React from 'react';

import {TrashSVG, Tooltip, Button} from '@pachyderm/components';

import DeleteRepoModal from './DeleteRepoModal';
import useDeleteRepoButton from './hooks/useDeleteRepoButton';

const DeleteRepoButton: React.FC = () => {
  const {modalOpen, setModalOpen, disableButton, tooltipText} =
    useDeleteRepoButton();

  return (
    <>
      {modalOpen && <DeleteRepoModal setModalOpen={setModalOpen} />}
      <Tooltip tooltipText={tooltipText}>
        <Button
          disabled={disableButton}
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

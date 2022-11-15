import React from 'react';

import {TrashSVG, Tooltip, Button} from '@pachyderm/components';

import DeleteRepoModal from './DeleteRepoModal';
import useDeleteRepoButton from './hooks/useDeleteRepoButton';

const DeleteRepoButton: React.FC = () => {
  const {modalOpen, setModalOpen, canDelete} = useDeleteRepoButton();

  return (
    <>
      {modalOpen && <DeleteRepoModal setModalOpen={setModalOpen} />}
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

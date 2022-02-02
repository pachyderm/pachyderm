import {File} from '@graphqlTypes';
import {ButtonLink, BasicModal} from '@pachyderm/components';
import React from 'react';

import useDeleteFileButton from './hooks/useDeleteFileButton';

type DeleteFileButtonProps = {
  file: File;
};

const DeleteRepoButton: React.FC<DeleteFileButtonProps> = ({
  file,
  children,
}) => {
  const {isOpen, openModal, closeModal, deleteFile, loading, error} =
    useDeleteFileButton(file);

  return (
    <>
      <BasicModal
        show={isOpen}
        onHide={closeModal}
        headerContent="Are you sure you want to delete this File?"
        actionable
        small
        confirmText="Delete"
        onConfirm={deleteFile}
        loading={loading}
        errorMessage={error?.message}
      >
        {file.path}
        <br />
        {`${file.repoName}@${file.commitId}`}
      </BasicModal>
      <ButtonLink onClick={openModal} data-testid="DeleteFileButton__link">
        {children}
      </ButtonLink>
    </>
  );
};

export default DeleteRepoButton;

import {File} from '@graphqlTypes';
import React, {Children} from 'react';

import {BasicModal, Button, TrashSVG} from '@pachyderm/components';

import useDeleteFileButton from './hooks/useDeleteFileButton';

type DeleteFileButtonProps = {
  file: File;
};

const DeleteRepoButton: React.FC<DeleteFileButtonProps> = ({
  file,
  children,
}) => {
  const {
    isOpen,
    openModal,
    closeModal,
    deleteFile,
    loading,
    error,
    deleteDisabled,
  } = useDeleteFileButton(file);

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
      {!deleteDisabled && (
        <Button
          buttonType="ghost"
          onClick={openModal}
          data-testid="DeleteFileButton__link"
          IconSVG={Children.count(children) === 0 ? TrashSVG : undefined}
        >
          {children}
        </Button>
      )}
    </>
  );
};

export default DeleteRepoButton;

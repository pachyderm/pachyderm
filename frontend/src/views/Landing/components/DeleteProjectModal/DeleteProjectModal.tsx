import React, {FunctionComponent} from 'react';

import {Group, Input, FormModal, ErrorText} from '@pachyderm/components';

import useDeleteProjectModal from './hooks/useDeleteProjectModal';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
  projectName: string;
};

const DeleteProjectModal: FunctionComponent<ModalProps> = ({
  show,
  onHide,
  projectName,
}) => {
  const {formCtx, error, handleSubmit, isFormComplete, loading, reset} =
    useDeleteProjectModal(projectName, onHide);

  return (
    <FormModal
      onHide={() => {
        onHide && onHide();
        reset({name: ''});
      }}
      error={error}
      updating={loading}
      formContext={formCtx}
      onSubmit={handleSubmit}
      loading={loading}
      confirmText="Delete Project"
      headerText={`Delete Project: ${projectName}`}
      disabled={!isFormComplete}
      isOpen={show}
    >
      <Group data-testid="DeleteProjectModal__modal" vertical spacing={16}>
        <Group vertical>
          <p>
            This action deletes the project, including its repos and pipelines.
            This cannot be undone.
          </p>
          <p>
            <ErrorText>
              Please type the name of your project to confirm deletion:{' '}
              <strong>{projectName}</strong>
            </ErrorText>
          </p>
          <Input
            data-testid="DeleteProjectModal__name"
            type="text"
            id="name"
            name="name"
            placeholder={`${projectName}`}
            clearable
            disabled={loading}
            autoFocus={true}
          />
        </Group>
      </Group>
    </FormModal>
  );
};

export default DeleteProjectModal;

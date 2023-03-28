import React from 'react';

import {Group, Label, TextArea, FormModal} from '@pachyderm/components';

import useUpdateProjectModal from './hooks/useUpdateProjectModal';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
  projectName?: string;
  description?: string | null;
};

const UpdateProjectModal: React.FC<ModalProps> = ({
  show,
  onHide,
  projectName,
  description,
}) => {
  const {formCtx, error, handleSubmit, isFormComplete, loading, reset} =
    useUpdateProjectModal(show, projectName, description, onHide);

  return (
    <FormModal
      onHide={() => {
        onHide && onHide();
        reset({description: ''});
      }}
      error={error}
      updating={loading}
      formContext={formCtx}
      onSubmit={handleSubmit}
      loading={loading}
      confirmText="Confirm Changes"
      headerText={`Edit Project: ${projectName}`}
      disabled={!isFormComplete}
      isOpen={show}
    >
      <Group data-testid="UpdateProjectModal__modal" vertical spacing={16}>
        <Group vertical spacing={32}>
          <div>
            <Label htmlFor="description" optional label="Description" />
            <TextArea
              data-testid="UpdateProjectModal__description"
              id="description"
              name="description"
              autoExpand
              autoFocus
              clearable
              disabled={loading}
            />
          </div>
        </Group>
      </Group>
    </FormModal>
  );
};

export default UpdateProjectModal;

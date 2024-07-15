import React, {FunctionComponent} from 'react';

import {Group, Input, Label, TextArea, FormModal} from '@pachyderm/components';

import useCreateProjectModal from './hooks/useCreateProjectModal';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
};

const CreateProjectModal: FunctionComponent<ModalProps> = ({show, onHide}) => {
  const {
    formCtx,
    error,
    handleSubmit,
    isFormComplete,
    loading,
    validateProjectName,
    reset,
  } = useCreateProjectModal(onHide);

  return (
    <FormModal
      onHide={() => {
        onHide && onHide();
        reset({name: '', description: ''});
      }}
      error={error}
      updating={loading}
      formContext={formCtx}
      onSubmit={handleSubmit}
      loading={loading}
      confirmText="Create"
      headerText="Create New Project"
      disabled={!isFormComplete}
      isOpen={show}
    >
      <Group data-testid="CreateProjectModal__modal" vertical spacing={16}>
        <Group vertical spacing={32}>
          <div>
            <Label
              htmlFor="name"
              label="Name (cannot be modified later)"
              maxLength={51}
            />
            <Input
              data-testid="CreateProjectModal__name"
              type="text"
              id="name"
              name="name"
              validationOptions={{
                maxLength: {
                  value: 51,
                  message: 'Project name exceeds maximum allowed length',
                },
                pattern: {
                  value: /^[a-zA-Z0-9-][a-zA-Z0-9-_]*$/,
                  message:
                    'Name can only contain alphanumeric characters, underscores, and dashes',
                },
                validate: validateProjectName,
              }}
              clearable
              disabled={loading}
              autoFocus={true}
            />
          </div>

          <div>
            <Label htmlFor="description" optional label="Description" />
            <TextArea
              data-testid="CreateProjectModal__description"
              id="description"
              name="description"
              autoExpand
              clearable
              disabled={loading}
            />
          </div>
        </Group>
      </Group>
    </FormModal>
  );
};

export default CreateProjectModal;

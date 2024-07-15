import React, {FunctionComponent} from 'react';

import {Group, Input, Label, TextArea, FormModal} from '@pachyderm/components';

import useCreateRepoModal from './hooks/useCreateRepoModal';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
};

const CreateRepoModal: FunctionComponent<ModalProps> = ({show, onHide}) => {
  const {
    formCtx,
    error,
    handleSubmit,
    isFormComplete,
    loading,
    validateRepoName,
    reset,
  } = useCreateRepoModal(onHide);

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
      headerText="Create New Repo"
      disabled={!isFormComplete}
      isOpen={show}
    >
      <Group data-testid="CreateRepoModal__modal" vertical spacing={16}>
        <Group vertical spacing={32}>
          <div>
            <Label htmlFor="name" label="Name" maxLength={63} />
            <Input
              data-testid="CreateRepoModal__name"
              type="text"
              id="name"
              name="name"
              validationOptions={{
                maxLength: {
                  value: 63,
                  message: 'Repo name exceeds maximum allowed length',
                },
                pattern: {
                  value: /^[a-zA-Z0-9-_]+$/,
                  message:
                    'Repo name can only contain alphanumeric characters, underscores, and dashes',
                },
                validate: validateRepoName,
              }}
              clearable
              disabled={loading}
              autoFocus={true}
            />
          </div>

          <div>
            <Label
              htmlFor="description"
              optional
              label="Description"
              maxLength={90}
            />
            <TextArea
              data-testid="CreateRepoModal__description"
              id="description"
              name="description"
              validationOptions={{
                maxLength: {
                  value: 90,
                  message: 'Repo description exceeds maximum allowed length',
                },
              }}
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

export default CreateRepoModal;

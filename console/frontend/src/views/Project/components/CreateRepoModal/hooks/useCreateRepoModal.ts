import isEmpty from 'lodash/isEmpty';
import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import {useCreateRepo} from '@dash-frontend/hooks/useCreateRepo';
import {useRepoSearch} from '@dash-frontend/hooks/useRepoSearch';
import useUrlState from '@dash-frontend/hooks/useUrlState';

type CreateRepoFormValues = {
  name: string;
  description: string;
};

const useCreateRepoModal = (onHide?: () => void) => {
  const {projectId} = useUrlState();
  const {createRepo, loading, error} = useCreateRepo({
    onSuccess: onHide,
  });
  const formCtx = useForm<CreateRepoFormValues>({mode: 'onChange'});

  const {
    watch,
    reset,
    formState: {errors: formErrors},
    setError,
  } = formCtx;

  const name = watch('name');

  const {getRepo} = useRepoSearch(() => {
    setError('name', {
      type: 'error',
      message: 'Repo name already in use',
    });
  });

  const validateRepoName = useCallback(
    (val: string) => {
      getRepo({
        repo: {project: {name: projectId}, name: val},
      });
      return undefined;
    },
    [getRepo, projectId],
  );

  const isFormComplete = Boolean(name) && isEmpty(formErrors);

  const handleSubmit = useCallback(
    async (values: CreateRepoFormValues) => {
      try {
        createRepo(
          {
            repo: {
              project: {name: projectId},
              name: values.name.trim(),
            },
            description: values.description,
          },
          {
            onSuccess: () => reset(),
          },
        );
      } catch (e) {
        return;
      }
    },
    [createRepo, projectId, reset],
  );

  return {
    formCtx,
    error,
    handleSubmit,
    isFormComplete,
    loading,
    validateRepoName,
    reset,
  };
};

export default useCreateRepoModal;

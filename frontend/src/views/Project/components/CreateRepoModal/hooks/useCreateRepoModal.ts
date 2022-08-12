import isEmpty from 'lodash/isEmpty';
import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import useCreateRepo from '@dash-frontend/hooks/useCreateRepo';
import useRepos from '@dash-frontend/hooks/useRepos';
import useUrlState from '@dash-frontend/hooks/useUrlState';

type CreateRepoFormValues = {
  name: string;
  description: string;
};

const useCreateRepoModal = (onHide?: () => void) => {
  const {projectId} = useUrlState();
  const {createRepo, loading, error} = useCreateRepo(onHide);
  const {
    repos,
    loading: reposLoading,
    error: reposError,
  } = useRepos({projectId});

  const formCtx = useForm<CreateRepoFormValues>({mode: 'onChange'});

  const {
    watch,
    reset,
    formState: {errors: formErrors},
  } = formCtx;

  const name = watch('name');

  const validateRepoName = useCallback(
    (value: string) => {
      if (repos && repos.map((repo) => repo?.name).includes(value)) {
        return 'Repo name already in use';
      }
    },
    [repos],
  );

  const isFormComplete = Boolean(name) && isEmpty(formErrors);

  const handleSubmit = useCallback(
    (values: CreateRepoFormValues) => {
      createRepo({
        projectId,
        name: values.name.trim(),
        description: values.description,
      });
      reset();
    },
    [createRepo, projectId, reset],
  );

  return {
    formCtx,
    error: reposError?.message || error?.message,
    handleSubmit,
    isFormComplete,
    loading: loading || reposLoading,
    validateRepoName,
    reset,
  };
};

export default useCreateRepoModal;

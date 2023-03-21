import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import useCreateProject from '@dash-frontend/hooks/useCreateProject';

type DeleteProjectFormValues = {
  name: string;
};

const useDeleteProjectModal = (projectName: string, onHide?: () => void) => {
  const {
    createProject,
    loading: createProjectLoading,
    error,
  } = useCreateProject(onHide); // TODO
  const formCtx = useForm<DeleteProjectFormValues>({mode: 'onChange'});
  const {watch, reset} = formCtx;
  const name = watch('name');
  const isFormComplete = Boolean(name) && name === projectName;
  const handleSubmit = useCallback(
    (values: DeleteProjectFormValues) => {
      createProject({
        name: values.name.trim(),
      });
      reset();
    },
    [createProject, reset],
  ); // TODO
  return {
    formCtx,
    error: error?.message,
    handleSubmit,
    isFormComplete,
    loading: createProjectLoading,
    reset,
  };
};

export default useDeleteProjectModal;

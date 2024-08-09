import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import {useDeleteProject} from '@dash-frontend/hooks/useDeleteProject';

type DeleteProjectFormValues = {
  name: string;
};

const useDeleteProjectModal = (projectName: string, onHide?: () => void) => {
  const {
    deleteProject,
    loading: deleteProjectLoading,
    error,
  } = useDeleteProject(onHide);
  const formCtx = useForm<DeleteProjectFormValues>({mode: 'onChange'});
  const {watch, reset} = formCtx;
  const name = watch('name');
  const isFormComplete =
    Boolean(name) && name.toLowerCase() === projectName.toLowerCase();

  const handleSubmit = useCallback(async () => {
    try {
      deleteProject(
        {
          project: {
            name: projectName.trim(),
          },
        },
        {
          onSuccess: () => reset(),
        },
      );
      reset();
    } catch (e) {
      return;
    }
  }, [projectName, deleteProject, reset]);

  return {
    formCtx,
    error,
    handleSubmit,
    isFormComplete,
    loading: deleteProjectLoading,
    reset,
  };
};

export default useDeleteProjectModal;

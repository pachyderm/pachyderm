import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import {useDeleteProjectAndResources} from '@dash-frontend/hooks/useDeleteProjectAndResources';

type DeleteProjectFormValues = {
  name: string;
};

const useDeleteProjectModal = (projectName: string, onHide?: () => void) => {
  const {
    deleteProject,
    loading: deleteProjectLoading,
    error,
  } = useDeleteProjectAndResources(onHide);
  const formCtx = useForm<DeleteProjectFormValues>({mode: 'onChange'});
  const {watch, reset} = formCtx;
  const name = watch('name');
  const isFormComplete =
    Boolean(name) && name.toLowerCase() === projectName.toLowerCase();
  const handleSubmit = useCallback(() => {
    deleteProject({
      name: projectName.trim(),
    });
    reset();
  }, [projectName, deleteProject, reset]);
  return {
    formCtx,
    error: error?.message,
    handleSubmit,
    isFormComplete,
    loading: deleteProjectLoading,
    reset,
  };
};

export default useDeleteProjectModal;

import isEmpty from 'lodash/isEmpty';
import {useCallback, useEffect} from 'react';
import {useForm} from 'react-hook-form';

import {useUpdateProject} from '@dash-frontend/hooks/useUpdateProject';

type UpdateProjectFormValues = {
  description: string;
};

const useUpdateProjectModal = (
  show: boolean,
  projectName?: string,
  description?: string | null,
  onHide?: () => void,
) => {
  const {
    updateProject,
    loading: updateProjectLoading,
    error,
  } = useUpdateProject(onHide);

  const formCtx = useForm<UpdateProjectFormValues>({
    mode: 'onChange',
  });

  useEffect(() => {
    formCtx.setValue('description', description || '');
    // This should NOT update when description changes. Otherwise this will
    // fire if description is updated via pachctl and requeried via poll interval
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formCtx, show]);

  const {
    reset,
    formState: {errors: formErrors},
  } = formCtx;

  const isFormComplete = isEmpty(formErrors);

  const handleSubmit = useCallback(
    async (values: UpdateProjectFormValues) => {
      try {
        updateProject(
          {
            project: {
              name: projectName,
            },
            ...values,
          },
          {
            onSuccess: () => reset(),
          },
        );
      } catch (e) {
        return;
      }
    },
    [updateProject, projectName, reset],
  );

  return {
    formCtx,
    error: error,
    handleSubmit,
    isFormComplete,
    loading: updateProjectLoading,
    reset,
  };
};

export default useUpdateProjectModal;

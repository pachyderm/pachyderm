import isEmpty from 'lodash/isEmpty';
import {useCallback, useEffect} from 'react';
import {useForm} from 'react-hook-form';

import {useUpdateProjectMutation} from '@dash-frontend/generated/hooks';

type UpdateProjectFormValues = {
  description: string;
};

const useUpdateProjectModal = (
  show: boolean,
  projectName?: string,
  description?: string | null,
  onHide?: () => void,
) => {
  const [updateProjectMutation, {loading: updateProjectLoading, error}] =
    useUpdateProjectMutation({
      onCompleted: onHide,
    });

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
        await updateProjectMutation({
          variables: {
            args: {
              name: projectName || '',
              description: values.description,
            },
          },
        });
        reset();
      } catch (e) {
        return;
      }
    },
    [updateProjectMutation, projectName, reset],
  );

  return {
    formCtx,
    error: error?.message,
    handleSubmit,
    isFormComplete,
    loading: updateProjectLoading,
    reset,
  };
};

export default useUpdateProjectModal;

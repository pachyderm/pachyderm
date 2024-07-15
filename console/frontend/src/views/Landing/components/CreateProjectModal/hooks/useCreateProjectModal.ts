import isEmpty from 'lodash/isEmpty';
import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import {useCreateProject} from '@dash-frontend/hooks/useCreateProject';
import {useProjects} from '@dash-frontend/hooks/useProjects';

type CreateProjectFormValues = {
  name: string;
  description: string;
};

const useCreateProjectModal = (onHide?: () => void) => {
  const {
    createProject,
    loading: createProjectLoading,
    error,
  } = useCreateProject(onHide);
  const {
    projects,
    loading: projectsLoading,
    error: projectsError,
  } = useProjects();

  const formCtx = useForm<CreateProjectFormValues>({mode: 'onChange'});

  const {
    watch,
    reset,
    formState: {errors: formErrors},
  } = formCtx;

  const name = watch('name');

  const validateProjectName = useCallback(
    (value: string) => {
      if (
        projects &&
        projects.map((project) => project.project?.name).includes(value)
      ) {
        return 'Project name already in use';
      }
    },
    [projects],
  );

  const isFormComplete = Boolean(name) && isEmpty(formErrors);

  const handleSubmit = useCallback(
    async (values: CreateProjectFormValues) => {
      try {
        createProject(
          {
            project: {name: values.name.trim()},
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
    [createProject, reset],
  );

  return {
    formCtx,
    error: projectsError || error,
    handleSubmit,
    isFormComplete,
    loading: createProjectLoading || projectsLoading,
    validateProjectName,
    reset,
  };
};

export default useCreateProjectModal;

import {useMutation, useQueryClient} from '@tanstack/react-query';

import {createProject, CreateProjectRequest} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useCreateProject = (onSettled?: () => void) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['createProject'],
    mutationFn: (req: CreateProjectRequest) => createProject(req),
    onSuccess: () => {
      client.invalidateQueries({queryKey: queryKeys.projects, exact: true});
    },
    onSettled,
  });

  return {
    createProject: mutate,
    loading,
    error: getErrorMessage(error),
  };
};

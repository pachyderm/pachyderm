import {useMutation, useQueryClient} from '@tanstack/react-query';

import {updateProject, CreateProjectRequest} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useUpdateProject = (onSuccess?: () => void) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['updateProject'],
    mutationFn: (req: CreateProjectRequest) => updateProject(req),
    onSuccess: () => {
      client.invalidateQueries({queryKey: queryKeys.projects, exact: true});
      onSuccess && onSuccess();
    },
  });

  return {
    updateProject: mutate,
    loading,
    error: getErrorMessage(error),
  };
};

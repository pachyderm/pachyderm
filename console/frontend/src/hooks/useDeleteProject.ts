import {useMutation, useQueryClient} from '@tanstack/react-query';

import {deleteProject, DeleteProjectRequest} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useDeleteProject = (onSettled?: () => void) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['deleteProject'],
    mutationFn: (req: DeleteProjectRequest) => deleteProject(req),
    onSuccess: () => {
      client.invalidateQueries({queryKey: queryKeys.projects, exact: true});
    },
    onSettled,
  });

  return {
    deleteProject: mutate,
    loading,
    error: getErrorMessage(error),
  };
};

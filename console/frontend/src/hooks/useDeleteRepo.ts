import {useMutation, useQueryClient} from '@tanstack/react-query';

import {deleteRepo, DeleteRepoRequest} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useDeleteRepo = (onSuccess: () => void) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['deleteRepo'],
    mutationFn: (req: DeleteRepoRequest) => deleteRepo(req),
    onSuccess: (_data, variables) => {
      client.invalidateQueries({
        queryKey: queryKeys.repos({
          projectId: variables.repo?.project?.name,
        }),
        exact: true,
      });
      onSuccess && onSuccess();
    },
  });

  return {
    deleteRepo: mutate,
    error: getErrorMessage(error),
    loading,
  };
};

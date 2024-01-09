import {useMutation, useQueryClient} from '@tanstack/react-query';

import {createRepo, CreateRepoRequest} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type useCreateRepoArgs = {
  onSuccess?: () => void;
};

export const useCreateRepo = ({onSuccess}: useCreateRepoArgs) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['createRepo'],
    mutationFn: (req: CreateRepoRequest) => createRepo(req),
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
    createRepo: mutate,
    loading,
    error: getErrorMessage(error),
  };
};

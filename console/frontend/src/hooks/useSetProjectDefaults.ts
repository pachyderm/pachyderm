import {useMutation, useQueryClient} from '@tanstack/react-query';

import {
  setProjectDefaults,
  SetProjectDefaultsRequest,
} from '@dash-frontend/api/pps';
import queryKeys from '@dash-frontend/lib/queryKeys';

type useSetProjectDefaultsArgs = {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
};

export const useSetProjectDefaults = ({
  onSuccess,
  onError,
}: useSetProjectDefaultsArgs) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
    data,
  } = useMutation({
    mutationKey: ['setProjectDefaults'],
    mutationFn: (req: SetProjectDefaultsRequest) => setProjectDefaults(req),
    onSuccess: (_data, variables) => {
      client.invalidateQueries({
        queryKey: queryKeys.projectDefaults({
          projectId: variables.project?.name || '',
        }),
      });
      onSuccess && onSuccess();
    },
    onError: (error) => onError && onError(error),
  });

  return {
    setProjectDefaultsResponse: data,
    error,
    setProjectDefaults: mutate,
    loading,
  };
};

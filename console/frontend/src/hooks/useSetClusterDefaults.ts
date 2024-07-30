import {useMutation, useQueryClient} from '@tanstack/react-query';

import {
  setClusterDefaults,
  SetClusterDefaultsRequest,
} from '@dash-frontend/api/pps';
import queryKeys from '@dash-frontend/lib/queryKeys';

type useSetClusterDefaultsArgs = {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
};

export const useSetClusterDefaults = ({
  onSuccess,
  onError,
}: useSetClusterDefaultsArgs) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
    data,
  } = useMutation({
    mutationKey: ['setClusterDefaults'],
    mutationFn: (req: SetClusterDefaultsRequest) => setClusterDefaults(req),
    onSuccess: () => {
      client.invalidateQueries({queryKey: queryKeys.clusterDefaults});
      onSuccess && onSuccess();
    },
    onError: (error) => onError && onError(error),
  });

  return {
    setClusterDefaultsResponse: data,
    error,
    setClusterDefaults: mutate,
    loading,
  };
};

import {useQuery} from '@tanstack/react-query';

import {inspectCluster} from '@dash-frontend/api/admin';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useInspectCluster = (enabled = true) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.inspectCluster,
    queryFn: () => inspectCluster(),
    enabled,
  });

  return {
    error,
    cluster: data,
    loading,
  };
};

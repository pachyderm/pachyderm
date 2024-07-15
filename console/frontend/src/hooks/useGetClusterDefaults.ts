import {useQuery} from '@tanstack/react-query';

import {getClusterDefaults} from '@dash-frontend/api/pps';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useGetClusterDefaults = () => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.clusterDefaults,
    queryFn: () => getClusterDefaults(),
  });

  return {
    error,
    clusterDefaults: data,
    loading,
  };
};

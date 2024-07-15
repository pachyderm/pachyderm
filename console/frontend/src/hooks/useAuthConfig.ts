import {useQuery} from '@tanstack/react-query';

import {config} from '@dash-frontend/api/auth';

interface UseAuthConfigArgs {
  enabled?: boolean;
}

const useAuthConfig = ({enabled = true}: UseAuthConfigArgs = {}) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ['authConfig'],
    queryFn: () => config(),
    enabled,
    throwOnError: false,
  });

  return {
    authConfig: data,
    loading,
    error,
  };
};

export default useAuthConfig;

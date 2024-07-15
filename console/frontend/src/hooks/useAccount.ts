import {useQuery} from '@tanstack/react-query';

import {account} from '@dash-frontend/api/auth';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useAccount = (enabled = true) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.account,
    queryFn: () => account(),
    enabled,
  });

  return {
    error: getErrorMessage(error),
    account: data,
    loading,
    displayName: data?.name || data?.email,
  };
};

import {useQuery} from '@tanstack/react-query';

import {getVersion} from '@dash-frontend/api/version';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseVersionProps = {
  enabled?: boolean;
};

export const useVersion = ({enabled = true}: UseVersionProps = {}) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.version,
    queryFn: () => getVersion(),
    enabled,
  });

  return {
    version: data,
    error: getErrorMessage(error),
    loading,
  };
};

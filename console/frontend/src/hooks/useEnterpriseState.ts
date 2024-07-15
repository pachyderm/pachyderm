import {useQuery} from '@tanstack/react-query';

import {getEnterpriseState} from '@dash-frontend/api/enterprise';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseEnterpriseStateProps = {
  enabled?: boolean;
};

export const useEnterpriseState = ({
  enabled = true,
}: UseEnterpriseStateProps = {}) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.enterpriseState,
    queryFn: () => getEnterpriseState(),
    enabled,
  });

  return {
    enterpriseState: data,
    error: getErrorMessage(error),
    loading,
  };
};

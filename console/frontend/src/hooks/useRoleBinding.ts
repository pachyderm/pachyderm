import {useQuery} from '@tanstack/react-query';

import {getRoleBinding, GetRoleBindingRequest} from '@dash-frontend/api/auth';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useRoleBinding = (args: GetRoleBindingRequest, enabled = true) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.roleBinding<GetRoleBindingRequest>({args}),
    queryFn: () => getRoleBinding(args),
    enabled,
  });

  return {
    loading,
    roleBinding: data,
    error: getErrorMessage(error),
  };
};

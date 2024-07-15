import {useMutation, useQueryClient} from '@tanstack/react-query';

import {
  modifyRoleBinding,
  GetRoleBindingRequest,
  ModifyRoleBindingRequest,
} from '@dash-frontend/api/auth';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useModifyRoleBinding = (req: GetRoleBindingRequest) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['modifyRoleBinding'],
    mutationFn: (req: ModifyRoleBindingRequest) => modifyRoleBinding(req),
    onSuccess: () =>
      client.invalidateQueries({
        queryKey: queryKeys.roleBinding<GetRoleBindingRequest>({args: req}),
      }),
  });

  return {
    modifyRoleBinding: mutate,
    loading,
    error: getErrorMessage(error),
  };
};

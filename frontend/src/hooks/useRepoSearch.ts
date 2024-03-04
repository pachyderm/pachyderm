import {useMutation} from '@tanstack/react-query';

import {InspectRepoRequest, inspectRepo} from '@dash-frontend/api/pfs';
import {isNotFound} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';

export const useRepoSearch = (onSuccess: () => void) => {
  const {mutate, data, error} = useMutation({
    throwOnError: (e) => !isNotFound(e),
    mutationFn: (req: InspectRepoRequest) => inspectRepo(req),
    onSuccess: () => {
      onSuccess && onSuccess();
    },
  });
  return {
    getRepo: mutate,
    error: getErrorMessage(error),
    repo: data,
  };
};

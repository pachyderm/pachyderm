import {useQuery} from '@tanstack/react-query';

import {listBranch} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseBranches = {
  projectId: string;
  repoId: string;
};

export const useBranches = (args: UseBranches) => {
  const {projectId, repoId} = {...args};
  const {
    data,
    error,
    isLoading: loading,
  } = useQuery({
    queryKey: queryKeys.branches({projectId, repoId}),
    queryFn: () =>
      listBranch({
        repo: {
          name: repoId,
          type: 'user',
          project: {name: projectId},
        },
      }),
  });
  return {
    error: getErrorMessage(error),
    loading,
    branches: data,
  };
};

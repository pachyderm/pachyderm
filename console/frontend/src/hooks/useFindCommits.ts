import {useQuery} from '@tanstack/react-query';

import {findCommitsPaged, FindCommitsRequest} from '@dash-frontend/api/pfs';
import {DEFAULT_FIND_COMMITS_LIMIT} from '@dash-frontend/constants/limits';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useFindCommits = (req: FindCommitsRequest) => {
  const {
    data,
    isLoading: loading,
    error,
    refetch,
  } = useQuery({
    queryKey: queryKeys.findCommits<FindCommitsRequest>({
      projectId: req.start?.repo?.project?.name,
      repoId: req.start?.repo?.name,
      args: req,
    }),
    queryFn: () => {
      if (!req.start?.id && !req.start?.branch?.name) {
        throw new Error(
          `INVALID_ARGUMENT: Need to specify commitId and or branchId`,
        );
      }
      return findCommitsPaged({
        ...req,
        limit: req.limit || DEFAULT_FIND_COMMITS_LIMIT,
      });
    },
    enabled: false,
  });

  return {
    findCommits: refetch,
    commits: data?.commits,
    cursor: data?.cursor,
    loading,
    error: getErrorMessage(error),
  };
};

import {CommitsQueryArgs} from '@graphqlTypes';

import {COMMITS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useGetCommitsQuery} from '@dash-frontend/generated/hooks';

export const COMMIT_LIMIT = 100;

type UseCommitArgs = {
  args: CommitsQueryArgs;
  skip?: boolean;
};

const useCommits = ({args, skip = false}: UseCommitArgs) => {
  const {data, error, loading} = useGetCommitsQuery({
    pollInterval: COMMITS_POLL_INTERVAL_MS,
    variables: {args},
    skip,
  });

  return {
    commits: data?.commits,
    error,
    loading,
  };
};

export default useCommits;

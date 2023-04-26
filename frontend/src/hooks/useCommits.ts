import {WatchQueryFetchPolicy} from '@apollo/client';
import {CommitsQueryArgs} from '@graphqlTypes';

import {COMMITS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useGetCommitsQuery} from '@dash-frontend/generated/hooks';

export const COMMIT_LIMIT = 100;

type UseCommitArgs = {
  args: CommitsQueryArgs;
  skip?: boolean;
  fetchPolicy?: WatchQueryFetchPolicy;
};

const useCommits = ({args, skip = false, fetchPolicy}: UseCommitArgs) => {
  const getCommitsQuery = useGetCommitsQuery({
    pollInterval: COMMITS_POLL_INTERVAL_MS,
    variables: {args},
    skip,
    fetchPolicy,
  });

  return {
    ...getCommitsQuery,
    commits: getCommitsQuery.data?.commits.items || [],
    cursor: getCommitsQuery.data?.commits.cursor,
    parentCommit: getCommitsQuery.data?.commits.parentCommit,
  };
};

export default useCommits;

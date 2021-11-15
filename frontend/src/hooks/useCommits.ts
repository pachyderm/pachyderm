import {CommitsQueryArgs} from '@graphqlTypes';

import {useGetCommitsQuery} from '@dash-frontend/generated/hooks';

export const COMMIT_LIMIT = 100;

const useCommits = (args: CommitsQueryArgs) => {
  const {data, error, loading} = useGetCommitsQuery({
    variables: {args},
  });

  return {
    commits: data?.commits,
    error,
    loading,
  };
};

export default useCommits;

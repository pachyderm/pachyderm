import {QueryFunctionOptions} from '@apollo/client';
import {CommitSearchQueryArgs} from '@graphqlTypes';

import {useCommitSearchQuery} from '@dash-frontend/generated/hooks';

const useCommitSearch = (
  args: CommitSearchQueryArgs,
  opts?: QueryFunctionOptions,
) => {
  const commitSearchQuery = useCommitSearchQuery({
    variables: {args},
    skip: opts?.skip,
  });

  return {
    ...commitSearchQuery,
    commit: commitSearchQuery.data?.commitSearch,
  };
};

export default useCommitSearch;

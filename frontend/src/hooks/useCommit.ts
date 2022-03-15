import {CommitQueryArgs} from '@graphqlTypes';

import {useCommitQuery} from '@dash-frontend/generated/hooks';

type UseCommitArgs = {
  args: CommitQueryArgs;
};

const useCommit = ({args}: UseCommitArgs) => {
  const {data, error, loading} = useCommitQuery({
    variables: {args},
  });

  return {
    commit: data?.commit,
    error,
    loading,
  };
};

export default useCommit;

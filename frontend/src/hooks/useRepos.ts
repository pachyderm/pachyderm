import {ReposQueryArgs} from '@graphqlTypes';

import {useReposQuery} from '@dash-frontend/generated/hooks';

const useRepos = (args: ReposQueryArgs) => {
  const {data, error, loading} = useReposQuery({
    variables: {args},
  });

  return {
    repos: data?.repos,
    error,
    loading,
  };
};

export default useRepos;

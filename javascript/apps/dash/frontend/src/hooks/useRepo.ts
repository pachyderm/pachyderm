import {useRepoQuery} from '@dash-frontend/generated/hooks';
import {RepoQueryArgs} from '@graphqlTypes';

const useRepo = (args: RepoQueryArgs) => {
  const {data, error, loading} = useRepoQuery({
    variables: {args},
  });

  return {
    repo: data?.repo,
    error,
    loading,
  };
};

export default useRepo;

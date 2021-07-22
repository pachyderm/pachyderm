import {useAuthConfigQuery} from '@dash-frontend/generated/hooks';

interface UseAuthConfigArgs {
  skip?: boolean;
}

const useAuthConfig = ({skip = false}: UseAuthConfigArgs = {}) => {
  const {data, loading, error} = useAuthConfigQuery({skip});

  return {
    authConfig: data?.authConfig,
    loading,
    error,
  };
};

export default useAuthConfig;

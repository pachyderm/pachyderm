import {getRandomName} from '@pachyderm/components';

import {useGetAccountQuery} from '@dash-frontend/generated/hooks';

const tutorialId = getRandomName();

interface useAccountArgs {
  skip?: boolean;
}

const useAccount = ({skip = false}: useAccountArgs = {}) => {
  const {data, error, loading} = useGetAccountQuery({skip});

  return {
    error,
    account: data?.account,
    displayName: data?.account.name || data?.account.email,
    loading,
    tutorialId,
  };
};

export default useAccount;

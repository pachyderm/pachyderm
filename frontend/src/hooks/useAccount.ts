import {useEffect} from 'react';

import {useGetAccountQuery} from '@dash-frontend/generated/hooks';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {getRandomName} from '@pachyderm/components';
interface useAccountArgs {
  skip?: boolean;
}

const useAccount = ({skip = false}: useAccountArgs = {}) => {
  const {data, error, loading} = useGetAccountQuery({skip});
  const [tutorialId, setTutorialData] = useLocalProjectSettings({
    projectId: 'account-data',
    key: 'tutorial_id',
  });

  useEffect(() => {
    if (!tutorialId) {
      setTutorialData(getRandomName());
    }
  }, [setTutorialData, tutorialId]);

  return {
    error,
    account: data?.account,
    displayName: data?.account.name || data?.account.email,
    loading,
    tutorialId: tutorialId,
  };
};

export default useAccount;

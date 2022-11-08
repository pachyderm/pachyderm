import {useFlags} from 'launchdarkly-react-client-sdk';

const useFeatureFlags = () => {
  const flags = useFlags();

  return flags;
};

export default useFeatureFlags;

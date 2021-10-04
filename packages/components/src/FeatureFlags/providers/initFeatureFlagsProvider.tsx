import {useLDClient, withLDProvider} from 'launchdarkly-react-client-sdk';
import React, {useEffect} from 'react';

type FeatureFlagProviderProps = {
  email?: string;
  key?: string;
  name?: string;
};

const FeatureFlagsProvider: React.FC<FeatureFlagProviderProps> = ({
  children,
  email,
  key,
  name,
}) => {
  const client = useLDClient();

  useEffect(() => {
    if (client && email && key && name) {
      client?.identify({
        email,
        key,
        name,
      });
    }
  }, [client, email, key, name]);

  return <>{children}</>;
};

const initFeatureFlagProvider = (apiKey?: string) =>
  apiKey
    ? withLDProvider({clientSideID: apiKey})(FeatureFlagsProvider)
    : FeatureFlagsProvider;

export default initFeatureFlagProvider;

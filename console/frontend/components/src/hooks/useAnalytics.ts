import {useCallback, useEffect} from 'react';

import {
  captureTrackingCookies,
  fireIdentify,
  initPageTracker,
  initClickTracker,
  fireUTM,
  fireClusterInfo,
} from '../lib/analytics';

export type UseAnalyticsProps = {
  createdAt?: number;
  email?: string;
  id?: string;
  clusterId?: string;
  provider: {
    /* eslint-disable @typescript-eslint/no-explicit-any */
    identify: (...args: any[]) => void;
    page: (...args: any[]) => void;
    track: (...args: any[]) => void;
    /* eslint-enable @typescript-eslint/no-explicit-any */
  };
};

const useAnalytics = ({
  createdAt,
  email,
  id,
  clusterId,
  provider,
}: UseAnalyticsProps) => {
  const init = useCallback(() => {
    if (window.analyticsInitialized) {
      return;
    }

    window.analyticsInitialized = true;

    // Capture incoming cookies
    captureTrackingCookies();

    // Track clicks
    initClickTracker(provider.track);

    // Track UTM
    fireUTM(provider.track);

    // Track page views
    initPageTracker(provider.page);
  }, [provider]);

  useEffect(() => {
    if (!window.analyticsIdentified && createdAt && email && id) {
      window.analyticsIdentified = true;
      fireIdentify(id, email, createdAt, provider.identify, provider.track);
    }
  }, [createdAt, id, email, provider]);

  useEffect(() => {
    if (!window.clusterIdentified && clusterId) {
      window.clusterIdentified = true;
      fireClusterInfo(clusterId, provider.track);
    }
  }, [clusterId, provider]);

  return {init};
};

export default useAnalytics;

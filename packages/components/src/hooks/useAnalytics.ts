import {useCallback, useEffect} from 'react';

import {
  captureTrackingCookies,
  fireIdentify,
  initPageTracker,
  initClickTracker,
  fireUTM,
} from '../lib/analytics';

export type UseAnalyticsProps = {
  createdAt?: number;
  email?: string;
  id?: string;
  provider: {
    /* eslint-disable @typescript-eslint/no-explicit-any */
    getAnonymousId: (...args: any[]) => void;
    identify: (...args: any[]) => void;
    page: (...args: any[]) => void;
    track: (...args: any[]) => void;
    /* eslint-enable @typescript-eslint/no-explicit-any */
  };
};

const useAnalytics = ({createdAt, email, id, provider}: UseAnalyticsProps) => {
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
    if (window.analyticsIdentified) {
      return;
    }

    if (createdAt && email && id) {
      window.analyticsIdentified = true;
      fireIdentify(
        id,
        email,
        createdAt,
        provider.identify,
        provider.track,
        provider.getAnonymousId,
      );
    }
  }, [createdAt, id, email, provider]);

  return {init};
};

export default useAnalytics;

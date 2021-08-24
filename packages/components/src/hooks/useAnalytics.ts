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
};

const useAnalytics = ({createdAt, email, id}: UseAnalyticsProps) => {
  const init = useCallback(() => {
    if (window.analyticsInitialized) {
      return;
    }

    window.analyticsInitialized = true;

    // Capture incoming cookies
    captureTrackingCookies();

    // Track clicks
    initClickTracker();

    // Track UTM
    fireUTM();

    // Track page views
    initPageTracker();
  }, []);

  useEffect(() => {
    if (window.analyticsIdentified) {
      return;
    }

    if (createdAt && email && id) {
      window.analyticsIdentified = true;
      fireIdentify(id, email, createdAt);
    }
  }, [createdAt, id, email]);

  return {init};
};

export default useAnalytics;

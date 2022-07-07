import {useAnalytics} from '@pachyderm/components';
import * as Sentry from '@sentry/react';
import LogRocket from 'logrocket';
import React from 'react';
import {getAnonymousId, identify, page, track} from 'rudder-sdk-js';

import useAccount from '@dash-frontend/hooks/useAccount';
import useAuth from '@dash-frontend/hooks/useAuth';

const enableTelemetry =
  process.env.REACT_APP_RUNTIME_DISABLE_TELEMETRY !== 'true';

const AnalyticsProvider: React.FC = ({children}) => {
  const {loggedIn} = useAuth();
  const {account} = useAccount({skip: !loggedIn});

  const analytics = useAnalytics({
    createdAt: Date.now(),
    email: account?.email,
    id: account?.id,
    provider: {
      getAnonymousId,
      identify,
      page,
      track,
    },
  });

  if (enableTelemetry) {
    Sentry.setUser({
      id: account?.id,
      email: account?.email,
    });
    if (account?.email) {
      LogRocket.identify(account.email);
    }

    analytics.init();
  }

  return <>{children}</>;
};

export default AnalyticsProvider;

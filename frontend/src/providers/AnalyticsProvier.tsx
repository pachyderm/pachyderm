import {useAnalytics} from '@pachyderm/components';
import * as Sentry from '@sentry/react';
import LogRocket from 'logrocket';
import React from 'react';
import {identify, page, track} from 'rudder-sdk-js';

import useAccount from '@dash-frontend/hooks/useAccount';
import useAdminInfo from '@dash-frontend/hooks/useAdminInfo';
import useAuth from '@dash-frontend/hooks/useAuth';
import {getDisableTelemetry} from '@dash-frontend/lib/runtimeVariables';

const enableTelemetry = !getDisableTelemetry();

const AnalyticsProvider: React.FC = ({children}) => {
  const {loggedIn} = useAuth();
  const {account} = useAccount({skip: !loggedIn});
  const {clusterId} = useAdminInfo({skip: !loggedIn});
  const analytics = useAnalytics({
    createdAt: Date.now(),
    email: account?.email,
    id: account?.id,
    clusterId: clusterId || undefined,
    provider: {
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

import * as Sentry from '@sentry/react';
import React from 'react';
import {identify, page, track} from 'rudder-sdk-js';

import useAccount from '@dash-frontend/hooks/useAccount';
import useAdminInfo from '@dash-frontend/hooks/useAdminInfo';
import useAuth from '@dash-frontend/hooks/useAuth';
import {getDisableTelemetry} from '@dash-frontend/lib/runtimeVariables';
import {useAnalytics} from '@pachyderm/components';

const enableTelemetry = !getDisableTelemetry();

const AnalyticsProvider = ({children}: {children?: React.ReactNode}) => {
  return enableTelemetry ? (
    <AnalyticsProviderEnabled>{children}</AnalyticsProviderEnabled>
  ) : (
    <>{children}</>
  );
};

const AnalyticsProviderEnabled = ({children}: {children?: React.ReactNode}) => {
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

  Sentry.setUser({
    id: account?.id,
    email: account?.email,
  });

  analytics.init();

  return <>{children}</>;
};

export default AnalyticsProvider;

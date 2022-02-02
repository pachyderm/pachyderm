import {useAnalytics} from '@pachyderm/components';
import React from 'react';
import {useLocation} from 'react-router';
import {getAnonymousId, identify, page, track} from 'rudder-sdk-js';

import useAccount from '@dash-frontend/hooks/useAccount';
import useAuth from '@dash-frontend/hooks/useAuth';

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

  analytics.init();

  const {search} = useLocation();
  const params = new URLSearchParams(search);
  const promoCode = params.get('utm_content');

  if (promoCode) {
    window.localStorage.setItem('promoCode', promoCode);
  }

  return <>{children}</>;
};

export default AnalyticsProvider;

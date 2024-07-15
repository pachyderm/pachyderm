import React from 'react';

import {State} from '@dash-frontend/api/enterprise';
import {useEnterpriseState} from '@dash-frontend/hooks/useEnterpriseState';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';

import CommunityEditionBanner from './CommunityEditionBanner';

const THIRTY_DAYS = 30 * 24 * 60 * 60;

const CommunityEditionBannerProvider: React.FC = () => {
  const {enterpriseState, loading, error} = useEnterpriseState();
  if (loading || error) return null;
  const active = enterpriseState?.state === State.ACTIVE;
  const expires = getUnixSecondsFromISOString(enterpriseState?.info?.expires);
  const expiring = active && expires - Date.now() / 1000 <= THIRTY_DAYS;
  if (active && !expiring) return null;

  return <CommunityEditionBanner expiration={expiring ? expires : undefined} />;
};

export default CommunityEditionBannerProvider;

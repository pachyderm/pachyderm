import {EnterpriseState} from '@graphqlTypes';
import React from 'react';

import {useGetEnterpriseInfoQuery} from '@dash-frontend/generated/hooks';

import CommunityEditionBanner from './CommunityEditionBanner';

const THIRTY_DAYS = 30 * 24 * 60 * 60;

const CommunityEditionBannerProvider: React.FC = () => {
  const {data, loading, error} = useGetEnterpriseInfoQuery();
  if (loading || error) return null;
  const active = data?.enterpriseInfo.state === EnterpriseState.ACTIVE;
  const expiring =
    active &&
    data?.enterpriseInfo?.expiration - Date.now() / 1000 <= THIRTY_DAYS;
  if (active && !expiring) return null;

  return (
    <CommunityEditionBanner
      expiration={expiring ? data?.enterpriseInfo?.expiration : undefined}
    />
  );
};

export default CommunityEditionBannerProvider;

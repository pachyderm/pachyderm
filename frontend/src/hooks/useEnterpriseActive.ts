import {EnterpriseState} from '@graphqlTypes';

import {useGetEnterpriseInfoQuery} from '@dash-frontend/generated/hooks';

export const useEnterpriseActive = () => {
  const {data, loading} = useGetEnterpriseInfoQuery();
  const enterpriseActive =
    data?.enterpriseInfo.state === EnterpriseState.ACTIVE;

  return {loading, enterpriseActive};
};

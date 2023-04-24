import {EnterpriseState} from '@graphqlTypes';

import {useGetEnterpriseInfoQuery} from '@dash-frontend/generated/hooks';

export const useEnterpriseActive = () => {
  const {data} = useGetEnterpriseInfoQuery();
  const enterpriseActive =
    data?.enterpriseInfo.state === EnterpriseState.ACTIVE;

  return {enterpriseActive};
};

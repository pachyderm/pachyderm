import {EnterpriseState} from '@graphqlTypes';

import {useGetEnterpriseInfoQuery} from '@dash-frontend/generated/hooks';

import useLoggedIn from './useLoggedIn';

export const useEnterpriseActive = (disableCheck = false) => {
  const {loggedIn} = useLoggedIn();
  const {data, loading} = useGetEnterpriseInfoQuery({
    skip: disableCheck || !loggedIn,
  });
  const enterpriseActive =
    data?.enterpriseInfo.state === EnterpriseState.ACTIVE;

  return {loading, enterpriseActive};
};

import {State} from '@dash-frontend/api/enterprise';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';

import {useEnterpriseState} from './useEnterpriseState';
import useLoggedIn from './useLoggedIn';

export const useEnterpriseActive = (disableCheck = false) => {
  const {loggedIn} = useLoggedIn();
  const {enterpriseState, error, loading} = useEnterpriseState({
    enabled: !disableCheck || loggedIn,
  });
  const enterpriseActive = enterpriseState?.state === State.ACTIVE;

  return {
    enterpriseActive,
    error: getErrorMessage(error),
    loading,
  };
};

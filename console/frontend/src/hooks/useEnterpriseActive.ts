import {State} from '@dash-frontend/api/enterprise';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';

import {useEnterpriseState} from './useEnterpriseState';
import useLoggedIn from './useLoggedIn';

export const useEnterpriseActiveDeprecated = (disableCheck = false) => {
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

// 12/6/2024 Community edition is no longer supported, and all enterprise features are enabled by default.
export const useEnterpriseActive = () => {
  return {
    enterpriseActive: true,
    error: undefined,
    loading: false,
  };
};

import {useLoginWindow} from '@pachyderm/components';
import {useEffect, useMemo} from 'react';
import {useLocation} from 'react-router';

import useAuth from '@dash-frontend/hooks/useAuth';

const useAuthenticatedRoute = () => {
  const {
    exchangeCodeError,
    loggedIn,
    exchangeCode,
    authConfig,
    authConfigError,
  } = useAuth();
  const {
    initiateOauthFlow,
    loginWindowError,
    loginWindowSucceeded,
  } = useLoginWindow({
    onSuccess: exchangeCode,
  });
  const location = useLocation();
  const error = exchangeCodeError || loginWindowError || authConfigError;

  const redirectSearchString = useMemo(() => {
    const redirect = location.pathname + location.search;
    const params = new URLSearchParams({redirect});

    return `?${params.toString()}`;
  }, [location]);

  useEffect(() => {
    if (!error && !loggedIn && !loginWindowSucceeded && authConfig) {
      initiateOauthFlow({
        authUrl: authConfig.authUrl,
        clientId: authConfig.clientId,
        scope: [
          'openid',
          'email',
          'profile',
          `audience:server:client_id:${authConfig.pachdClientId}`,
        ].join('+'),
        openWindow: false,
      });
    }
  }, [error, loggedIn, initiateOauthFlow, loginWindowSucceeded, authConfig]);

  return {
    error,
    loggedIn,
    redirectSearchString,
  };
};

export default useAuthenticatedRoute;

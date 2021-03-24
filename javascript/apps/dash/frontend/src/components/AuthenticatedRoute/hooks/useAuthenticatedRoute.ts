import {useLoginWindow} from '@pachyderm/components';
import {useEffect, useMemo} from 'react';
import {useLocation} from 'react-router';

import useAuth from '@dash-frontend/hooks/useAuth';

const pachdClientId = process.env.REACT_APP_OAUTH_PACHD_CLIENT_ID;
const authUrl = process.env.REACT_APP_OAUTH_URL || '';
const clientId = process.env.REACT_APP_OAUTH_CLIENT_ID || '';

const useAuthenticatedRoute = () => {
  const {exchangeCodeError, loggedIn, exchangeCode} = useAuth();
  const {
    initiateOauthFlow,
    loginWindowError,
    loginWindowSucceeded,
  } = useLoginWindow({
    onSuccess: exchangeCode,
  });
  const location = useLocation();
  const error = exchangeCodeError || loginWindowError;

  const redirectSearchString = useMemo(() => {
    const redirect = location.pathname + location.search;
    const params = new URLSearchParams({redirect});

    return `?${params.toString()}`;
  }, [location]);

  useEffect(() => {
    if (!error && !loggedIn && !loginWindowSucceeded) {
      initiateOauthFlow({
        authUrl,
        clientId,
        scope: [
          'openid',
          'email',
          'profile',
          `audience:server:client_id:${pachdClientId}`,
        ].join('+'),
        openWindow: false,
      });
    }
  }, [error, loggedIn, initiateOauthFlow, loginWindowSucceeded]);

  return {
    error,
    loggedIn,
    redirectSearchString,
  };
};

export default useAuthenticatedRoute;

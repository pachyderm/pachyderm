import {useLoginWindow} from '@pachyderm/components';
import {useEffect, useMemo, useState} from 'react';
import {useLocation} from 'react-router';

import useAuth from 'hooks/useAuth';

const pachdClientId = process.env.REACT_APP_OAUTH_PACHD_CLIENT_ID;
const authUrl = process.env.REACT_APP_OAUTH_URL || '';
const clientId = process.env.REACT_APP_OAUTH_CLIENT_ID || '';

const useAuthenticatedRoute = () => {
  const {exchangeCodeError, loggedIn, exchangeCode} = useAuth();
  const {initiateOauthFlow, loginWindowError} = useLoginWindow({
    onSuccess: exchangeCode,
  });
  const location = useLocation();
  const [authenticating] = useState(() =>
    Boolean(window.localStorage.getItem('oauthCode')),
  );

  const error = exchangeCodeError || loginWindowError;

  const redirectSearchString = useMemo(() => {
    const redirect = location.pathname + location.search;
    const params = new URLSearchParams({redirect});

    return `?${params.toString()}`;
  }, [location]);

  useEffect(() => {
    if (!error && !loggedIn && !authenticating) {
      initiateOauthFlow({
        authUrl,
        clientId,
        scope: `openid+email+audience:server:client_id:${pachdClientId}`,
        openWindow: false,
      });
    }
  }, [error, loggedIn, initiateOauthFlow, authenticating]);

  return {
    error,
    loggedIn,
    redirectSearchString,
  };
};

export default useAuthenticatedRoute;

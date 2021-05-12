import {useLoginWindow} from '@pachyderm/components';
import {useEffect, useMemo} from 'react';
import {useLocation, useHistory} from 'react-router';

import useAuth from '@dash-frontend/hooks/useAuth';

const useAuthenticatedRoute = () => {
  const {
    exchangeCodeError,
    loggedIn,
    exchangeCode,
    authConfig,
    authConfigError,
  } = useAuth();
  const {initiateOauthFlow, loginWindowError, loginWindowSucceeded} =
    useLoginWindow({
      onSuccess: exchangeCode,
    });
  const {search} = useLocation();
  const routerHistory = useHistory();
  const params = useMemo(() => new URLSearchParams(search), [search]);
  const error = exchangeCodeError || loginWindowError || authConfigError;

  const workspaceParam = params.get('workspaceName');
  if (workspaceParam) {
    window.localStorage.setItem('workspaceName', workspaceParam);
    routerHistory.push('/');
  }

  const redirectSearchString = `?${params.toString()}`;

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

import {useLoginWindow} from '@pachyderm/components';
import {useEffect, useMemo} from 'react';
import {useLocation, useHistory} from 'react-router';

import useAuth from '@dash-frontend/hooks/useAuth';
import {getIssuerUri} from '@dash-frontend/lib/runtimeVariables';

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
  const pachVersion = params.get('pachVersion');
  const pachdAddress = params.get('pachdAddress');

  if (workspaceParam || pachVersion || pachdAddress) {
    workspaceParam &&
      window.localStorage.setItem('workspaceName', workspaceParam);
    pachVersion && window.localStorage.setItem('pachVersion', pachVersion);
    pachdAddress && window.localStorage.setItem('pachdAddress', pachdAddress);
    routerHistory.push('/');
  }

  const loginHintParam = params.get('login_hint');
  const connectionParam = params.get('connection');
  const redirectSearchString = `?${params.toString()}`;

  useEffect(() => {
    if (!error && !loggedIn && !loginWindowSucceeded && authConfig) {
      const issuerUri = getIssuerUri();

      let connection = undefined;
      if (connectionParam) {
        connection = connectionParam;
      }

      let loginHint = undefined;
      if (loginHintParam) {
        loginHint = loginHintParam;
      }

      const authUrl = new URL(issuerUri);

      if (authUrl.pathname.endsWith('/')) {
        // remove leading slashes
        authUrl.pathname += authConfig.authEndpoint.replace(/^\/+/g, '');
      } else {
        authUrl.pathname += authConfig.authEndpoint;
      }

      initiateOauthFlow({
        authUrl: authUrl.toString(),
        clientId: authConfig.clientId,
        scope: [
          'openid',
          'email',
          'profile',
          `audience:server:client_id:${authConfig.pachdClientId}`,
        ].join('+'),
        openWindow: false,
        connection,
        loginHint,
      });
    }
  }, [
    error,
    loggedIn,
    initiateOauthFlow,
    loginWindowSucceeded,
    authConfig,
    loginHintParam,
    connectionParam,
  ]);

  return {
    error,
    loggedIn,
    redirectSearchString,
  };
};

export default useAuthenticatedRoute;

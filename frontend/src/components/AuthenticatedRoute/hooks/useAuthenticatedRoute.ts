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
      const {REACT_APP_RUNTIME_ISSUER_URI, pachDashConfig} = process.env;

      // when running in the node runtime environment,
      // environment variables can't be objects. Therefore
      // we'll need jest to override this in order to be set.
      let issuerUri = '';
      if (REACT_APP_RUNTIME_ISSUER_URI) {
        issuerUri = REACT_APP_RUNTIME_ISSUER_URI;
      } else {
        issuerUri = pachDashConfig.REACT_APP_RUNTIME_ISSUER_URI;
      }

      let connection = undefined;
      if (connectionParam) {
        connection = connectionParam;
      }

      let loginHint = undefined;
      if (loginHintParam) {
        loginHint = loginHintParam;
      }

      initiateOauthFlow({
        authUrl: `${issuerUri}${authConfig.authEndpoint}`,
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

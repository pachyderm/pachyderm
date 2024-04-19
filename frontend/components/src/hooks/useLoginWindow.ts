import noop from 'lodash/noop';
import {useCallback, useEffect, useState} from 'react';

import getRandomString from '@pachyderm/components/lib/getRandomString';

type UseLoginWindowProps = {
  onSuccess?: (code: string) => void;
};

interface InitiateOauthFlowArgs {
  authUrl: string;
  clientId: string;
  redirect?: string;
  openWindow?: boolean;
  connection?: string;
  scope?: string;
  loginHint?: string;
}

const useLoginWindow = ({onSuccess = noop}: UseLoginWindowProps = {}) => {
  const [error, setError] = useState<string | null>(null);
  const [succeeded, setSucceeded] = useState<boolean>(false);
  const [loginWindow, setLoginWindow] = useState<Window | null>(null);

  const initiateOauthFlow = useCallback(
    ({
      authUrl,
      clientId,
      redirect = `${window.location.origin}/oauth/callback/`,
      openWindow = true,
      connection,
      scope = 'openid+profile+email+user_id',
      loginHint,
    }: InitiateOauthFlowArgs) => {
      if (loginWindow && !loginWindow.closed && openWindow) {
        return;
      }

      window.localStorage.removeItem('oauthState');
      window.localStorage.removeItem('oauthError');

      const x = window.screenX + (window.outerWidth - 500) / 2;
      const y = window.screenY + (window.outerHeight - 500) / 2.5;
      const nonce = getRandomString(20);
      const features = `width=${500},height=${500},left=${x},top=${y}`;

      window.localStorage.setItem(
        `${nonce}_original_page`,
        window.location.pathname + window.location.search,
      );

      let url = `${authUrl}?${[
        `client_id=${clientId}`,
        `redirect_uri=${openWindow ? redirect : `${redirect}?inline=true`}`,
        `response_type=code`,
        `scope=${scope}`,
        `state=${nonce}`,
      ].join('&')}`;

      if (connection) {
        url += `&connection=${connection}`;
      }

      if (loginHint) {
        url += `&login_hint=${loginHint}`;
      }

      window.localStorage.setItem('oauthState', nonce);

      // This flag is for use by the e2e testing suite
      if (openWindow) {
        setLoginWindow(window.open(url, '', features));
      } else {
        window.location.assign(url);
      }
    },
    [loginWindow, setLoginWindow],
  );

  const extractStateFromLocalStorage = useCallback(() => {
    const oauthCode = window.localStorage.getItem('oauthCode');
    const oauthError = window.localStorage.getItem('oauthError');

    window.localStorage.removeItem('oauthCode');
    window.localStorage.removeItem('oauthError');

    if (oauthCode) {
      onSuccess(oauthCode);
      setSucceeded(true);
    } else if (oauthError) {
      setError(oauthError);
    }

    return {oauthCode, oauthError};
  }, [onSuccess]);

  useEffect(() => {
    let interval: NodeJS.Timeout | undefined;

    let {oauthCode, oauthError} = extractStateFromLocalStorage();

    if (!interval && loginWindow) {
      interval = setInterval(() => {
        const results = extractStateFromLocalStorage();

        oauthCode = results.oauthCode;
        oauthError = results.oauthError;

        if (interval && (oauthCode || oauthError)) {
          clearInterval(interval);
        }
      }, 500);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [loginWindow, onSuccess, setError, extractStateFromLocalStorage]);

  return {
    loginWindowSucceeded:
      succeeded || Boolean(window.localStorage.getItem('oauthCode')),
    loginWindowError: error || window.localStorage.getItem('oauthError'),
    initiateOauthFlow,
  };
};

export default useLoginWindow;

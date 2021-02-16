import noop from 'lodash/noop';
import {useCallback, useEffect, useState} from 'react';

import getRandomString from 'lib/getRandomString';

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
}

const useLoginWindow = ({onSuccess = noop}: UseLoginWindowProps = {}) => {
  const [error, setError] = useState<string | null>(null);
  const [loginWindow, setLoginWindow] = useState<Window | null>(null);

  const initiateOauthFlow = useCallback(
    ({
      authUrl,
      clientId,
      redirect = `${window.location.origin}/oauth/callback/`,
      openWindow = true,
      connection,
      scope = 'openid+profile+email+user_id',
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

  useEffect(() => {
    let interval: NodeJS.Timeout | undefined;

    let oauthCode = window.localStorage.getItem('oauthCode');
    let oauthError = window.localStorage.getItem('oauthError');

    window.localStorage.removeItem('oauthCode');
    window.localStorage.removeItem('oauthError');

    if (oauthCode) {
      onSuccess(oauthCode);
    } else if (oauthError) {
      setError(oauthError);
    }

    if (!interval && loginWindow) {
      interval = setInterval(() => {
        oauthCode = window.localStorage.getItem('oauthCode');
        oauthError = window.localStorage.getItem('oauthError');

        window.localStorage.removeItem('oauthCode');
        window.localStorage.removeItem('oauthError');

        if (oauthCode) {
          onSuccess(oauthCode);
        } else if (oauthError) {
          setError(oauthError);
        }

        if (interval && (oauthCode || oauthError)) {
          clearInterval(interval);
        }
      }, 500);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [loginWindow, onSuccess, setError]);

  return {
    loginWindowError: error,
    initiateOauthFlow,
  };
};

export default useLoginWindow;

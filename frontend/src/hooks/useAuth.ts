import {ApolloError} from '@apollo/client';
import {MutationExchangeCodeArgs} from '@graphqlTypes';
import Cookies from 'js-cookie';
import noop from 'lodash/noop';
import {useCallback, useEffect, useMemo} from 'react';

import {useExchangeCodeMutation} from '@dash-frontend/generated/hooks';
import useAuthConfig from '@dash-frontend/hooks/useAuthConfig';

import useLoggedIn from './useLoggedIn';

const COOKIE_EXPIRES = 365; // TODO: align with api token expiry
const IS_DEV = process.env.NODE_ENV === 'development';

interface UseAuthArgs {
  onError?: (error: ApolloError) => void;
}

const useAuth = ({onError = noop}: UseAuthArgs = {}) => {
  const {loggedIn, setLoggedIn} = useLoggedIn();
  const {authConfig, error: authConfigError} = useAuthConfig({
    skip: loggedIn,
  });
  const [exchangeCodeMutation, {data: codeMutationData, error, loading}] =
    useExchangeCodeMutation({onError});

  const unauthenticated = useMemo(() => {
    if (authConfig) {
      return Object.values(authConfig)
        .filter((val) => val !== 'AuthConfig')
        .every((val) => val === '');
    } else return false;
  }, [authConfig]);

  const exchangeCode = useCallback(
    (code: MutationExchangeCodeArgs['code']) => {
      exchangeCodeMutation({
        variables: {
          code,
        },
      });
    },
    [exchangeCodeMutation],
  );

  useEffect(() => {
    if (codeMutationData) {
      window.localStorage.setItem(
        'auth-token',
        codeMutationData.exchangeCode.pachToken,
      );
      window.localStorage.setItem(
        'id-token',
        codeMutationData.exchangeCode.idToken,
      );

      // We need these cookies to enable file downloads and previews
      Cookies.set(
        'dashAuthToken',
        window.localStorage.getItem('auth-token') || '',
        {
          expires: COOKIE_EXPIRES,
          sameSite: IS_DEV ? 'lax' : 'strict',
          secure: window.location.protocol === 'https:',
        },
      );

      setLoggedIn(
        Boolean(
          window.localStorage.getItem('auth-token') !== null &&
            window.localStorage.getItem('id-token') !== null,
        ),
      );
    } else if (authConfig && unauthenticated) {
      window.localStorage.setItem('auth-token', '');
      window.localStorage.setItem('id-token', '');

      // We need these cookies to enable file downloads and previews
      Cookies.set(
        'dashAuthToken',
        window.localStorage.getItem('auth-token') || '',
        {
          expires: COOKIE_EXPIRES,
          sameSite: IS_DEV ? 'lax' : 'strict',
          secure: window.location.protocol === 'https:',
        },
      );

      setLoggedIn(
        Boolean(
          window.localStorage.getItem('auth-token') !== null &&
            window.localStorage.getItem('id-token') !== null,
        ),
      );
    }
  }, [codeMutationData, setLoggedIn, authConfig, unauthenticated]);

  return {
    exchangeCode,
    exchangeCodeError: error,
    exchangeCodeLoading: loading,
    loggedIn,
    authConfig,
    authConfigError,
    unauthenticated,
  };
};

export default useAuth;

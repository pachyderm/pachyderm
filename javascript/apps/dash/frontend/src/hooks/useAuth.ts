import {ApolloError, useApolloClient} from '@apollo/client';
import Cookies from 'js-cookie';
import noop from 'lodash/noop';
import {useCallback, useEffect} from 'react';

import {useExchangeCodeMutation} from '@dash-frontend/generated/hooks';
import useAuthConfig from '@dash-frontend/hooks/useAuthConfig';
import useSubscriptionClient from '@dash-frontend/providers/SubscriptionClientProvider/hooks/useSubscriptionClient';
import {GET_LOGGED_IN_QUERY} from '@dash-frontend/queries/GetLoggedInQuery';
import {MutationExchangeCodeArgs} from '@graphqlTypes';

import useLoggedIn from './useLoggedIn';

const COOKIE_EXPIRES = 365; // TODO: align with api token expiry
const IS_DEV = process.env.NODE_ENV === 'development';

interface UseAuthArgs {
  onError?: (error: ApolloError) => void;
}

const useAuth = ({onError = noop}: UseAuthArgs = {}) => {
  const client = useApolloClient();
  const {client: subscriptionClient} = useSubscriptionClient();
  const loggedIn = useLoggedIn();
  const {authConfig, error: authConfigError} = useAuthConfig({
    skip: loggedIn,
  });
  const [exchangeCodeMutation, {data: codeMutationData, error, loading}] =
    useExchangeCodeMutation({onError});

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
          secure: true,
        },
      );

      client.writeQuery({
        query: GET_LOGGED_IN_QUERY,
        data: {
          loggedIn: Boolean(
            window.localStorage.getItem('auth-token') &&
              window.localStorage.getItem('id-token'),
          ),
        },
      });

      // reset websocket connection to add new auth info
      subscriptionClient?.close(false, false);
    }
  }, [client, codeMutationData, subscriptionClient]);

  return {
    exchangeCode,
    exchangeCodeError: error,
    exchangeCodeLoading: loading,
    loggedIn,
    authConfig,
    authConfigError,
  };
};

export default useAuth;

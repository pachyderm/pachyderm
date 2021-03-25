import {ApolloError, gql, useApolloClient, useQuery} from '@apollo/client';
import Cookies from 'js-cookie';
import noop from 'lodash/noop';
import {useCallback, useEffect} from 'react';

import {useExchangeCodeMutation} from '@dash-frontend/generated/hooks';
import {MutationExchangeCodeArgs} from '@graphqlTypes';

interface LogInResponse {
  loggedIn: false;
}

export const LOGGED_IN_QUERY = gql`
  query Login {
    loggedIn @client
  }
`;

const COOKIE_EXPIRES = 365; // TODO: align with api token expiry

interface UseAuthArgs {
  onError?: (error: ApolloError) => void;
}

const useAuth = ({onError = noop}: UseAuthArgs = {}) => {
  const client = useApolloClient();
  const [
    exchangeCodeMutation,
    {data: codeMutationData, error, loading},
  ] = useExchangeCodeMutation({onError});
  const {data: loggedInData} = useQuery<LogInResponse>(LOGGED_IN_QUERY);

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
          httpOnly: true,
          sameSite: 'strict',
          secure: true,
        },
      );
      Cookies.set('dashAddress', process.env.REACT_APP_PACHD_ADDRESS || '', {
        expires: COOKIE_EXPIRES,
        httpOnly: true,
        sameSite: 'strict',
        secure: true,
      });

      client.writeQuery({
        query: LOGGED_IN_QUERY,
        data: {
          loggedIn: Boolean(
            window.localStorage.getItem('auth-token') &&
              window.localStorage.getItem('id-token'),
          ),
        },
      });
    }
  }, [client, codeMutationData]);

  return {
    exchangeCode,
    exchangeCodeError: error,
    exchangeCodeLoading: loading,
    loggedIn: Boolean(loggedInData?.loggedIn),
  };
};

export default useAuth;

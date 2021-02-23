import {gql, useApolloClient, useMutation, useQuery} from '@apollo/client';
import {useCallback, useEffect} from 'react';

import {MutationExchangeCodeArgs} from '@graphqlTypes';
import {EXCHANGE_CODE_MUTATION} from 'mutations/ExchangeCode';

interface ExchangeCodeMutationResults {
  exchangeCode: string;
}

interface LogInResponse {
  loggedIn: false;
}

export const LOGGED_IN_QUERY = gql`
  query Login {
    loggedIn @client
  }
`;

const useAuth = () => {
  const client = useApolloClient();
  const [
    exchangeCodeMutation,
    {data: codeMutationData, error, loading},
  ] = useMutation<ExchangeCodeMutationResults, MutationExchangeCodeArgs>(
    EXCHANGE_CODE_MUTATION,
  );
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
      window.localStorage.setItem('auth-token', codeMutationData.exchangeCode);

      client.writeQuery({
        query: LOGGED_IN_QUERY,
        data: {
          loggedIn: Boolean(window.localStorage.getItem('auth-token')),
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

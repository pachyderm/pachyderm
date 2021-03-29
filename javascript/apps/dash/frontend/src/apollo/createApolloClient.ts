import {
  ApolloClient,
  InMemoryCache,
  NormalizedCacheObject,
  ApolloLink,
} from '@apollo/client';
import {History as BrowserHistory} from 'history';

import {errorLink} from '@dash-frontend/apollo/links/errorLink';
import {GET_LOGGED_IN_QUERY} from '@dash-frontend/queries/GetLoggedInQuery';

import {contextLink} from './links/contextLink';
import {httpLink} from './links/httpLink';

const createApolloClient = (
  browserHistory: BrowserHistory,
): ApolloClient<NormalizedCacheObject> => {
  const cache = new InMemoryCache();

  const link = ApolloLink.from([contextLink(), errorLink(), httpLink()]);

  const resolvers = {};

  const client = new ApolloClient({cache, link, resolvers});

  client.writeQuery({
    data: {
      loggedIn: Boolean(
        window.localStorage.getItem('auth-token') &&
          window.localStorage.getItem('id-token'),
      ),
    },
    query: GET_LOGGED_IN_QUERY,
  });

  client.onResetStore(() =>
    Promise.resolve(
      client.writeQuery({
        data: {
          loggedIn: Boolean(
            window.localStorage.getItem('auth-token') &&
              window.localStorage.getItem('id-token'),
          ),
        },
        query: GET_LOGGED_IN_QUERY,
      }),
    ),
  );

  return client;
};

export default createApolloClient;

import {
  ApolloClient,
  InMemoryCache,
  NormalizedCacheObject,
  ApolloLink,
} from '@apollo/client';
import {History as BrowserHistory} from 'history';

import {LOGGED_IN_QUERY} from '@dash-frontend/hooks/useAuth';
import {httpLink} from 'apollo/links/httpLink';

import {contextLink} from './links/contextLink';

const createApolloClient = (
  browserHistory: BrowserHistory,
): ApolloClient<NormalizedCacheObject> => {
  const cache = new InMemoryCache();

  const link = ApolloLink.from([contextLink(), httpLink()]);

  const resolvers = {};

  const client = new ApolloClient({cache, link, resolvers});

  client.writeQuery({
    data: {
      loggedIn: Boolean(window.localStorage.getItem('auth-token')),
    },
    query: LOGGED_IN_QUERY,
  });

  client.onResetStore(() =>
    Promise.resolve(
      client.writeQuery({
        data: {
          loggedIn: Boolean(window.localStorage.getItem('auth-token')),
        },
        query: LOGGED_IN_QUERY,
      }),
    ),
  );

  return client;
};

export default createApolloClient;

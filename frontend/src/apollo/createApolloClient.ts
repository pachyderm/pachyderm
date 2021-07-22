import {
  ApolloClient,
  InMemoryCache,
  NormalizedCacheObject,
  ApolloLink,
} from '@apollo/client';
import {sentryLink} from '@pachyderm/components';
import {History as BrowserHistory} from 'history';

import {errorLink} from '@dash-frontend/apollo/links/errorLink';
import {GET_LOGGED_IN_QUERY} from '@dash-frontend/queries/GetLoggedInQuery';
import {LoggedInQuery} from '@graphqlTypes';

import cacheConfig from './cacheConfig';
import {contextLink} from './links/contextLink';
import {splitLink} from './links/splitLink';

const createApolloClient = (
  browserHistory: BrowserHistory,
): {
  client: ApolloClient<NormalizedCacheObject>;
} => {
  const cache = new InMemoryCache(cacheConfig);
  const {split, restartWebsocket} = splitLink();

  const link = ApolloLink.from([
    contextLink(),
    sentryLink(),
    errorLink(browserHistory),
    split,
  ]);
  const resolvers = {};

  const client = new ApolloClient({cache, link, resolvers});

  // restart websocket to update connectionParams with new auth
  let prevLoggedInValue = false;
  client.watchQuery<LoggedInQuery>({query: GET_LOGGED_IN_QUERY}).subscribe({
    next: (data) => {
      if (data.data.loggedIn !== prevLoggedInValue) {
        prevLoggedInValue = data.data.loggedIn;
        restartWebsocket();
      }
    },
  });

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

  return {client};
};

export default createApolloClient;

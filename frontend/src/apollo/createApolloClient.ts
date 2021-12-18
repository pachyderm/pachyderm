import {ApolloClient, InMemoryCache, ApolloLink} from '@apollo/client';
import {sentryLink} from '@pachyderm/components';
import {History as BrowserHistory} from 'history';

import {errorLink} from '@dash-frontend/apollo/links/errorLink';

import cacheConfig from './cacheConfig';
import {contextLink} from './links/contextLink';
import {splitLink} from './links/splitLink';

const createApolloClient = (browserHistory: BrowserHistory) => {
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

  return {client, restartWebsocket};
};

export default createApolloClient;

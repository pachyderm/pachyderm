import {
  ApolloClient,
  InMemoryCache,
  NormalizedCacheObject,
  ApolloLink,
} from '@apollo/client';
import {History as BrowserHistory} from 'history';

import {httpLink} from 'apollo/links/httpLink';

import {contextLink} from './links/contextLink';

const createApolloClient = (
  browserHistory: BrowserHistory,
): ApolloClient<NormalizedCacheObject> => {
  const cache = new InMemoryCache();

  const link = ApolloLink.from([contextLink(), httpLink()]);

  const resolvers = {};

  const client = new ApolloClient({cache, link, resolvers});

  return client;
};

export default createApolloClient;

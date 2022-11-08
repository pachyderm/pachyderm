import {ApolloClient, InMemoryCache, ApolloLink} from '@apollo/client';

import {sentryLink} from '@pachyderm/components';

import cacheConfig from './cacheConfig';
import {contextLink} from './links/contextLink';
import {retryLink} from './links/retryLink';
import {splitLink} from './links/splitLink';

const createApolloClient = (errorLink: ApolloLink) => {
  const cache = new InMemoryCache(cacheConfig);
  const {split, restartWebsocket} = splitLink();

  const link = ApolloLink.from([
    contextLink(),
    sentryLink(),
    errorLink,
    retryLink(),
    split,
  ]);
  const resolvers = {};

  const client = new ApolloClient({cache, link, resolvers});

  return {client, restartWebsocket};
};

export default createApolloClient;

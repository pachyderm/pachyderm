import {InMemoryCacheConfig} from '@apollo/client';
import {Vertex} from '@graphqlTypes';

const cacheConfig: InMemoryCacheConfig = {
  typePolicies: {
    Subscription: {
      fields: {
        dags: {
          merge(_existing: Vertex[], incoming: Vertex[]) {
            return incoming;
          },
        },
      },
    },
    Job: {
      // This is important, as a Job ID is not globally unique. However,
      // the combination of both id and pipelineName is.
      keyFields: ['id', 'pipelineName'],
    },
    Commit: {
      keyFields: ['id', 'repoName'],
    },
  },
};

export default cacheConfig;

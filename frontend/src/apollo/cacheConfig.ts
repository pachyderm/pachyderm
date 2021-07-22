import {InMemoryCacheConfig} from '@apollo/client';

import {Dag, Link} from '@graphqlTypes';

const cacheConfig: InMemoryCacheConfig = {
  typePolicies: {
    Subscription: {
      fields: {
        dags: {
          merge(_existing: Dag[], incoming: Dag[]) {
            return incoming;
          },
        },
      },
    },
    Dag: {
      fields: {
        nodes: {
          merge(_existing: Node[], incoming: Node[]) {
            return incoming;
          },
        },
        links: {
          merge(_existing: Link[], incoming: Link[]) {
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

import {FieldFunctionOptions, InMemoryCacheConfig} from '@apollo/client';
import {Vertex} from '@graphqlTypes';

const mergeJson = (
  existing: string,
  incoming: string,
  mergeObjects: FieldFunctionOptions['mergeObjects'],
) => {
  let existingObj, incomingObj;

  try {
    existingObj = existing ? JSON.parse(existing) : undefined;
    incomingObj = incoming ? JSON.parse(incoming) : undefined;
  } catch (e) {
    console.error(e);
  }

  if (existingObj && incomingObj) {
    const mergedObj = mergeObjects(existingObj, incomingObj);
    return JSON.stringify(mergedObj);
  } else if (!incoming) {
    return existing;
  } else {
    return incoming;
  }
};

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
    Datum: {
      // This is important, as a Datum ID is not globally unique. However,
      // the combination of both a Datum ID and Job ID is.
      keyFields: ['id', 'requestedJobId'],
    },
    Job: {
      // This is important, as a Job ID is not globally unique. However,
      // the combination of both id and pipelineName is.
      keyFields: ['id', 'pipelineName'],
      fields: {
        inputString: {
          merge(existing, incoming, {mergeObjects}) {
            return mergeJson(existing, incoming, mergeObjects);
          },
        },
        transformString: {
          merge(existing, incoming, {mergeObjects}) {
            return mergeJson(existing, incoming, mergeObjects);
          },
        },
        transform: {
          merge(existing, incoming, {mergeObjects}) {
            return mergeObjects(existing, incoming);
          },
        },
      },
    },
    Commit: {
      keyFields: ['id', 'repoName'],
    },
    Branch: {
      keyFields: ['name'],
    },
    Query: {
      fields: {
        dag: {
          merge(_existing: Vertex[], incoming: Vertex[]) {
            return incoming;
          },
        },
      },
    },
  },
};

export default cacheConfig;

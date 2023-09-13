import uniqueId from 'lodash/uniqueId';
import objectHash from 'object-hash';

import {PachClient} from '@dash-backend/lib/types';
import withSubscription from '@dash-backend/lib/withSubscription';
import {QueryResolvers, SubscriptionResolvers, Vertex} from '@graphqlTypes';

import {
  getProjectVertices,
  getProjectGlobalIdVertices,
} from './vertices/createVertices';

interface VerticesResolver {
  Query: {
    vertices: QueryResolvers['vertices'];
  };
  Subscription: {
    vertices: SubscriptionResolvers['vertices'];
  };
}

/**
 * hashHasChanged mutates prevDataHash
 */
const hashHasChanged = (data: Vertex[], prevDataHashRef: [string]) => {
  const dataHash = objectHash(data, {
    unorderedArrays: true,
  });

  if (prevDataHashRef[0] === dataHash) {
    return false;
  }

  prevDataHashRef[0] = dataHash;
  return true;
};

const verticesSubscriptionResolver = async ({
  projectId,
  jobSetId,
  pachClient,
  prevDataHashRef,
}: {
  projectId: string;
  jobSetId?: string | null;
  pachClient: PachClient;
  prevDataHashRef: [string];
}) => {
  let data: Vertex[] = [];

  if (jobSetId) {
    data = await getProjectGlobalIdVertices({
      jobSetId,
      projectId,
      pachClient,
    });
  } else {
    data = await getProjectVertices({
      projectId,
      pachClient,
    });
  }

  // prevDataHashRef is passed by reference and mutated to the current hash
  if (!hashHasChanged(data, prevDataHashRef)) {
    return;
  }

  return data;
};

const verticesResolver: VerticesResolver = {
  Query: {
    vertices: async (_field, {args: {projectId}}, {pachClient}) => {
      return await getProjectVertices({projectId, pachClient});
    },
  },
  Subscription: {
    vertices: {
      subscribe: (_field, {args: {projectId, jobSetId}}, {pachClient}) => {
        const id = uniqueId();
        // prevDataHashRef is a string passed as a reference so that it can be mutated by sub-functions
        const prevDataHashRef: [string] = [''];

        return {
          [Symbol.asyncIterator]: () =>
            withSubscription<Vertex[]>({
              triggerName: `${id}_VERTICES_UPDATED${
                jobSetId ? `_JOB_${jobSetId}` : ''
              }`,
              resolver: () =>
                verticesSubscriptionResolver({
                  projectId,
                  jobSetId,
                  pachClient,
                  prevDataHashRef,
                }),
              intervalKey: id,
            }),
        };
      },
      resolve: (result: Vertex[]) => {
        return result;
      },
    },
  },
};

export default verticesResolver;

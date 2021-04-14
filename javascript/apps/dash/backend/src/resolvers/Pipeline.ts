import client from '@dash-backend/grpc/client';
import {QueryResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';

interface PipelineResolver {
  Query: {
    pipeline: QueryResolvers['pipeline'];
  };
}

const pipelineResolver: PipelineResolver = {
  Query: {
    pipeline: async (
      _field,
      {args: {projectId, id}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      const pachClient = client({pachdAddress, projectId, authToken, log});

      const pipeline = await pachClient.pps().inspectPipeline(id);

      return pipelineInfoToGQLPipeline(pipeline);
    },
  },
};

export default pipelineResolver;

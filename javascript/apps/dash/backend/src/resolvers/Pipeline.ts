import {QueryResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';

interface PipelineResolver {
  Query: {
    pipeline: QueryResolvers['pipeline'];
  };
}

const pipelineResolver: PipelineResolver = {
  Query: {
    pipeline: async (_field, {args: {id}}, {pachClient}) => {
      const pipeline = await pachClient.pps().inspectPipeline(id);

      return pipelineInfoToGQLPipeline(pipeline);
    },
  },
};

export default pipelineResolver;

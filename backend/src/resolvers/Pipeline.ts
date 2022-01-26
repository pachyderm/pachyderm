import {MutationResolvers, QueryResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';

interface PipelineResolver {
  Query: {
    pipeline: QueryResolvers['pipeline'];
    pipelines: QueryResolvers['pipelines'];
  };
  Mutation: {
    createPipeline: MutationResolvers['createPipeline'];
  };
}

const pipelineResolver: PipelineResolver = {
  Query: {
    pipeline: async (_field, {args: {id}}, {pachClient}) => {
      const pipeline = await pachClient.pps().inspectPipeline(id);

      return pipelineInfoToGQLPipeline(pipeline);
    },
    pipelines: async (_parent, _args, {pachClient}) => {
      return (await pachClient.pps().listPipeline()).map((pipeline) =>
        pipelineInfoToGQLPipeline(pipeline),
      );
    },
  },
  Mutation: {
    createPipeline: async (
      _field,
      {args: {name, transform, pfs, crossList, description, update}},
      {pachClient},
    ) => {
      const crossListInputs = (crossList || []).map((input) => {
        return {
          pfs: {
            name: input.name,
            glob: input.glob || '/',
            repo: input.repo.name,
            branch: input.branch || 'master',
          },
        };
      });
      const pfsInput = pfs
        ? {
            name: pfs.name,
            glob: pfs.glob || '/',
            repo: pfs.repo.name,
            branch: pfs.branch || 'master',
          }
        : undefined;

      await pachClient.pps().createPipeline({
        pipeline: {name},
        transform: {
          cmdList: transform.cmdList,
          image: transform.image,
          stdinList:
            transform.stdinList && transform.stdinList.length
              ? transform.stdinList
              : undefined,
        },
        description: description || undefined,
        input: {
          crossList: crossListInputs,
          pfs: pfsInput,
        },
        update: update || undefined,
      });

      const pipeline = await pachClient.pps().inspectPipeline(name);
      return pipelineInfoToGQLPipeline(pipeline);
    },
  },
};

export default pipelineResolver;

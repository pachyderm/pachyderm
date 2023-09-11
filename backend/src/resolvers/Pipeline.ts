import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {MutationResolvers, QueryResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';
interface PipelineResolver {
  Query: {
    pipeline: QueryResolvers['pipeline'];
    pipelines: QueryResolvers['pipelines'];
    getClusterDefaults: QueryResolvers['getClusterDefaults'];
  };
  Mutation: {
    createPipeline: MutationResolvers['createPipeline'];
    createPipelineV2: MutationResolvers['createPipelineV2'];
    deletePipeline: MutationResolvers['deletePipeline'];
    setClusterDefaults: MutationResolvers['setClusterDefaults'];
  };
}

const pipelineResolver: PipelineResolver = {
  Query: {
    pipeline: async (
      _field,
      {args: {id: pipelineId, projectId}},
      {pachClient},
    ) => {
      const pipeline = await pachClient.pps.inspectPipeline({
        pipelineId,
        projectId,
      });

      return pipelineInfoToGQLPipeline(pipeline);
    },
    pipelines: async (
      _parent,
      {args: {jobSetId, projectIds}},
      {pachClient},
    ) => {
      let jq = '';

      if (jobSetId) {
        const jobs = await getJobsFromJobSet({
          jobSet: await pachClient.pps.inspectJobSet({
            id: jobSetId,
            projectId: projectIds[0] || '',
          }),
          projectId: projectIds[0] || '',
          pachClient,
        });

        jq = `select(${jobs
          .map((job) => `.pipeline.name == "${job.job?.pipeline?.name}"`)
          .join(' or ')})`;
      }

      return (await pachClient.pps.listPipeline({jq, projectIds})).map(
        (pipeline) => pipelineInfoToGQLPipeline(pipeline),
      );
    },
    getClusterDefaults: async (_field, _args, {pachClient}) => {
      return await pachClient.pps.getClusterDefaults();
    },
  },
  Mutation: {
    createPipeline: async (
      _field,
      {args: {name, transform, pfs, crossList, description, update, projectId}},
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

      await pachClient.pps.createPipeline({
        pipeline: {name, project: {name: projectId}},
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
        dryRun: false,
      });

      const pipeline = await pachClient.pps.inspectPipeline({
        pipelineId: name,
        projectId,
      });
      return pipelineInfoToGQLPipeline(pipeline);
    },
    createPipelineV2: async (_field, {args}, {pachClient}) => {
      return await pachClient.pps.createPipelineV2(args);
    },
    deletePipeline: async (_field, {args: {projectId, name}}, {pachClient}) => {
      await pachClient.pps.deletePipeline({projectId, pipeline: {name}});
      return true;
    },
    setClusterDefaults: async (_field, {args}, {pachClient}) => {
      return await pachClient.pps.setClusterDefaults(args);
    },
  },
};

export default pipelineResolver;

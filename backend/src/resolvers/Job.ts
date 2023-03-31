import {ApolloError} from 'apollo-server-errors';

import {DEFAULT_JOBS_LIMIT} from '@dash-backend/constants/limits';
import {QueryResolvers} from '@dash-backend/generated/types';
import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';

import {
  jobInfosToGQLJobSet,
  jobInfoToGQLJob,
  jobSetsToGQLJobSets,
} from './builders/pps';
interface PipelineJobResolver {
  Query: {
    job: QueryResolvers['job'];
    jobs: QueryResolvers['jobs'];
    jobsByPipeline: QueryResolvers['jobsByPipeline'];
    jobSet: QueryResolvers['jobSet'];
    jobSets: QueryResolvers['jobSets'];
  };
}

const pipelineJobResolver: PipelineJobResolver = {
  Query: {
    job: async (
      _parent,
      {args: {id, pipelineName, projectId}},
      {pachClient},
    ) => {
      return jobInfoToGQLJob(
        await pachClient.pps().inspectJob({id, pipelineName, projectId}),
      );
    },
    jobs: async (
      _parent,
      {
        args: {
          limit,
          pipelineId,
          pipelineIds,
          jobSetIds,
          projectId,
          cursor,
          reverse,
        },
      },
      {pachClient},
    ) => {
      let jqFilter = '';
      limit = limit || DEFAULT_JOBS_LIMIT;

      if (
        jobSetIds &&
        jobSetIds.length > 0 &&
        pipelineIds &&
        pipelineIds.length > 0
      ) {
        throw new ApolloError(
          'Cannot filter by both pipelineIds and jobSetIds',
        );
      }

      if (jobSetIds && jobSetIds.length > 0) {
        jqFilter = `select(${jobSetIds
          .map((jobSetId) => `.job.id == "${jobSetId}"`)
          .join(' or ')})`;
      }

      if (pipelineIds && pipelineIds.length > 0) {
        jqFilter = `select(${pipelineIds
          .map((pipeline) => `.job.pipeline.name == "${pipeline}"`)
          .join(' or ')})`;
      }

      const jobs = await pachClient.pps().listJobs({
        pipelineId,
        jqFilter,
        projectId,
        cursor: cursor || undefined,
        number: limit + 1,
        reverse: reverse || undefined,
      });

      let nextCursor = undefined;

      //If jobs.length is not greater than limit there are no pages left
      if (jobs.length > limit) {
        jobs.pop(); //remove the extra job from the response
        nextCursor = jobs[jobs.length - 1].started;
      }

      return {
        items: jobs.map(jobInfoToGQLJob),
        cursor: nextCursor,
        hasNextPage: !!nextCursor,
      };
    },
    jobsByPipeline: async (
      _parent,
      {args: {limit, pipelineIds, projectId}},
      {pachClient},
    ) => {
      const number = limit || DEFAULT_JOBS_LIMIT;
      const pipelineJobs = await Promise.all(
        pipelineIds
          ? pipelineIds.map((pipelineId) =>
              pachClient.pps().listJobs({
                number,
                pipelineId,
                projectId,
              }),
            )
          : [],
      );

      return pipelineJobs.flat().map(jobInfoToGQLJob);
    },
    jobSet: async (_parent, {args: {id, projectId}}, {pachClient}) => {
      return jobInfosToGQLJobSet(
        await getJobsFromJobSet({
          jobSet: await pachClient
            .pps()
            .inspectJobSet({id, projectId, details: false}),
          projectId,
          pachClient,
        }),
        id,
      );
    },
    jobSets: async (
      _parent,
      {args: {projectId, cursor, limit, reverse}},
      {pachClient},
    ) => {
      limit = limit || DEFAULT_JOBS_LIMIT;

      const jobSets = await pachClient.pps().listJobSets({
        projectIds: [projectId],
        details: false,
        cursor: cursor || undefined,
        number: limit + 1,
        reverse: reverse || undefined,
      });
      let nextCursor = undefined;

      //If jobSets.length is not greater than limit there are no pages left
      if (jobSets.length > limit) {
        jobSets.pop(); //remove the extra job from the response
        nextCursor = jobSets[jobSets.length - 1].jobsList[0].started;
      }

      return {
        items: jobSetsToGQLJobSets(jobSets),
        cursor: nextCursor,
        hasNextPage: !!nextCursor,
      };
    },
  },
};

export default pipelineJobResolver;

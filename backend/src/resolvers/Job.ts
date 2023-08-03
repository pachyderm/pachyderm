import {ApolloError} from 'apollo-server-errors';

import {DEFAULT_JOBS_LIMIT} from '@dash-backend/constants/limits';
import {QueryResolvers} from '@dash-backend/generated/types';
import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {nodeStateToJobStateEnum} from '@dash-backend/lib/nodeStateMappers';

import {jqSelect, jqIn, jqCombine} from '../lib/jqHelpers';

import {
  jobInfosToGQLJobSet,
  jobInfoToGQLJob,
  jobsMapToGQLJobSets,
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
      if (!id) {
        const latestJob = (
          await pachClient.pps.listJobs({
            pipelineId: pipelineName,
            projectId,
            number: 1,
          })
        )[0];
        return jobInfoToGQLJob(latestJob);
      }
      return jobInfoToGQLJob(
        await pachClient.pps.inspectJob({id, pipelineName, projectId}),
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
          nodeStateFilter,
          projectId,
          details,
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
        jqFilter = jqCombine(jqFilter, jqSelect(jqIn('.job.id', jobSetIds)));
      }

      if (pipelineIds && pipelineIds.length > 0) {
        jqFilter = jqCombine(
          jqFilter,
          jqSelect(jqIn('.job.pipeline.name', pipelineIds)),
        );
      }

      if (nodeStateFilter && nodeStateFilter.length > 0) {
        const jobsStates = nodeStateFilter
          .map((nodeState) => nodeStateToJobStateEnum(nodeState))
          .flat();
        jqFilter = jqCombine(jqFilter, jqSelect(jqIn('.state', jobsStates)));
      }

      const jobs = await pachClient.pps.listJobs({
        details: details ?? undefined,
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
        nextCursor = jobs[jobs.length - 1].created;
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
              pachClient.pps.listJobs({
                number,
                pipelineId,
                projectId,
              }),
            )
          : [],
      );

      return pipelineJobs.flat().map(jobInfoToGQLJob);
    },
    // TODO: is this also a problem? Who calls this?
    jobSet: async (_parent, {args: {id, projectId}}, {pachClient}) => {
      return jobInfosToGQLJobSet(
        await getJobsFromJobSet({
          jobSet: await pachClient.pps.inspectJobSet({
            id,
            projectId,
            details: false,
          }),
          projectId,
          pachClient,
        }),
        id,
      );
    },
    jobSets: async (_parent, {args: {projectId, limit}}, {pachClient}) => {
      const jobsMap = await pachClient.pps.listJobSetServerDerivedFromListJobs({
        projectId,
        limit: limit || DEFAULT_JOBS_LIMIT,
      });

      return {
        items: jobsMapToGQLJobSets(jobsMap),
        cursor: undefined,
        hasNextPage: false,
      };
    },
  },
};

export default pipelineJobResolver;

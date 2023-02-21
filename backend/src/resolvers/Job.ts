import {ApolloError} from 'apollo-server-errors';

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
      {args: {limit, pipelineId, pipelineIds, jobSetIds, projectId}},
      {pachClient},
    ) => {
      let jqFilter = '';

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
        limit,
        pipelineId,
        jqFilter,
        projectId,
      });

      return jobs.map(jobInfoToGQLJob);
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
    jobSets: async (_parent, {args: {projectId}}, {pachClient}) => {
      return jobSetsToGQLJobSets(
        await pachClient
          .pps()
          .listJobSets({projectIds: [projectId], details: false}),
      );
    },
  },
};

export default pipelineJobResolver;

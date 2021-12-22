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
    jobs: async (_parent, {args: {limit, pipelineId}}, {pachClient}) => {
      const jobs = await pachClient.pps().listJobs({
        limit,
        pipelineId,
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
    jobSets: async (_parent, _args, {pachClient}) => {
      return jobSetsToGQLJobSets(
        await pachClient.pps().listJobSets({details: false}),
      );
    },
  },
};

export default pipelineJobResolver;

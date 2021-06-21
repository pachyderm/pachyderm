import {QueryResolvers} from '@dash-backend/generated/types';

import {jobInfosToGQLJobset, jobInfoToGQLJob} from './builders/pps';

interface PipelineJobResolver {
  Query: {
    job: QueryResolvers['job'];
    jobs: QueryResolvers['jobs'];
    jobset: QueryResolvers['jobset'];
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
    jobs: async (_parent, {args: {pipelineId, limit}}, {pachClient}) => {
      let jq = '';

      if (pipelineId) {
        jq = `select(.job.pipeline.name == "${pipelineId}")`;
      }

      const jobs = await pachClient
        .pps()
        .listJobs({limit: limit || undefined, jq});

      return jobs.map(jobInfoToGQLJob);
    },
    jobset: async (_parent, {args: {id, projectId}}, {pachClient}) => {
      return jobInfosToGQLJobset(
        await pachClient.pps().inspectJobset({id, projectId}),
      );
    },
  },
};

export default pipelineJobResolver;

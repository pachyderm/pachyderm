import {QueryResolvers} from '@dash-backend/generated/types';

import {jobInfoToGQLJob} from './builders/pps';

interface JobResolver {
  Query: {
    jobs: QueryResolvers['jobs'];
  };
}

const jobResolver: JobResolver = {
  Query: {
    jobs: async (_parent, {args: {pipelineId, limit}}, {pachClient}) => {
      let jq = '';

      if (pipelineId) {
        jq = `select(.pipeline.name == "${pipelineId}")`;
      }

      const jobs = await pachClient
        .pps()
        .listJobs({limit: limit || undefined, jq});

      return jobs.map(jobInfoToGQLJob);
    },
  },
};

export default jobResolver;

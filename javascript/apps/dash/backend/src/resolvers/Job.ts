import {QueryResolvers} from '@dash-backend/generated/types';
import client from '@dash-backend/grpc/client';

import {jobInfoToGQLJob} from './builders/pps';

interface JobResolver {
  Query: {
    jobs: QueryResolvers['jobs'];
  };
}

const jobResolver: JobResolver = {
  Query: {
    jobs: async (
      _parent,
      {args: {projectId, pipelineId, limit}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      let jq = '';

      if (pipelineId) {
        jq = `select(.pipeline.name == "${pipelineId}")`;
      }

      const jobs = await client({pachdAddress, authToken, projectId, log})
        .pps()
        .listJobs({limit: limit || undefined, jq});

      return jobs.map(jobInfoToGQLJob);
    },
  },
};

export default jobResolver;

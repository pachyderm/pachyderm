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
      {args: {projectId, limit}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      const jobs = await client({pachdAddress, authToken, projectId, log})
        .pps()
        .listJobs(limit || undefined);

      return jobs.map(jobInfoToGQLJob);
    },
  },
};

export default jobResolver;

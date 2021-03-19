import {QueryResolvers} from '@dash-backend/generated/types';
import client from '@dash-backend/grpc/client';

interface JobResolver {
  Query: {
    jobs: QueryResolvers['jobs'];
  };
}

const jobResolver: JobResolver = {
  Query: {
    jobs: async (
      _parent,
      {args: {projectId}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      const jobs = await client({pachdAddress, authToken, projectId, log})
        .pps()
        .listJobs();

      return jobs.map((job) => ({
        id: job.job?.id || '',
        state: job.state,
        createdAt: job.started?.seconds || 0,
      }));
    },
  },
};

export default jobResolver;

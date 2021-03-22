import {useQuery} from '@apollo/client';

import {GET_JOBS_QUERY} from '@dash-frontend/queries/GetJobsQuery';
import {Job, QueryJobsArgs} from '@graphqlTypes';

type JobsQueryResponse = {
  jobs: Job[];
};

export const useJobs = (projectId = '') => {
  const {data, error, loading} = useQuery<JobsQueryResponse, QueryJobsArgs>(
    GET_JOBS_QUERY,
    {variables: {args: {projectId}}},
  );

  return {
    error,
    jobs: data?.jobs || [],
    loading,
  };
};

import {useQuery} from '@tanstack/react-query';

import {Project} from '@dash-frontend/api/pfs';
import {Job, Pipeline, inspectJob} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export type JobArgs = {
  id?: Job['id'];
  pipelineName: Pipeline['name'];
  projectId: Project['name'];
};

export const useJob = (
  {id, pipelineName, projectId}: JobArgs,
  enabled = true,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.job({
      projectId: projectId,
      pipelineId: pipelineName,
      jobId: id,
    }),
    queryFn: () =>
      inspectJob({
        job: {
          id,
          pipeline: {
            name: pipelineName,
            project: {
              name: projectId,
            },
          },
        },
        details: true,
      }),
    enabled,
  });

  return {job: data, loading, error: getErrorMessage(error)};
};

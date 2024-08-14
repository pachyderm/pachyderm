import {useMutation, useQueryClient} from '@tanstack/react-query';

import {stopJob, StopJobRequest} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

import {useJobsArgs} from './useJobs';

export const useStopJob = () => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['stopJob'],
    mutationFn: (req: StopJobRequest) => {
      return stopJob(req);
    },
    onSuccess: (_data, variables) => {
      const pipelineId = variables.job?.pipeline?.name;
      const projectId = variables.job?.pipeline?.project?.name;
      client.invalidateQueries({
        queryKey: queryKeys.pipeline({
          projectId,
          pipelineId,
        }),
        exact: true,
      });
      client.invalidateQueries({
        queryKey: queryKeys.jobs<useJobsArgs>({
          projectId,
          args: {
            limit: 1,
            pipelineIds: [pipelineId || ''],
            projectName: projectId,
          },
        }),
      });
    },
  });

  return {
    stopJob: mutate,
    error: getErrorMessage(error),
    loading,
  };
};

export default useStopJob;

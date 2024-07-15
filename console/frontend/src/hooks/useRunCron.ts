import {useMutation, useQueryClient} from '@tanstack/react-query';

import {runCron, RunCronRequest} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

import {useJobsArgs} from './useJobs';

export const useRunCron = (onSettled?: () => void) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['runCron'],
    mutationFn: (req: RunCronRequest) => {
      return runCron(req);
    },
    onSuccess: (_data, variables) => {
      const pipelineId = variables?.pipeline?.name;
      const projectId = variables.pipeline?.project?.name;
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
    onSettled,
  });

  return {
    runCron: mutate,
    error: getErrorMessage(error),
    loading,
  };
};

export default useRunCron;

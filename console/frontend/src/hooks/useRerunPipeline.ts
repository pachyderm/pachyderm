import {useMutation, useQueryClient} from '@tanstack/react-query';

import {rerunPipeline, RerunPipelineRequest} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

import {useJobsArgs} from './useJobs';

export const useRerunPipeline = (onSettled?: () => void) => {
  const client = useQueryClient();
  const {
    mutate,
    error,
    isPending: loading,
  } = useMutation({
    mutationKey: ['rerunPipeline'],
    mutationFn: (req: RerunPipelineRequest) => {
      return rerunPipeline(req);
    },
    onSuccess: (_data, variables) => {
      const projectId = variables.pipeline?.project?.name;
      const pipelineId = variables.pipeline?.name;
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
    rerunPipeline: mutate,
    error: getErrorMessage(error),
    loading,
  };
};

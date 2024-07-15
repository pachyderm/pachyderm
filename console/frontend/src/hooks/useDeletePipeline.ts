import {useMutation, useQueryClient} from '@tanstack/react-query';

import {deletePipeline, DeletePipelineRequest} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type useDeletePipelineArgs = {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
};

const useDeletePipeline = ({onSuccess, onError}: useDeletePipelineArgs) => {
  const client = useQueryClient();
  const {
    mutate: deletePipelineMutation,
    isPending: loading,
    error,
    data,
  } = useMutation({
    mutationKey: ['deletePipeline'],
    mutationFn: (req: DeletePipelineRequest) => deletePipeline(req),
    onSuccess: async (_data, variables) => {
      client.invalidateQueries({
        queryKey: queryKeys.pipelines({
          projectId: variables.pipeline?.project?.name,
        }),
        exact: true,
      });
      client.invalidateQueries({
        queryKey: queryKeys.repos({
          projectId: variables.pipeline?.project?.name,
        }),
        exact: true,
      });
      client.invalidateQueries({
        queryKey: queryKeys.allPipelines,
      });
      onSuccess && onSuccess();
    },
    onError: (error) => onError && onError(error),
  });

  return {
    deletePipelineResponse: data,
    deletePipeline: deletePipelineMutation,
    error: getErrorMessage(error),
    loading,
  };
};

export default useDeletePipeline;

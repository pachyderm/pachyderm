import {useMutation, useQueryClient} from '@tanstack/react-query';

import {createPipeline, CreatePipelineV2Request} from '@dash-frontend/api/pps';
import queryKeys from '@dash-frontend/lib/queryKeys';

type useCreatePipelineArgs = {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
};

const useCreatePipeline = ({onSuccess, onError}: useCreatePipelineArgs) => {
  const client = useQueryClient();
  const {
    mutate: createPipelineMutation,
    isPending: loading,
    error,
    data,
  } = useMutation({
    mutationKey: ['createPipeline'],
    mutationFn: (req: CreatePipelineV2Request) => createPipeline(req),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: queryKeys.allPipelines,
      });
      onSuccess && onSuccess();
    },
    onError: (error) => onError && onError(error),
  });

  return {
    createPipelineResponse: data,
    createPipeline: createPipelineMutation,
    error,
    loading,
  };
};

export default useCreatePipeline;

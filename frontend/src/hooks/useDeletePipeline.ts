import {useApolloClient} from '@apollo/client';

import {useDeletePipelineMutation} from '@dash-frontend/generated/hooks';

export const useDeletePipeline = (onCompleted?: () => void) => {
  const client = useApolloClient();
  const [deletePipeline, {loading, error}] = useDeletePipelineMutation({
    onCompleted: () => {
      onCompleted && onCompleted();
      client.cache.reset();
    },
  });

  return {
    deletePipeline,
    error,
    loading,
  };
};

import {useApolloClient} from '@apollo/client';

import {useRerunPipelineMutation} from '@dash-frontend/generated/hooks';

export const useRerunPipeline = (onCompleted?: () => void) => {
  const client = useApolloClient();
  const [rerunPipeline, {loading, error}] = useRerunPipelineMutation({
    onCompleted: () => {
      onCompleted && onCompleted();
      client.cache.reset();
    },
  });

  return {
    rerunPipeline,
    error,
    loading,
  };
};

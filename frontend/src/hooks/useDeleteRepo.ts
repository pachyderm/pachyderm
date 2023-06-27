import {useApolloClient} from '@apollo/client';

import {useDeleteRepoMutation} from '@dash-frontend/generated/hooks';

export const useDeleteRepo = (onCompleted?: () => void) => {
  const client = useApolloClient();
  const [deleteRepo, {loading, error}] = useDeleteRepoMutation({
    onCompleted: () => {
      onCompleted && onCompleted();
      client.resetStore();
    },
  });

  return {
    deleteRepo,
    error,
    loading,
  };
};

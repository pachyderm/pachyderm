import {CreateRepoArgs} from '@graphqlTypes';

import {useCreateRepoMutation} from '@dash-frontend/generated/hooks';

const useCreateRepo = (onCompleted?: () => void) => {
  const [createRepoMutation, {loading, error}] = useCreateRepoMutation({
    onCompleted,
  });

  return {
    createRepo: (args: CreateRepoArgs) =>
      createRepoMutation({variables: {args}}),
    loading,
    error,
  };
};

export default useCreateRepo;

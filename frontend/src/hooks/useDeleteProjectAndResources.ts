import {useApolloClient} from '@apollo/client';
import {DeleteProjectAndResourcesArgs} from '@graphqlTypes';

import {useDeleteProjectAndResourcesMutation} from '@dash-frontend/generated/hooks';

export const useDeleteProjectAndResources = (onCompleted?: () => void) => {
  const client = useApolloClient();
  const [deleteProjectMutation, mutationResult] =
    useDeleteProjectAndResourcesMutation({
      onCompleted: () => {
        onCompleted && onCompleted();
        client.cache.reset();
      },
    });

  return {
    deleteProject: (args: DeleteProjectAndResourcesArgs) =>
      deleteProjectMutation({variables: {args}}),
    ...mutationResult,
  };
};

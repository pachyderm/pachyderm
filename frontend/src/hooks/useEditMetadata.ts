import {useMutation, useQueryClient} from '@tanstack/react-query';

import {editMetadata, Edit} from '@dash-frontend/api/metadata';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type useEditMetadataArgs = {
  onSuccess?: () => void;
};

export const useEditMetadata = ({onSuccess}: useEditMetadataArgs) => {
  const client = useQueryClient();
  const {
    mutate,
    isPending: loading,
    error,
  } = useMutation({
    mutationKey: ['editRepoMetadata'],
    mutationFn: (req: Edit) => editMetadata({edits: [req]}),
    onSuccess: (_data, variables) => {
      const projectId =
        variables.repo?.name?.project?.name ||
        variables.commit?.id?.repo?.name?.project?.name;
      const repoId =
        variables.repo?.name?.name || variables.commit?.id?.repo?.name?.name;

      client.invalidateQueries({
        queryKey: queryKeys.repo({
          projectId,
          repoId,
        }),
      });
      client.invalidateQueries({
        queryKey: queryKeys.projects,
      });
      client.invalidateQueries({
        queryKey: queryKeys.inspectCluster,
      });
      onSuccess && onSuccess();
    },
  });

  return {
    updateMetadata: mutate,
    loading,
    error: getErrorMessage(error),
  };
};

import {useQuery} from '@tanstack/react-query';
import uniqBy from 'lodash/uniqBy';

import {globFile} from '@dash-frontend/api/pfs';
import {isUnknown} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseSearchFilesProps = {
  commitId: string;
  pattern: string;
  projectId: string;
  repoId: string;
};

const MAX_FILES = 1000;

export const useSearchFiles = ({
  commitId,
  pattern,
  projectId,
  repoId,
}: UseSearchFilesProps) => {
  const {
    data,
    refetch,
    isLoading: loading,
    error,
    isFetching,
  } = useQuery({
    queryKey: [queryKeys.filesSearch({commitId, projectId, repoId, pattern})],
    enabled: false,
    throwOnError: (e) => !isUnknown(e),
    queryFn: async () => {
      if (!pattern) return [];

      // TODO: These queries are ideally moved to the Meta API when ready
      const top = await globFile(
        {
          commit: {
            id: commitId,
            repo: {
              name: repoId,
              project: {
                name: projectId,
              },
              type: 'user',
            },
          },
          pattern: `*${pattern}*`,
        },
        MAX_FILES,
      );
      const recursive = await globFile(
        {
          commit: {
            id: commitId,
            repo: {
              name: repoId,
              project: {
                name: projectId,
              },
              type: 'user',
            },
          },
          pattern: `**/*${pattern}*`,
        },
        MAX_FILES,
      );

      return uniqBy([...top, ...recursive], (file) => file?.file?.path);
    },
  });

  return {
    files: data,
    searchFiles: refetch,
    loading,
    isFetching,
    error: getErrorMessage(error),
  };
};

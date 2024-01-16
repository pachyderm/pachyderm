import {useQuery} from '@tanstack/react-query';

import {listFilesPaged} from '@dash-frontend/api/pfs';
import {isUnknown} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseFilesArgs = {
  number?: number;
  reverse?: boolean;
  cursorPath?: string;
};

type UseFiles = {
  projectName: string;
  repoName: string;
  commitId?: string;
  path?: string;
  args: UseFilesArgs;
};

export const useFiles = (
  {projectName, repoName, commitId, path, args}: UseFiles,
  enabled = true,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.files<UseFilesArgs>({
      projectId: projectName,
      repoId: repoName,
      commitId,
      path,
      args,
    }),
    enabled: enabled,
    throwOnError: (e) => !isUnknown(e),
    queryFn: () => {
      return listFilesPaged({
        file: {
          commit: {
            id: commitId,
            repo: {
              name: repoName,
              type: 'user',
              project: {
                name: projectName,
              },
            },
          },
          path: path,
        },
        paginationMarker: {
          commit: {
            id: commitId,
            repo: {
              name: repoName,
              type: 'user',
              project: {
                name: projectName,
              },
            },
          },
          path: args.cursorPath,
        },
        reverse: args.reverse,
        number: args.number ? String(args.number) : undefined,
      });
    },
  });

  return {
    ...data,
    loading,
    error: getErrorMessage(error),
  };
};

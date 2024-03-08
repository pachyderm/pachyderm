import {useQuery} from '@tanstack/react-query';

import {listFiles, FileInfo} from '@dash-frontend/api/pfs';
import {isUnknown} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import getParentPath from '@dash-frontend/lib/getParentPath';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseFilesNextPrevious = {
  file?: FileInfo;
};

export const useFilesNextPrevious = (
  {file}: UseFilesNextPrevious = {},
  enabled = true,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.filesNextPrevious({
      projectId: file?.file?.commit?.repo?.project?.name,
      repoId: file?.file?.commit?.repo?.name,
      branchName: file?.file?.commit?.branch?.name,
      commitId: file?.file?.commit?.id,
      path: file?.file?.path,
    }),
    enabled: enabled,
    throwOnError: (e) => !isUnknown(e),
    queryFn: async () => {
      // The root directory will not have a "next" or "previous" file.
      // All other files and directories will potentially have "next" and "previous" files.
      const isNotRootDir =
        file?.file?.path !== '/' && file?.file?.path !== '//';
      const parentPath = getParentPath(file?.file?.path);
      const next = isNotRootDir
        ? (
            await listFiles({
              file: {...file?.file, path: parentPath},
              paginationMarker: file?.file,
              number: '1',
            })
          )[0]
        : undefined;
      const previous = isNotRootDir
        ? (
            await listFiles({
              file: {...file?.file, path: parentPath},
              paginationMarker: file?.file,
              number: '1',
              reverse: true,
            })
          )[0]
        : undefined;

      return {next, previous};
    },
  });

  return {
    ...data,
    loading,
    error: getErrorMessage(error),
  };
};

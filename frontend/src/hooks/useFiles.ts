import {FileQueryArgs} from '@graphqlTypes';

import {useGetFilesQuery} from '@dash-frontend/generated/hooks';

type UseFilesArgs = {
  args: FileQueryArgs;
  skip?: boolean;
};

export const useFiles = ({
  args: {
    projectId,
    commitId = 'master',
    path = '/',
    repoName,
    branchName,
    limit,
    cursorPath,
    reverse,
  },
  skip,
}: UseFilesArgs) => {
  // TODO: This might be better as a lazy query with options
  const {data, error, loading} = useGetFilesQuery({
    variables: {
      args: {
        projectId,
        commitId,
        path,
        repoName,
        branchName,
        limit,
        cursorPath,
        reverse,
      },
    },
    skip,
  });

  return {
    error,
    files: data?.files,
    loading,
    hasNextPage: data?.files.hasNextPage,
    cursor: data?.files.cursor,
  };
};

export default useFiles;

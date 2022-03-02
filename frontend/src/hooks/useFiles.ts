import {FileQueryArgs} from '@graphqlTypes';

import {useGetFilesQuery} from '@dash-frontend/generated/hooks';

export const useFiles = ({
  projectId,
  commitId = 'master',
  path = '/',
  repoName,
  branchName,
}: FileQueryArgs) => {
  // TODO: This might be better as a lazy query with options
  const {data, error, loading} = useGetFilesQuery({
    variables: {args: {projectId, commitId, path, repoName, branchName}},
  });

  return {
    error,
    files: data?.files,
    loading,
  };
};

export default useFiles;

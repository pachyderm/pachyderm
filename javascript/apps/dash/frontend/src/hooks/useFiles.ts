import {useGetFilesQuery} from '@dash-frontend/generated/hooks';
import {FileQueryArgs} from '@graphqlTypes';

export const useFiles = ({
  commitId = 'master',
  path = '/',
  repoName,
  branchName,
}: FileQueryArgs) => {
  // TODO: This might be better as a lazy query with options
  const {data, error, loading} = useGetFilesQuery({
    variables: {args: {commitId, path, repoName, branchName}},
  });

  return {
    error,
    files: data?.files || [],
    loading,
  };
};

export default useFiles;

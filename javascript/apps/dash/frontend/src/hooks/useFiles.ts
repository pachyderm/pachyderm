import {useGetFilesQuery} from '@dash-frontend/generated/hooks';

type UseFilesProps = {
  commitId?: string;
  path?: string;
  repoName: string;
};

export const useFiles = ({
  commitId = 'master',
  path = '/',
  repoName,
}: UseFilesProps) => {
  // TODO: This might be better as a lazy query with options
  const {data, error, loading} = useGetFilesQuery({
    variables: {args: {commitId, path, repoName}},
  });

  return {
    error,
    files: data?.files || [],
    loading,
  };
};

export default useFiles;

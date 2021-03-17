import {useQuery} from '@apollo/client';

import {GET_FILES_QUERY} from '@dash-frontend/queries/GetFilesQuery';
import {File} from '@graphqlTypes';

type FilesQueryResponse = {
  files: File[];
};

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
  const {data, error, loading} = useQuery<FilesQueryResponse>(GET_FILES_QUERY, {
    variables: {args: {commitId, path, repoName}},
  });

  return {
    error,
    files: data?.files || [],
    loading,
  };
};

export default useFiles;

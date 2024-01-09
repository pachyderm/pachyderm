import {useQuery} from '@tanstack/react-query';

import {diffFile, DiffFileRequest} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useCommitDiff = (req: DiffFileRequest, enabled = true) => {
  const commit = req.newFile?.commit;
  const {project, name: repoName} = req.newFile?.commit?.repo || {};
  const {
    data: fileDiff,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.commitDiff<DiffFileRequest>({
      projectId: project?.name,
      repoId: repoName,
      commitId: commit?.id,
      args: req,
    }),
    queryFn: () => diffFile(req),
    enabled,
    refetchInterval: false,
  });

  return {
    fileDiff,
    error: getErrorMessage(error),
    loading,
  };
};

import {CommitInfo} from '@dash-frontend/api/pfs';
import {useCommitDiff} from '@dash-frontend/hooks/useCommitDiff';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {formatDiffOnlyTotals} from '@dash-frontend/lib/formatDiff';

const useFileHistory = (commit: CommitInfo) => {
  const {repoId, projectId, filePath} = useUrlState();
  const {fileDiff, loading, error} = useCommitDiff({
    newFile: {
      commit: {
        id: commit.commit?.id,
        repo: {
          name: repoId,
          project: {
            name: projectId,
          },
          type: 'user',
        },
        branch: {
          name: commit?.commit?.branch?.name,
        },
      },
      path: filePath || '/',
    },
  });
  const formattedDiff = fileDiff && formatDiffOnlyTotals(fileDiff);

  return {
    commitAction: formattedDiff && Object.values(formattedDiff)[0],
    loading,
    error,
  };
};

export default useFileHistory;

import {useQuery} from '@tanstack/react-query';

import {CommitState, inspectCommit} from '@dash-frontend/api/pfs';
import {isNotFound} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

type UseCommitSearch = {
  projectId: string;
  repoId: string;
  commitId: string;
};

export const useCommitSearch = (args: UseCommitSearch, skip = false) => {
  const {projectId, repoId, commitId} = args;
  const {
    data,
    error,
    isLoading: loading,
  } = useQuery({
    queryKey: queryKeys.commitSearch({projectId, repoId, commitId}),
    enabled: !skip,
    throwOnError: (e) => !isNotFound(e),
    queryFn: () =>
      // We want to specify the repo name but not the branch name
      // as we are searching with only commit id.
      inspectCommit({
        commit: {
          id: commitId,
          branch: {
            name: '',
            repo: {name: repoId, project: {name: projectId}, type: 'user'},
          },
        },
        wait: CommitState.COMMIT_STATE_UNKNOWN,
      }),
  });
  return {
    error: getErrorMessage(error),
    loading,
    commit: data,
  };
};

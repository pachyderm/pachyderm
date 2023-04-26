import {FoundCommit} from '@graphqlTypes';
import {useEffect, useMemo, useState} from 'react';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useFindCommits from '@dash-frontend/hooks/useFindCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDateOnly} from '@dash-frontend/lib/dateTime';

const useFileHistory = () => {
  const {repoId, branchId, projectId, filePath, commitId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const [hasCalledFind, setHasCalledFind] = useState(false);
  const [commitList, setCommitList] = useState<FoundCommit[]>([]);

  const {findCommits, commits, loading, hasNextPage, cursor, error} =
    useFindCommits();

  useEffect(() => {
    if (commits && !loading) {
      setHasCalledFind(true);
      setCommitList((prev) => [...prev, ...commits]);
    }
  }, [commits, loading]);

  const lazyQueryArgs = {
    variables: {
      args: {
        projectId,
        repoId,
        filePath: `/${filePath}`,
        commitId: cursor ? cursor : commitId,
        branchId,
      },
    },
  };

  const dateRange = useMemo(() => {
    return commitList.length !== 0
      ? `${getStandardDateOnly(commitList[0]?.started)} - ${
          commitList &&
          getStandardDateOnly(commitList[commitList.length - 1]?.started)
        }`
      : null;
  }, [commitList]);

  const disableSearch = !hasCalledFind ? false : !hasNextPage;

  return {
    findCommits,
    loading,
    commitList,
    getPathToFileBrowser,
    dateRange,
    disableSearch,
    lazyQueryArgs,
    error,
  };
};

export default useFileHistory;

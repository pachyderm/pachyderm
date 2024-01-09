import {useEffect, useState} from 'react';

import {usePreviousValue} from '@dash-frontend/../components/src';
import {useCommits} from '@dash-frontend/hooks/useCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {COMMIT_PAGE_SIZE} from '@dash-frontend/views/FileBrowser/constants/FileBrowser';

const useLeftPanel = (selectedCommitId?: string) => {
  const [page, setPage] = useState(1);
  const [cursors, setCursors] = useState<string[]>(['']);
  const [currentCursor, setCurrentCursor] = useState('');
  const [refetching, setRefetching] = useState(false);

  const {projectId, repoId, branchId} = useUrlState();
  const previousBranchId = usePreviousValue(branchId);

  useEffect(() => {
    if (previousBranchId !== branchId) {
      setPage(1);
      setCursors(['']);
      setCurrentCursor('');
    }
  }, [branchId, previousBranchId]);

  const {
    commits,
    parentCommit: cursor,
    loading,
    refetch,
  } = useCommits({
    projectName: projectId,
    repoName: repoId,
    args: {
      // We only need to specity branchId when getting the first page
      // this avoids race conditions with pulling branchId from url
      branchName: !currentCursor ? branchId : undefined,
      commitIdCursor: currentCursor,
      number: COMMIT_PAGE_SIZE,
    },
  });

  useEffect(() => {
    if (cursor && cursors.length !== 0 && cursors.length < page) {
      setCurrentCursor(cursor);
      setCursors((arr) => [...arr, cursor]);
    } else {
      setCurrentCursor(cursors[page - 1]);
    }
  }, [commits, currentCursor, cursor, cursors, page]);

  // We want to update the results for the first page whenever we navigate to it.
  // This provides up to date information if new commits came in while we were on
  // different pages and will update the cursor. We are also watching "commitId"
  // because this will refetch the commit list in cases where a user is on the
  // first page and deletes a file or a user navigates to "/latest" for a branch.
  useEffect(() => {
    if (page === 1) {
      const ids = commits?.map((commit) => commit.commit?.id);
      if (!ids?.find((id) => id === selectedCommitId)) {
        const reload = async () => {
          setRefetching(true);
          await refetch();
          setRefetching(false);
          setCursors(['']);
        };
        reload();
      }
    }
  }, [selectedCommitId, page, refetch, commits]);

  const contentLength =
    (cursors.length - 1) * COMMIT_PAGE_SIZE + (commits?.length || 0);

  const hasNextPage = !!cursor;

  return {
    commits,
    loading: loading || refetching,
    page,
    setPage,
    hasNextPage,
    contentLength,
  };
};

export default useLeftPanel;

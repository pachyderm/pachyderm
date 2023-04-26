import {useEffect, useState} from 'react';

import {usePreviousValue} from '@dash-frontend/../components/src';
import useCommits from '@dash-frontend/hooks/useCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {COMMIT_PAGE_SIZE} from '@dash-frontend/views/FileBrowser/constants/FileBrowser';

const useLeftPanel = () => {
  const [page, setPage] = useState(1);
  const [cursors, setCursors] = useState<string[]>(['']);
  const [currentCursor, setCurrentCursor] = useState('');

  const {projectId, repoId, branchId, commitId} = useUrlState();
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
    args: {
      number: COMMIT_PAGE_SIZE,
      projectId: projectId,
      repoName: repoId,
      // We only need to specity branchId when getting the first page
      // this avoids race conditions with pulling branchId from url
      branchName: !currentCursor ? branchId : null,
      commitIdCursor: currentCursor,
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
      refetch({
        args: {
          number: COMMIT_PAGE_SIZE,
          projectId: projectId,
          repoName: repoId,
          branchName: !currentCursor ? branchId : null,
        },
      });
      setCursors(['']);
    }
  }, [branchId, currentCursor, page, projectId, refetch, repoId, commitId]);

  const contentLength =
    (cursors.length - 1) * COMMIT_PAGE_SIZE + commits.length;

  const hasNextPage = !!cursor;

  return {
    commits,
    loading,
    page,
    setPage,
    hasNextPage,
    contentLength,
  };
};

export default useLeftPanel;

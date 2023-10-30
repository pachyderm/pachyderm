import {Commit} from '@graphqlTypes';
import {useCallback, useEffect} from 'react';
import {useHistory} from 'react-router';

import useCommits from '@dash-frontend/hooks/useCommits';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useTimestampPagination from '@dash-frontend/hooks/useTimestampPagination';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

export const COMMITS_DEFAULT_PAGE_SIZE = 15;

const useCommitsList = (selectedRepo: string, reverseOrder: boolean) => {
  const {projectId} = useUrlState();
  const {getUpdatedSearchParams} = useUrlQueryState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const browserHistory = useHistory();
  const {
    page,
    setPage,
    pageSize,
    setPageSize,
    cursors,
    updateCursor,
    resetCursors,
  } = useTimestampPagination(COMMITS_DEFAULT_PAGE_SIZE);

  const {commits, loading, error, cursor} = useCommits({
    args: {
      projectId,
      repoName: selectedRepo,
      number: pageSize,
      reverse: reverseOrder,
      cursor: cursors[page - 1],
    },
  });
  const hasNextPage = !!cursor;

  const updatePage = useCallback(
    (page: number) => {
      updateCursor(cursor);
      setPage(page);
    },
    [cursor, setPage, updateCursor],
  );

  useEffect(() => {
    setPage(1);
    resetCursors();
  }, [resetCursors, reverseOrder, setPage]);

  const globalIdRedirect = (runId: string) => {
    const searchParams = getUpdatedSearchParams(
      {
        globalIdFilter: runId,
      },
      true,
    );

    return browserHistory.push(
      `${lineageRoute(
        {
          projectId,
        },
        false,
      )}?${searchParams}`,
    );
  };

  const inspectCommitRedirect = (commit: Commit) => {
    const fileBrowserLink = getPathToFileBrowser(
      {
        projectId,
        repoId: commit.repoName,
        commitId: commit.id,
        branchId: commit.branch?.name || 'default',
      },
      true,
    );
    browserHistory.push(fileBrowserLink);
  };

  const onOverflowMenuSelect = (commit: Commit) => (id: string) => {
    switch (id) {
      case 'apply-run':
        return globalIdRedirect(commit.id);
      case 'inspect-commit':
        return inspectCommitRedirect(commit);
      default:
        return null;
    }
  };

  const iconItems: DropdownItem[] = [
    {
      id: 'inspect-commit',
      content: 'Inspect commit',
      closeOnClick: true,
    },
    {
      id: 'apply-run',
      content: 'Apply Global ID and view in DAG',
      closeOnClick: true,
    },
  ];

  return {
    commits,
    loading,
    error,
    iconItems,
    onOverflowMenuSelect,
    page,
    cursors,
    updatePage,
    pageSize,
    setPageSize,
    hasNextPage,
  };
};

export default useCommitsList;

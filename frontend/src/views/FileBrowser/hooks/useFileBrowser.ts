import {useMemo, useCallback, useState, useEffect} from 'react';
import {useHistory} from 'react-router';

import useCommits from '@dash-frontend/hooks/useCommits';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import {useFiles} from '@dash-frontend/hooks/useFiles';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  repoRoute,
  fileBrowserRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {useModal, usePreviousValue} from '@pachyderm/components';

import {FILE_DEFAULT_PAGE_SIZE} from '../constants/FileBrowser';

const useFileBrowser = () => {
  const browserHistory = useHistory();
  const {repoId, commitId, branchId, filePath, projectId} = useUrlState();
  const {closeModal, isOpen} = useModal(true);
  const {getPathFromFileBrowser} = useFileBrowserNavigation();
  const previousCommitId = usePreviousValue(commitId);
  const previousFilePath = usePreviousValue(filePath);

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(FILE_DEFAULT_PAGE_SIZE);
  const [cursors, setCursors] = useState<(string | null)[]>([null]);

  const path = `/${filePath}`;
  const isDirectory = path.endsWith('/');
  const isRoot = path === '/';

  const {
    commits,
    loading: commitsLoading,
    error: commitsError,
  } = useCommits({
    args: {
      projectId: projectId,
      repoName: repoId,
      branchName: branchId,
      commitIdCursor: commitId || '',
      number: 1,
    },
    fetchPolicy: 'no-cache',
  });

  useEffect(() => {
    if (previousCommitId !== commitId || previousFilePath !== filePath) {
      setPage(1);
      setCursors(['']);
    }
  }, [branchId, commitId, filePath, previousCommitId, previousFilePath]);

  const selectedCommitId = !commitId ? commits[0]?.id : commitId;
  const isCommitOpen = commits[0] ? commits[0].finished === -1 : true;

  const {
    files,
    loading: filesLoading,
    error: filesError,
    cursor,
    hasNextPage,
  } = useFiles({
    args: {
      projectId,
      commitId: commitId ? commitId : '',
      path: path || '/',
      branchName: branchId,
      repoName: repoId,
      limit: isDirectory ? pageSize : undefined,
      cursorPath: cursors[page - 1],
    },
    skip: isCommitOpen,
  });

  const updatePage = useCallback(
    (page: number) => {
      if (cursor && page + 1 > cursors.length) {
        setCursors((arr) => [...arr, cursor]);
      }
      setPage(page);
    },
    [cursor, cursors.length],
  );

  const fileToPreview = useMemo(() => {
    const hasFileType = path.slice(path.lastIndexOf('.') + 1) !== '';

    return hasFileType && files?.files.find((file) => file.path === path);
  }, [path, files]);

  const handleHide = useCallback(() => {
    closeModal();

    setTimeout(
      () =>
        browserHistory.push(
          getPathFromFileBrowser(repoRoute({projectId, repoId}, false)),
        ),
      500,
    );
  }, [projectId, repoId, browserHistory, closeModal, getPathFromFileBrowser]);

  const handleBackNav = () => {
    const filePaths = filePath.split('/').slice(0, -2);
    const parentPath = filePaths.join('/').concat('/');
    browserHistory.push(
      fileBrowserRoute({
        repoId,
        branchId,
        projectId,
        commitId,
        filePath: parentPath === '/' ? undefined : parentPath,
      }),
    );
  };

  return {
    handleHide,
    handleBackNav,
    isOpen,
    files: files?.files || [],
    loading: commitsLoading || filesLoading,
    error: commitsError || filesError,
    fileToPreview,
    isDirectory,
    isRoot,
    selectedCommitId,
    pageSize,
    setPageSize,
    page,
    updatePage,
    cursors,
    hasNextPage,
    isCommitOpen,
  };
};

export default useFileBrowser;

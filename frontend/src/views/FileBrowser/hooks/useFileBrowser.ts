import {useMemo, useCallback, useState, useEffect} from 'react';
import {useHistory} from 'react-router';

import {useCommits} from '@dash-frontend/hooks/useCommits';
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
  // Users can enter the file browser with or without a commitId (/latest). This
  // call is to retrieve the latest commit when users are on the /latest path.
  const {
    commits,
    loading: commitsLoading,
    error: commitsError,
  } = useCommits({
    projectName: projectId,
    repoName: repoId,
    args: {
      branchName: branchId,
      commitIdCursor: commitId || '',
      number: 1,
    },
  });

  useEffect(() => {
    if (previousCommitId !== commitId || previousFilePath !== filePath) {
      setPage(1);
      setCursors(['']);
    }
  }, [branchId, commitId, filePath, previousCommitId, previousFilePath]);

  const selectedCommitId = commitId || commits?.[0]?.commit?.id;
  const isCommitOpen = !commits?.[0]?.finished;

  const {
    files,
    loading: filesLoading,
    error: filesError,
    cursor,
  } = useFiles(
    {
      projectName: projectId,
      commitId: commitId ? commitId : '',
      path: path || '/',
      branchName: branchId,
      repoName: repoId,
      args: {
        cursorPath: cursors[page - 1] || undefined,
        number: isDirectory ? pageSize : undefined,
      },
    },
    !isCommitOpen,
  );

  const updatePage = useCallback(
    (page: number) => {
      if (cursor?.file?.path && page + 1 > cursors.length) {
        setCursors((arr) => [...arr, cursor?.file?.path || '']);
      }
      setPage(page);
    },
    [cursor, cursors.length],
  );

  const fileToPreview = useMemo(() => {
    const hasFileType = path.slice(path.lastIndexOf('.') + 1) !== '';

    return (
      hasFileType &&
      files?.find((file) => {
        if (!file.file?.path) {
          return false;
        }
        return file.file.path === path;
      })
    );
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
    files: files || [],
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
    hasNextPage: !!cursor,
    isCommitOpen,
  };
};

export default useFileBrowser;

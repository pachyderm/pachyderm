import {useCallback, useEffect} from 'react';
import {useHistory} from 'react-router';

import {FileInfo} from '@dash-frontend/api/pfs';
import {useFilesNextPrevious} from '@dash-frontend/hooks/useFilesNextPrevious';
import getParentPath from '@dash-frontend/lib/getParentPath';
import {filePreviewRoute} from '@dash-frontend/views/Project/utils/routes';

import {NAVIGATION_KEYS} from '../constants/FileBrowser';

type UseFilePreviewNavigationProps = {
  file: FileInfo;
};

const useFilePreviewNavigation = ({file}: UseFilePreviewNavigationProps) => {
  const browserHistory = useHistory();
  const {next, previous, loading} = useFilesNextPrevious({file});
  const handleNavigateNext = useCallback(() => {
    if (!next?.file) return;

    browserHistory.push(
      filePreviewRoute({
        repoId: next.file.commit?.repo?.name || '',
        projectId: next.file.commit?.repo?.project?.name || '',
        commitId: next.file.commit?.id || '',
        filePath: (next.file.path || '').replace(/^\//, '') || undefined,
      }),
    );
  }, [browserHistory, next]);
  const handleNavigatePrevious = useCallback(() => {
    if (!previous?.file) return;

    browserHistory.push(
      filePreviewRoute({
        repoId: previous.file.commit?.repo?.name || '',
        projectId: previous.file.commit?.repo?.project?.name || '',
        commitId: previous.file.commit?.id || '',
        filePath: (previous.file.path || '').replace(/^\//, '') || undefined,
      }),
    );
  }, [browserHistory, previous]);
  const handleNavigateParent = useCallback(() => {
    if (!file?.file) return;

    browserHistory.push(
      filePreviewRoute({
        repoId: file.file.commit?.repo?.name || '',
        projectId: file.file.commit?.repo?.project?.name || '',
        commitId: file.file.commit?.id || '',
        filePath: getParentPath(file.file.path).replace(/^\//, '') || undefined,
      }),
    );
  }, [browserHistory, file]);

  useEffect(() => {
    const handleKeyDown = (evt: KeyboardEvent) => {
      // TODO: Update these when hotkeys are finalized as they might need modifiers
      if (evt.key === NAVIGATION_KEYS.PREVIOUS) {
        handleNavigatePrevious();
      } else if (evt.key === NAVIGATION_KEYS.NEXT) {
        handleNavigateNext();
      } else if (evt.key === NAVIGATION_KEYS.PARENT) {
        handleNavigateParent();
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleNavigateNext, handleNavigatePrevious, handleNavigateParent]);

  return {
    handleNavigateNext,
    handleNavigateParent,
    handleNavigatePrevious,
    loading,
  };
};

export default useFilePreviewNavigation;

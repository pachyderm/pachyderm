import {useEffect, useRef} from 'react';
import {useHistory} from 'react-router';

import getParentPath from '@dash-frontend/lib/getParentPath';
import parse from '@dash-frontend/lib/parsePath';
import removeStartingSlash from '@dash-frontend/lib/removeStartingSlash';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import {NAVIGATION_KEYS} from '../constants';
import {FileTreeNodeProps} from '../types';

const useFileTreeFile = ({
  filePath = '',
  repoId,
  projectId,
  selectedCommitId,
  currentPath,
  previousPath,
  nextPath,
}: FileTreeNodeProps) => {
  const browserHistory = useHistory();
  const nameRef = useRef<HTMLDivElement>(null);
  const {base} = parse(filePath);
  const link = fileBrowserRoute({
    repoId,
    projectId,
    commitId: selectedCommitId,
    filePath: removeStartingSlash(filePath),
  });
  const isActive = currentPath === filePath;
  const parentPath = getParentPath(currentPath);

  useEffect(() => {
    const handleKeyDown = (evt: KeyboardEvent) => {
      if (evt.key === NAVIGATION_KEYS.PREVIOUS && previousPath) {
        evt.preventDefault();
        browserHistory.push(
          fileBrowserRoute({
            repoId,
            projectId,
            commitId: selectedCommitId,
            filePath: removeStartingSlash(previousPath),
          }),
        );
      } else if (evt.key === NAVIGATION_KEYS.NEXT && nextPath) {
        evt.preventDefault();
        browserHistory.push(
          fileBrowserRoute({
            repoId,
            projectId,
            commitId: selectedCommitId,
            filePath: removeStartingSlash(nextPath),
          }),
        );
      } else if (evt.key === NAVIGATION_KEYS.CLOSE) {
        browserHistory.push(
          fileBrowserRoute({
            repoId,
            projectId,
            commitId: selectedCommitId,
            filePath: removeStartingSlash(parentPath),
          }),
        );
      }
    };

    if (isActive) {
      document.addEventListener('keydown', handleKeyDown);
    } else {
      document.removeEventListener('keydown', handleKeyDown);
    }

    if (isActive && nameRef.current) {
      nameRef.current.scrollIntoView({block: 'center'});
    }

    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [
    isActive,
    nameRef,
    browserHistory,
    selectedCommitId,
    repoId,
    projectId,
    previousPath,
    nextPath,
    parentPath,
  ]);

  return {isActive, link, name: base, nameRef};
};

export default useFileTreeFile;

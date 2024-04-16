import {useEffect, useState, useRef} from 'react';
import {useHistory} from 'react-router';

import {useFiles} from '@dash-frontend/hooks/useFiles';
import getParentPath from '@dash-frontend/lib/getParentPath';
import parse from '@dash-frontend/lib/parsePath';
import removeStartingSlash from '@dash-frontend/lib/removeStartingSlash';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import {MAX_FILES, NAVIGATION_KEYS} from '../constants';
import {FileTreeNodeProps} from '../types';

const useFileTreeDirectory = ({
  filePath = '',
  repoId,
  projectId,
  selectedCommitId,
  currentPath,
  previousPath,
  nextPath,
  open,
}: FileTreeNodeProps) => {
  const browserHistory = useHistory();
  const nameRef = useRef<HTMLDivElement>(null);
  const [isOpen, setIsOpen] = useState(
    open || currentPath.startsWith(filePath),
  );
  const {base} = parse(filePath);
  const link = fileBrowserRoute({
    repoId,
    projectId,
    commitId: selectedCommitId,
    filePath,
  });
  const isActive = currentPath === filePath;
  const isDir = currentPath.endsWith('/');
  const parentPath = getParentPath(currentPath);
  const {files} = useFiles(
    {
      projectName: projectId,
      repoName: repoId,
      commitId: selectedCommitId,
      path: currentPath,
      args: {
        number: MAX_FILES,
      },
    },
    isOpen && isDir,
    0,
  );

  useEffect(() => {
    const handleKeyDown = (evt: KeyboardEvent) => {
      if (evt.key === NAVIGATION_KEYS.OPEN) {
        setIsOpen(true);
        browserHistory.push(
          fileBrowserRoute({
            repoId,
            projectId,
            commitId: selectedCommitId,
            filePath: removeStartingSlash(currentPath),
          }),
        );
      } else if (evt.key === NAVIGATION_KEYS.CLOSE) {
        setIsOpen(false);
        browserHistory.push(
          fileBrowserRoute({
            repoId,
            projectId,
            commitId: selectedCommitId,
            filePath: removeStartingSlash(parentPath),
          }),
        );
      } else if (evt.key === NAVIGATION_KEYS.PREVIOUS) {
        evt.preventDefault();

        if (previousPath) {
          browserHistory.push(
            fileBrowserRoute({
              repoId,
              projectId,
              commitId: selectedCommitId,
              filePath: removeStartingSlash(previousPath),
            }),
          );
        } else {
          browserHistory.push(
            fileBrowserRoute({
              repoId,
              projectId,
              commitId: selectedCommitId,
              filePath: removeStartingSlash(parentPath),
            }),
          );
        }
      } else if (evt.key === NAVIGATION_KEYS.NEXT) {
        evt.preventDefault();

        if (isOpen && files?.[0]?.file?.path) {
          browserHistory.push(
            fileBrowserRoute({
              repoId,
              projectId,
              commitId: selectedCommitId,
              filePath: removeStartingSlash(files[0].file.path),
            }),
          );
        } else if (!isOpen && nextPath) {
          browserHistory.push(
            fileBrowserRoute({
              repoId,
              projectId,
              commitId: selectedCommitId,
              filePath: removeStartingSlash(nextPath),
            }),
          );
        }
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
    filePath,
    nextPath,
    previousPath,
    selectedCommitId,
    browserHistory,
    projectId,
    repoId,
    parentPath,
    currentPath,
    isOpen,
    files,
  ]);

  return {isActive, name: base, nameRef, link, setIsOpen, isOpen};
};

export default useFileTreeDirectory;

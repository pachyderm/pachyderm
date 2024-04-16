import classNames from 'classnames';
import React from 'react';

import {FileType} from '@dash-frontend/api/pfs';
import getParentPath from '@dash-frontend/lib/getParentPath';
import removeStartingSlash from '@dash-frontend/lib/removeStartingSlash';
import {
  ChevronDownSVG,
  ChevronRightSVG,
  FolderSVG,
  DocumentSVG,
  StatusWarningSVG,
  Icon,
  Link,
} from '@pachyderm/components';

import {MAX_FILES} from './constants';
import styles from './DirectoryContents.module.css';
import useFileTree from './hooks/useFileTree';
import useFileTreeDirectory from './hooks/useFileTreeDirectory';
import useFileTreeFile from './hooks/useFileTreeFile';
import {FileTreeNodeProps, FileTreeProps} from './types';

const Directory = ({
  selectedCommitId,
  filePath,
  open,
  nextPath,
  ...props
}: FileTreeNodeProps) => {
  const {isActive, link, setIsOpen, isOpen, nameRef, name} =
    useFileTreeDirectory({
      selectedCommitId,
      filePath,
      open,
      nextPath,
      ...props,
    });

  return (
    <li aria-label={`Directory ${filePath}`}>
      <Link
        to={link}
        className={classNames(
          styles.directory,
          styles.link,
          isActive && styles.active,
        )}
        onClick={() => setIsOpen(true)}
      >
        <Icon
          small
          onClick={(evt) => {
            evt.preventDefault();
            evt.stopPropagation();
            setIsOpen(!isOpen);
          }}
        >
          {isOpen ? <ChevronDownSVG /> : <ChevronRightSVG />}{' '}
        </Icon>
        <div className={styles.link}>
          <Icon small color="black">
            <FolderSVG />
          </Icon>{' '}
          <span className={styles.name} ref={nameRef}>
            {name || '/'}
          </span>
        </div>
      </Link>

      <div className={styles.nested}>
        {isOpen && (
          <DirectoryContents
            selectedCommitId={selectedCommitId}
            path={filePath}
            nextAfterDirectory={nextPath}
          />
        )}
      </div>
    </li>
  );
};

const File = (props: FileTreeNodeProps) => {
  const {isActive, link, name, nameRef} = useFileTreeFile(props);
  return (
    <li aria-label={`File ${name}`}>
      <Link
        className={classNames(
          styles.file,
          styles.link,
          isActive && styles.active,
        )}
        to={link}
      >
        <Icon small>
          <DocumentSVG />
        </Icon>{' '}
        <div className={styles.name} ref={nameRef}>
          {name}
        </div>
      </Link>
    </li>
  );
};

const DirectoryContents = ({
  selectedCommitId = '',
  path = '/',
  open = false,
  initial = false,
  nextAfterDirectory,
}: FileTreeProps) => {
  const {files, filePath, projectId, repoId, cursor, loading} = useFileTree({
    selectedCommitId,
    path,
    initial,
  });

  return (
    <ul className={styles.base}>
      {files?.length ? (
        files.map((file, i) => {
          const parentPath = removeStartingSlash(getParentPath(filePath));
          const currentParentPath = removeStartingSlash(
            getParentPath(file?.file?.path),
          );
          const fileType = file?.fileType;
          const props = {
            filePath: removeStartingSlash(file?.file?.path),
            currentPath: removeStartingSlash(filePath),
            open:
              open &&
              (currentParentPath !== parentPath || file?.file?.path === '/'),
            fileType,
            selectedCommitId,
            projectId,
            repoId,
            nextPath: files?.[i + 1]?.file?.path ?? nextAfterDirectory,
            previousPath: files?.[i - 1]?.file?.path ?? path,
          };

          return fileType === FileType.DIR ? (
            <Directory {...props} key={file?.file?.path} />
          ) : (
            <File {...props} key={file?.file?.path} />
          );
        })
      ) : loading ? (
        <span>Loading...</span>
      ) : (
        <div className={styles.noFiles}>No files found</div>
      )}
      {cursor && (
        <div className={styles.fileLimitReached}>
          <Icon small className={styles.warningIcon}>
            <StatusWarningSVG />
          </Icon>{' '}
          Limited to {MAX_FILES} entries
        </div>
      )}
    </ul>
  );
};

export default DirectoryContents;

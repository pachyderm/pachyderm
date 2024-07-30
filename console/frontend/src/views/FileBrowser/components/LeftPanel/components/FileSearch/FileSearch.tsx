import classNames from 'classnames';
import debounce from 'lodash/debounce';
import noop from 'lodash/noop';
import React, {useCallback, useEffect, useMemo, useRef, useState} from 'react';

import {FileType, FileInfo} from '@dash-frontend/api/pfs';
import {useSearchFiles} from '@dash-frontend/hooks/useSearchFiles';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import removeStartingSlash from '@dash-frontend/lib/removeStartingSlash';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  useOutsideClick,
  Button,
  Icon,
  CloseSVG,
  SearchSVG,
  FolderSVG,
  DocumentSVG,
  Link,
} from '@pachyderm/components';

import styles from './FileSearch.module.css';

type FileSearchProps = {
  onClickFile?: (file: FileInfo) => void;
  selectedCommitId?: string;
};

const FileSearch = ({
  onClickFile = noop,
  selectedCommitId = '',
}: FileSearchProps) => {
  const [searchFilter, setSearchFilter] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const {projectId, repoId} = useUrlState();
  const {files, searchFiles, isFetching} = useSearchFiles({
    commitId: selectedCommitId,
    repoId,
    projectId,
    pattern: searchFilter,
  });
  const debouncedSearchFiles = useMemo(
    () => debounce(searchFiles, 200),
    [searchFiles],
  );
  const handleClickFile = useCallback(
    (file: FileInfo) => {
      setIsOpen(false);
      onClickFile(file);
    },
    [onClickFile, setIsOpen],
  );

  useOutsideClick(containerRef, () => {
    setIsOpen(false);
  });

  useEffect(() => {
    if (searchFilter) {
      debouncedSearchFiles();
    }
  }, [searchFilter, debouncedSearchFiles]);

  return (
    <div className={styles.base} ref={containerRef}>
      <div className={styles.wrapper}>
        <div className={styles.filter}>
          <div className={styles.search}>
            <Icon className={styles.searchIcon} small>
              <SearchSVG aria-hidden />
            </Icon>

            <input
              className={styles.input}
              placeholder="Find files"
              onChange={(e) => {
                setSearchFilter(e.target.value);
                setIsOpen(true);
              }}
              value={searchFilter}
              onFocus={() => searchFilter && setIsOpen(true)}
            />
          </div>
          {searchFilter && (
            <Button
              aria-label="Clear"
              buttonType="ghost"
              onClick={() => {
                setSearchFilter('');
                setIsOpen(false);
              }}
              IconSVG={CloseSVG}
            />
          )}
        </div>
        <div
          className={classNames(
            styles.fileList,
            isOpen ? styles.open : styles.closed,
          )}
        >
          {searchFilter && isFetching && !files?.length && (
            <div className={styles.fileListMessage}>Loading...</div>
          )}
          {files?.length ? (
            files.map((file) => {
              const link = fileBrowserRoute({
                repoId,
                projectId,
                commitId: selectedCommitId,
                filePath: removeStartingSlash(file?.file?.path),
              });

              return (
                <div className={styles.fileListItem} key={file?.file?.path}>
                  <Icon small>
                    {file?.fileType === FileType.DIR ? (
                      <FolderSVG />
                    ) : (
                      <DocumentSVG />
                    )}
                  </Icon>
                  <Link
                    className={styles.link}
                    to={link}
                    onClick={() => handleClickFile(file)}
                  >
                    {file?.file?.path || '/'}
                  </Link>
                </div>
              );
            })
          ) : (
            <div className={styles.fileListMessage}>No files found</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default FileSearch;

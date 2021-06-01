import {Group, ArrowSVG, Link, Search, Tooltip} from '@pachyderm/components';
import findIndex from 'lodash/findIndex';
import React, {useMemo} from 'react';

import CommitId from '@dash-frontend/components/CommitId';
import Header from '@dash-frontend/components/Header';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';

import styles from './FileHeader.module.css';

type FileHeaderProps = {
  fileFilter: string;
  setFileFilter: React.Dispatch<React.SetStateAction<string>>;
};

const FileHeader: React.FC<FileHeaderProps> = ({fileFilter, setFileFilter}) => {
  const {repoId, commitId, branchId, projectId} = useUrlState();
  const {repo} = useCurrentRepo();

  const {previousCommit, nextCommit} = useMemo(() => {
    const reposInBranch =
      repo?.commits.filter((commit) => commit.branch?.name === branchId) || [];
    const index = findIndex(reposInBranch, (commit) => commit.id === commitId);

    const previousCommit =
      reposInBranch.length > index + 1 ? reposInBranch[index + 1].id : null;
    const nextCommit = index > 0 ? reposInBranch[index - 1].id : null;

    return {previousCommit, nextCommit};
  }, [branchId, commitId, repo?.commits]);

  return (
    <Header appearance="light">
      <Group spacing={32} align="center" className={styles.header}>
        <h4 className={styles.commit}>
          {repoId}:{branchId}@<CommitId commit={commitId} />
        </h4>
        {previousCommit && (
          <Tooltip
            tooltipText="See Previous Commit"
            placement="bottom"
            tooltipKey="See Previous Commit"
          >
            <Link
              className={styles.navigate}
              to={fileBrowserRoute({
                repoId,
                branchId,
                projectId,
                commitId: previousCommit,
              })}
            >
              <ArrowSVG className={styles.directionLeft} />
              Previous
            </Link>
          </Tooltip>
        )}
        {nextCommit && (
          <Tooltip
            tooltipText="See Next Commit"
            placement="bottom"
            tooltipKey="See Next Commit"
          >
            <Link
              className={styles.navigate}
              to={fileBrowserRoute({
                repoId,
                branchId,
                projectId,
                commitId: nextCommit,
              })}
            >
              Next
              <ArrowSVG className={styles.directionRight} />
            </Link>
          </Tooltip>
        )}
        <Search
          value={fileFilter}
          className={styles.search}
          placeholder="Filter current folder by name"
          onSearch={setFileFilter}
          aria-label="search files"
        />
      </Group>
    </Header>
  );
};

export default FileHeader;

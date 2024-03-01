import React from 'react';
import {useRouteMatch} from 'react-router-dom';

import {CommitInfo} from '@dash-frontend/api/pfs';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {FormattedFileDiff} from '@dash-frontend/lib/types';
import {LINEAGE_REPO_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {CaptionTextSmall, Link, LoadingDots} from '@pachyderm/components';

import styles from './CommitDetails.module.css';

type CommitDetailsProps = {
  commit: CommitInfo;
  commitDiff?: FormattedFileDiff['diff'];
  diffLoading: boolean;
};

const getDeltaSymbol = (val?: number, symbol?: boolean) => {
  if (val && val > 0) {
    return symbol ? '+' : 'Positive';
  } else if (val && val < 0) {
    return symbol ? '-' : 'Negative';
  } else {
    return '';
  }
};

const CommitDetails: React.FC<CommitDetailsProps> = ({
  commit,
  commitDiff,
  diffLoading,
}) => {
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const isLineageRepoRoute = useRouteMatch({
    path: LINEAGE_REPO_PATH,
    exact: true,
  });

  if (!commit) return null;

  const diffFiles = [
    {label: 'New', ...commitDiff?.filesAdded},
    {label: 'Deleted', ...commitDiff?.filesDeleted},
    {label: 'Updated', ...commitDiff?.filesUpdated},
  ];

  const parentCommitSize = formatBytes(
    Number(commit.details?.sizeBytes || 0) - (commitDiff?.size || 0),
  );

  const parentCommitLinkText = `${commit.parentCommit?.id?.slice(0, 6)}...${
    commit.commit?.branch?.name ? `@${commit.commit?.branch?.name}` : ''
  }`;
  const parentCommitParams = {
    projectId: commit.commit?.repo?.project?.name || '',
    repoId: commit.commit?.repo?.name || '',
    commitId: commit.parentCommit?.id || '',
  };
  const isCommitOpen = !commit?.finished;

  return (
    <div className={styles.base}>
      <div className={styles.commitCardHeader}>
        {formatBytes(commit.details?.sizeBytes)}
        <CaptionTextSmall>
          {isCommitOpen ? (
            'Commit Open'
          ) : (
            <>
              {parentCommitSize}{' '}
              {commitDiff?.size !== 0 && (
                <CaptionTextSmall
                  className={styles[`delta${getDeltaSymbol(commitDiff?.size)}`]}
                >
                  {getDeltaSymbol(commitDiff?.size, true)}{' '}
                  {commitDiff?.sizeDisplay?.replace('-', '')}
                </CaptionTextSmall>
              )}
            </>
          )}
        </CaptionTextSmall>
      </div>
      <div className={styles.commitCardHeader}>
        {commit.parentCommit?.id && (
          <div className={styles.parentCommit}>
            <CaptionTextSmall>Parent Commit</CaptionTextSmall>
            <Link
              to={
                isLineageRepoRoute
                  ? getPathToFileBrowser(parentCommitParams)
                  : fileBrowserRoute(parentCommitParams)
              }
              aria-label="Inspect Parent Commit"
            >
              {parentCommitLinkText}
            </Link>
          </div>
        )}
      </div>
      <div className={styles.commitCardBody}>
        {diffLoading && (
          <div className={styles.diffLoadingContainer}>
            <LoadingDots />
          </div>
        )}
        {diffFiles.map(({label, count, sizeDelta}) =>
          count && count > 0 ? (
            <div className={styles.commitCardMetric} key={label}>
              <div
                data-testid={`InfoPanel__${label.toLowerCase()}`}
                className={styles.number}
              >
                {count}
                <CaptionTextSmall
                  className={styles[`delta${getDeltaSymbol(sizeDelta)}`]}
                >
                  {' '}
                  ({getDeltaSymbol(sizeDelta, true)}
                  {formatBytes(sizeDelta).replace('-', '')})
                </CaptionTextSmall>
              </div>
              <CaptionTextSmall>{label}</CaptionTextSmall>
            </div>
          ) : null,
        )}
      </div>
    </div>
  );
};

export default CommitDetails;

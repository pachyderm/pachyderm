import formatBytes from '@dash-backend/lib/formatBytes';
import {RepoQuery, CommitQuery, CommitDiffQuery} from '@graphqlTypes';
import React from 'react';

import {CaptionTextSmall, LoadingDots} from '@pachyderm/components';

import styles from './CommitDetails.module.css';

type CommitDetailsProps = {
  repo?: RepoQuery['repo'];
  commit: CommitQuery['commit'];
  commitDiff?: CommitDiffQuery['commitDiff'];
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
  if (!commit) return null;

  const diffFiles = [
    {label: 'New', ...commitDiff?.filesAdded},
    {label: 'Deleted', ...commitDiff?.filesDeleted},
    {label: 'Updated', ...commitDiff?.filesUpdated},
  ];

  const parentCommitSize = formatBytes(
    commit.sizeBytes - (commitDiff?.size || 0),
  );

  return (
    <div className={styles.base}>
      <div className={styles.commitCardHeader}>
        {commit.sizeDisplay}
        <CaptionTextSmall>
          {parentCommitSize}{' '}
          {commitDiff?.size !== 0 && (
            <CaptionTextSmall
              className={styles[`delta${getDeltaSymbol(commitDiff?.size)}`]}
            >
              {getDeltaSymbol(commitDiff?.size, true)}{' '}
              {commitDiff?.sizeDisplay?.replace('-', '')}
            </CaptionTextSmall>
          )}
        </CaptionTextSmall>
      </div>
      <div className={styles.commitCardHeader}>
        <CaptionTextSmall>@{commit.branch?.name}</CaptionTextSmall>
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
                  {formatBytes(sizeDelta || 0).replace('-', '')})
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

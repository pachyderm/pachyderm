import React from 'react';

import {CommitInfo} from '@dash-frontend/api/pfs';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {FormattedFileDiff} from '@dash-frontend/lib/types';
import {CaptionTextSmall, LoadingDots} from '@pachyderm/components';

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
  if (!commit) return null;

  const diffFiles = [
    {label: 'New', ...commitDiff?.filesAdded},
    {label: 'Deleted', ...commitDiff?.filesDeleted},
    {label: 'Updated', ...commitDiff?.filesUpdated},
  ];

  const parentCommitSize = formatBytes(
    Number(commit.details?.sizeBytes || 0) - (commitDiff?.size || 0),
  );

  return (
    <div className={styles.base}>
      <div className={styles.commitCardHeader}>
        {formatBytes(commit.details?.sizeBytes)}
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
        <CaptionTextSmall>@{commit.commit?.branch?.name}</CaptionTextSmall>
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

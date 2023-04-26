import formatBytes from '@dash-backend/lib/formatBytes';
import {RepoQuery, CommitQuery} from '@graphqlTypes';
import React from 'react';

import {CaptionTextSmall} from '@pachyderm/components';

import styles from './CommitDetails.module.css';

type CommitDetailsProps = {
  repo?: RepoQuery['repo'];
  commit: CommitQuery['commit'];
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

const CommitDetails: React.FC<CommitDetailsProps> = ({commit}) => {
  if (!commit) return null;

  const diffFiles = [
    {label: 'New', ...commit.diff?.filesAdded},
    {label: 'Deleted', ...commit.diff?.filesDeleted},
    {label: 'Updated', ...commit.diff?.filesUpdated},
  ];

  const parentCommitSize = formatBytes(
    commit.sizeBytes - (commit.diff?.size || 0),
  );

  return (
    <div className={styles.base}>
      <div className={styles.commitCardHeader}>
        {commit.sizeDisplay}
        <CaptionTextSmall>
          {parentCommitSize}{' '}
          {commit.diff?.size !== 0 && (
            <CaptionTextSmall
              className={styles[`delta${getDeltaSymbol(commit.diff?.size)}`]}
            >
              {getDeltaSymbol(commit.diff?.size, true)}{' '}
              {commit.diff?.sizeDisplay?.replace('-', '')}
            </CaptionTextSmall>
          )}
        </CaptionTextSmall>
      </div>
      <div className={styles.commitCardHeader}>
        <CaptionTextSmall>@{commit.branch?.name}</CaptionTextSmall>
      </div>
      <div className={styles.commitCardBody}>
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

import classnames from 'classnames';
import range from 'lodash/range';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import {
  Button,
  CaptionTextSmall,
  SkeletonDisplayText,
} from '@pachyderm/components';

import FileHistoryItem from './components/FileHistoryItem';
import styles from './FileHistory.module.css';
import useFileHistory from './hooks/useFileHistory';

const FileHistory: React.FC = () => {
  const {findCommits, loading, commitList, dateRange, disableSearch, error} =
    useFileHistory();

  if (error) {
    return (
      <div className={styles.base}>
        <EmptyState
          error
          title="We couldn't load the version history for this file"
          message="Please try refreshing the page."
        />
      </div>
    );
  }

  return (
    <div className={styles.base}>
      <CaptionTextSmall>File Versions</CaptionTextSmall>
      <div className={styles.list} data-testid="FileHistory__commitList">
        {commitList?.map((commit) => (
          <FileHistoryItem key={commit.commit?.id} commit={commit} />
        ))}

        {loading && (
          <>
            {range(5).map((i) => (
              <div
                className={classnames(styles.listItem, styles.skeletonWrapper)}
                key={i}
              >
                <div className={styles.leftSkeleton}>
                  <SkeletonDisplayText className={styles.skeletonDisplayText} />
                </div>
                <div className={styles.rightSkeleton}>
                  <SkeletonDisplayText className={styles.skeletonDisplayText} />
                </div>
              </div>
            ))}
          </>
        )}
      </div>

      {loading && (
        <div className={styles.loadingText}>Loading older file versions</div>
      )}

      {!loading && (
        <>
          {dateRange && (
            <div className={styles.versionDates}>
              <span>File versions from:</span>
              <span>{dateRange}</span>
            </div>
          )}
          <Button
            disabled={disableSearch}
            buttonType="secondary"
            className={styles.button}
            onClick={() => findCommits()}
          >
            Load older file versions
          </Button>
        </>
      )}
    </div>
  );
};

export default FileHistory;

import {FileCommitState} from '@graphqlTypes';
import classnames from 'classnames';
import capitalize from 'lodash/capitalize';
import range from 'lodash/range';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  Button,
  CaptionTextSmall,
  Link,
  CaptionText,
  SkeletonDisplayText,
} from '@pachyderm/components';

import styles from './FileHistory.module.css';
import useFileHistory from './hooks/useFileHistory';

const FileHistory: React.FC = () => {
  const {repoId, branchId, projectId, filePath, commitId} = useUrlState();

  const {
    findCommits,
    loading,
    commitList,
    getPathToFileBrowser,
    dateRange,
    disableSearch,
    lazyQueryArgs,
    error,
  } = useFileHistory();

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
          <Link
            key={commit.id}
            className={classnames(styles.listItem, styles.link, {
              [styles.selected]: commit.id === commitId,
            })}
            to={
              commit.commitAction &&
              commit.commitAction !== FileCommitState.DELETED
                ? getPathToFileBrowser({
                    projectId,
                    repoId: repoId,
                    commitId: commit.id,
                    branchId,
                    filePath,
                  })
                : undefined
            }
          >
            <div className={styles.dateAndIdWrapper}>
              <div className={styles.dateAndStatus}>
                <span>{getStandardDate(commit.started)}</span>
                <CaptionTextSmall
                  className={classnames({
                    [styles.deleted]:
                      commit.commitAction === FileCommitState.DELETED,
                  })}
                >
                  {capitalize(commit.commitAction || '-')}
                </CaptionTextSmall>
              </div>
              <div className={styles.textWrapper}>
                <CaptionText className={styles.commitId}>
                  {commit.id}
                </CaptionText>
              </div>
            </div>
            {commit.description && (
              <div className={styles.description}>{commit.description}</div>
            )}
          </Link>
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
            onClick={() => findCommits(lazyQueryArgs)}
          >
            Load older file versions
          </Button>
        </>
      )}
    </div>
  );
};

export default FileHistory;

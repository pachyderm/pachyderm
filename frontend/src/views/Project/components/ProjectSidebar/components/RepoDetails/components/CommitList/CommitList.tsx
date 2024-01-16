import React from 'react';

import {RepoInfo} from '@dash-frontend/api/pfs';
import {BrandedErrorIcon} from '@dash-frontend/components/BrandedIcon';
import {
  getStandardDateFromISOString,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {
  LoadingDots,
  CaptionTextSmall,
  Group,
  Button,
} from '@pachyderm/components';

import styles from './CommitList.module.css';
import useCommitList from './hooks/useCommitList';

const errorMessage = `Unable to load the latest commits`;
const errorMessageAction = `Your commits have been processed, but we couldn’t fetch a list of them from our end. Please try refreshing this page. If this issue keeps happening, contact our customer team.`;

type CommitListProps = {
  repo?: RepoInfo;
};

const CommitList: React.FC<CommitListProps> = ({repo}) => {
  const {
    loading,
    error,
    commits,
    previousCommits,
    getPathToFileBrowser,
    projectId,
    repoId,
  } = useCommitList(repo);

  if (!loading && error) {
    return (
      <div className={styles.errorMessage}>
        <BrandedErrorIcon className={styles.errorIcon} disableDefaultStyling />
        <h5>{errorMessage}</h5>
        <p>{errorMessageAction}</p>
      </div>
    );
  }

  if (!commits?.length) {
    return null;
  }

  return (
    <div className={styles.commits}>
      <Group className={styles.commitListDetails} align="center">
        {!loading &&
          (previousCommits?.length ? (
            <CaptionTextSmall>
              Last {previousCommits?.length} commit
              {(previousCommits?.length || 0) > 1 ? 's' : ''}
            </CaptionTextSmall>
          ) : (
            <CaptionTextSmall>No other commits found</CaptionTextSmall>
          ))}
      </Group>
      {loading ? (
        <div
          data-testid="CommitList__loadingdots"
          className={styles.loadingDots}
        >
          <LoadingDots />
        </div>
      ) : (
        previousCommits?.map((commit) => {
          return (
            <div
              className={styles.commit}
              key={commit.commit?.id}
              data-testid="CommitList__commit"
            >
              <CaptionTextSmall className={styles.commitText}>
                {getStandardDateFromISOString(commit.started)}
              </CaptionTextSmall>
              <CaptionTextSmall
                className={styles.commitText}
              >{`${commit.commit?.id?.slice(0, 6)}...${
                commit.commit?.branch?.name
                  ? `@${commit.commit?.branch?.name}`
                  : ''
              } | ${commit.origin?.kind?.toLowerCase()}`}</CaptionTextSmall>
              <span className={styles.bottomContent}>
                <span>{formatBytes(commit.details?.sizeBytes)}</span>
                {getUnixSecondsFromISOString(commit.finished) > 0 && (
                  <Button
                    buttonType="ghost"
                    to={getPathToFileBrowser({
                      projectId,
                      repoId,
                      commitId: commit.commit?.id || '',
                    })}
                  >
                    Inspect Commit
                  </Button>
                )}
              </span>
            </div>
          );
        })
      )}
    </div>
  );
};

export default CommitList;

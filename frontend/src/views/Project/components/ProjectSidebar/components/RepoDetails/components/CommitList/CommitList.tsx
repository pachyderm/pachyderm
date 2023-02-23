import {RepoQuery} from '@graphqlTypes';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';

import {standardFormat} from '@dash-frontend/constants/dateFormats';
import {
  LoadingDots,
  CaptionTextSmall,
  Group,
  DropdownItem,
  DefaultDropdown,
  ElephantEmptyState,
} from '@pachyderm/components';

import styles from './CommitList.module.css';
import useCommitList from './hooks/useCommitList';

const errorMessage = `Unable to load the latest commits`;
const errorMessageAction = `Your commits have been processed, but we couldn’t fetch a list of them from our end. Please try refreshing this page. If this issue keeps happening, contact our customer team.`;

const originFilters: DropdownItem[] = [
  {
    id: 'all-origin',
    content: 'All origins',
    closeOnClick: true,
  },
  {
    id: 'user-commits',
    content: 'User commits only',
    closeOnClick: true,
  },
];

type CommitListProps = {
  repo?: RepoQuery['repo'];
};

const CommitList: React.FC<CommitListProps> = ({repo}) => {
  const {
    loading,
    error,
    commits,
    handleOriginFilter,
    previousCommits,
    hideAutoCommits,
  } = useCommitList(repo);

  if (!loading && error) {
    return (
      <div className={styles.errorMessage}>
        <ElephantEmptyState className={styles.errorElephant} />
        <h5>{errorMessage}</h5>
        <p>{errorMessageAction}</p>
      </div>
    );
  }

  if (repo?.branches.length === 0 && !commits?.length) {
    return null;
  }

  return (
    <div className={styles.commits}>
      <Group className={styles.commitListDetails} align="center">
        <DefaultDropdown
          items={originFilters}
          onSelect={handleOriginFilter}
          menuOpts={{pin: 'right'}}
        >
          {hideAutoCommits ? 'User commits only' : 'All origins'}
        </DefaultDropdown>
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
        previousCommits.map((commit) => {
          return (
            <div
              className={styles.commit}
              key={commit.id}
              data-testid="CommitList__commit"
            >
              <CaptionTextSmall className={styles.commitText}>
                {format(fromUnixTime(commit.started), standardFormat)}
              </CaptionTextSmall>
              <CaptionTextSmall
                className={styles.commitText}
              >{`${commit.id.slice(0, 5)}...@${
                commit.branch?.name
              } | ${commit.originKind?.toLowerCase()}`}</CaptionTextSmall>
              {commit.sizeDisplay}
            </div>
          );
        })
      )}
    </div>
  );
};

export default CommitList;

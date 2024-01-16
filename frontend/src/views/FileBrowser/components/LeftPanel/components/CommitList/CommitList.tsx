import React from 'react';

import {SimplePager} from '@dash-frontend/../components/src/Pager';
import EmptyState from '@dash-frontend/components/EmptyState';
import ListItem from '@dash-frontend/components/ListItem';
import {CommitInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {COMMIT_PAGE_SIZE} from '@dash-frontend/views/FileBrowser/constants/FileBrowser';
import {
  CaretRightSVG,
  CloseSVG,
  Icon,
  LoadingDots,
  Search,
  StatusStopSVG,
} from '@pachyderm/components';

import styles from './CommitList.module.css';
import useCommitList from './hooks/useCommitList';

export type CommitListProps = {
  selectedCommitId?: string;
  commits?: CommitInfo[];
  loading: boolean;
  page: number;
  setPage: React.Dispatch<React.SetStateAction<number>>;
  hasNextPage?: boolean | null;
  contentLength: number;
};

const CommitList: React.FC<CommitListProps> = ({
  selectedCommitId,
  commits,
  loading,
  page,
  setPage,
  hasNextPage,
  contentLength,
}) => {
  const {
    updateSelectedCommit,
    displayCommits,
    searchValue,
    setSearchValue,
    clearSearch,
    showNoSearchResults,
    isSearchValid,
  } = useCommitList(commits);

  return (
    <div data-testid="CommitList__list" className={styles.base} role="list">
      <div className={styles.header}>
        <SimplePager
          page={page}
          updatePage={setPage}
          nextPageDisabled={!hasNextPage}
          pageSize={COMMIT_PAGE_SIZE}
          contentLength={contentLength}
          elementName="Commit"
        />

        <div className={styles.search}>
          <Search
            data-testid="CommitList__search"
            value={searchValue}
            placeholder="Enter exact commit ID"
            onSearch={setSearchValue}
            className={styles.searchInput}
          />
          {searchValue && (
            <Icon small color="black" className={styles.searchClose}>
              <CloseSVG
                onClick={clearSearch}
                data-testid="CommitList__searchClear"
              />
            </Icon>
          )}
        </div>

        {searchValue && !isSearchValid && (
          <ListItem LeftIconSVG={StatusStopSVG} text="Enter exact commit ID" />
        )}
        {showNoSearchResults && (
          <ListItem
            LeftIconSVG={StatusStopSVG}
            text="No matching commits found"
          />
        )}
      </div>

      {loading ? (
        <LoadingDots />
      ) : displayCommits && displayCommits.length !== 0 ? (
        displayCommits.map((commit) => {
          const selected = commit.commit?.id === selectedCommitId;
          const onClick = () =>
            !selected ? updateSelectedCommit(commit.commit?.id || '') : null;
          return (
            <ListItem
              data-testid="CommitList__listItem"
              key={commit.commit?.id}
              state={selected ? 'selected' : 'default'}
              text={getStandardDateFromISOString(commit.started)}
              RightIconSVG={CaretRightSVG}
              captionText={commit.commit?.id}
              onClick={onClick}
              role="listitem"
            />
          );
        })
      ) : (
        <EmptyState title="" message="No commits found for this repo." />
      )}
    </div>
  );
};

export default CommitList;

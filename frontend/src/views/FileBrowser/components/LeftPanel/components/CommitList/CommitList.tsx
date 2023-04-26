import {Commit} from '@graphqlTypes';
import React from 'react';

import {SimplePager} from '@dash-frontend/../components/src/Pager';
import EmptyState from '@dash-frontend/components/EmptyState';
import ListItem from '@dash-frontend/components/ListItem';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
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
  commits?: Commit[];
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
    <div data-testid="CommitList__list" className={styles.base}>
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
        displayCommits.map((commit) => (
          <ListItem
            data-testid="CommitList__listItem"
            key={commit.id}
            state={commit.id === selectedCommitId ? 'selected' : 'default'}
            text={getStandardDate(commit.started)}
            RightIconSVG={CaretRightSVG}
            captionText={commit.id}
            onClick={() => updateSelectedCommit(commit.id, commit.branch?.name)}
          />
        ))
      ) : (
        <EmptyState title="" message="No commits found for this repo." />
      )}
    </div>
  );
};

export default CommitList;

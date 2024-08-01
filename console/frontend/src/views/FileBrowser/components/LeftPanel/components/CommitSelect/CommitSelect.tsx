import React, {useCallback, useMemo, useState} from 'react';
import {useHistory} from 'react-router';

import {CommitInfo} from '@dash-frontend/api/pfs';
import {useCommits} from '@dash-frontend/hooks/useCommits';
import {useCommitSearch} from '@dash-frontend/hooks/useCommitSearch';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {
  fileBrowserRoute,
  projectReposRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  Button,
  CloseSVG,
  Dropdown,
  Icon,
  SearchSVG,
  IdText,
} from '@pachyderm/components';

import styles from './CommitSelect.module.css';

const CommitSelectItem = ({commit}: {commit: CommitInfo}) => {
  return (
    <span className={styles.item}>
      <span>{getStandardDateFromISOString(commit?.started)} </span>
      <IdText>{commit?.commit?.id?.slice(0, 6)}</IdText>
    </span>
  );
};

const CommitSelect = ({selectedCommitId}: {selectedCommitId?: string}) => {
  const [searchFilter, setSearchFilter] = useState('');
  const browserHistory = useHistory();
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();
  const {projectId, repoId} = useUrlState();
  const branchId = searchParams.branchId;
  const {commits} = useCommits({
    projectName: projectId,
    repoName: repoId,
    args: {
      branchName: branchId || '',
    },
  });
  const {commit: currentCommit} = useCommitSearch(
    {
      projectId: projectId,
      repoId: repoId,
      commitId: selectedCommitId || '',
    },
    !selectedCommitId,
  );
  const {commit: searchCommit} = useCommitSearch(
    {
      projectId: projectId,
      repoId: repoId,
      commitId: searchFilter,
    },
    !searchFilter,
  );

  const updateSelectedCommit = useCallback(
    (commitId?: string) => {
      if (!commitId) return;

      setSearchFilter('');

      browserHistory.push(
        `${fileBrowserRoute(
          {
            projectId,
            repoId,
            commitId,
          },
          false,
        )}?${getUpdatedSearchParams({branchId}, false)}`,
      );
    },
    [branchId, projectId, repoId, browserHistory, getUpdatedSearchParams],
  );

  const filteredCommits = useMemo(() => {
    if (!searchFilter) return commits;

    return searchCommit ? [searchCommit] : [];
  }, [commits, searchCommit, searchFilter]);

  return (
    <Dropdown className={styles.base}>
      <Dropdown.Button buttonType="input" data-testid="CommitSelect__button">
        {currentCommit ? (
          <CommitSelectItem commit={currentCommit} />
        ) : (
          'Select Commit'
        )}
      </Dropdown.Button>
      <Dropdown.Menu className={styles.menu}>
        <div className={styles.search}>
          <Icon className={styles.searchIcon} small>
            <SearchSVG aria-hidden />
          </Icon>

          <input
            className={styles.input}
            placeholder="Find exact commit id"
            onChange={(e) => setSearchFilter(e.target.value)}
            value={searchFilter}
          />

          {searchFilter && (
            <Button
              aria-label="Clear"
              buttonType="ghost"
              onClick={() => setSearchFilter('')}
              IconSVG={CloseSVG}
            />
          )}
        </div>

        <div className={styles.list}>
          {filteredCommits?.length ? (
            filteredCommits.map((commit) => {
              if (!commit?.commit?.id) {
                return null;
              }

              return (
                <Dropdown.MenuItem
                  closeOnClick
                  onClick={() => updateSelectedCommit(commit?.commit?.id)}
                  key={commit?.commit?.id}
                  id={commit?.commit?.id}
                >
                  <CommitSelectItem commit={commit} />
                </Dropdown.MenuItem>
              );
            })
          ) : (
            <div className={styles.noMatch}>No matching commits</div>
          )}
        </div>

        <div className={styles.allCommits}>
          <Button
            buttonType="secondary"
            to={`${projectReposRoute(
              {
                projectId,
                tabId: 'commits',
              },
              false,
            )}?selectedRepos=${repoId}`}
          >
            View all commits
          </Button>
        </div>
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default CommitSelect;

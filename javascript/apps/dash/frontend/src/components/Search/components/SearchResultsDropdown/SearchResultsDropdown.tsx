import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useSearchResults} from '@dash-frontend/hooks/useSearchResults';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

import {MAX_SEARCH_RESULTS} from '../../constants/searchConstants';
import {useSearch} from '../../hooks/useSearch';
import {NotFoundMessage, SectionHeader} from '../Messaging/Messaging';

import {SearchResultItem, SecondaryAction} from './components/SearchResultItem';
import styles from './SearchResultsDropdown.module.css';

const SearchResultsDropdown: React.FC = () => {
  const {
    closeDropdown,
    debouncedValue,
    clearSearch,
    addToSearchHistory,
    searchValue,
  } = useSearch();

  const onResultSelect = useCallback(
    (value: string) => {
      addToSearchHistory(value);
      clearSearch();
      closeDropdown();
    },
    [addToSearchHistory, clearSearch, closeDropdown],
  );

  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {loading, searchResults} = useSearchResults(
    projectId,
    debouncedValue,
    MAX_SEARCH_RESULTS,
  );

  const repoOnClick = useCallback(
    (repoName: string) => {
      onResultSelect(repoName);
      browserHistory.push(
        repoRoute({
          branchId: 'master',
          projectId,
          repoId: repoName,
        }),
      );
    },
    [browserHistory, onResultSelect, projectId],
  );

  const pipelineOnClick = useCallback(
    (pipelineName: string, tabId?: string) => {
      onResultSelect(pipelineName);
      browserHistory.push(
        pipelineRoute({
          projectId,
          pipelineId: pipelineName,
          tabId: tabId,
        }),
      );
    },
    [browserHistory, onResultSelect, projectId],
  );

  const repoResults = () => {
    if (searchResults && searchResults.repos.length > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Repos</SectionHeader>
          </div>
          <div className={styles.resultsContainer}>
            {searchResults.repos.map((repo) => {
              return (
                <SearchResultItem
                  key={repo.name}
                  title={repo.name}
                  searchValue={debouncedValue}
                  onClick={() => repoOnClick(repo.name)}
                >
                  <SecondaryAction linkText={'See Commits'} />
                </SearchResultItem>
              );
            })}
          </div>
        </>
      );
    }
    return (
      <div className={styles.margin}>
        <NotFoundMessage>No matching repos found.</NotFoundMessage>
      </div>
    );
  };

  const pipelineResults = () => {
    if (searchResults && searchResults.pipelines.length > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Pipelines</SectionHeader>
          </div>
          <div className={styles.resultsContainer}>
            {searchResults.pipelines.map((pipeline) => {
              return (
                <SearchResultItem
                  key={pipeline.name}
                  title={pipeline.name}
                  searchValue={debouncedValue}
                  onClick={() => pipelineOnClick(pipeline.name)}
                >
                  <SecondaryAction
                    linkText={'See Jobs'}
                    onClick={() => pipelineOnClick(pipeline.name, 'jobs')}
                  />
                </SearchResultItem>
              );
            })}
          </div>
        </>
      );
    }
    return (
      <div className={styles.margin}>
        <NotFoundMessage>No matching pipelines found.</NotFoundMessage>
      </div>
    );
  };

  if (loading || debouncedValue !== searchValue) {
    return <></>;
  }

  if (
    searchResults &&
    searchResults.pipelines.length === 0 &&
    searchResults.repos.length === 0 &&
    !searchResults.job
  ) {
    return (
      <div className={styles.noResults}>
        <NotFoundMessage>
          No matching pipelines, jobs or global ID found.
        </NotFoundMessage>
      </div>
    );
  }

  return (
    <div className={styles.base}>
      {repoResults()}
      <hr className={styles.searchHr} />
      {pipelineResults()}
    </div>
  );
};

export default SearchResultsDropdown;

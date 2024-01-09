import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useSearchResults} from '@dash-frontend/hooks/useSearchResults';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  jobRoute,
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
  const {getUpdatedSearchParams} = useUrlQueryState();
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

  const jobSetOnClick = useCallback(
    (jobId: string) => {
      onResultSelect(jobId);
      const newSearchParams = getUpdatedSearchParams({
        selectedJobs: [jobId],
        pipelineStep: [],
      });

      browserHistory.push(`${jobRoute({projectId}, false)}${newSearchParams}`);
    },
    [onResultSelect, getUpdatedSearchParams, browserHistory, projectId],
  );

  const repoResults = useCallback(() => {
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
                  key={repo.repo?.name}
                  title={repo.repo?.name || ''}
                  searchValue={debouncedValue}
                  onClick={() => repoOnClick(repo.repo?.name || '')}
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
  }, [debouncedValue, repoOnClick, searchResults]);

  const pipelineResults = useCallback(() => {
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
                  key={pipeline.pipeline?.name}
                  title={pipeline.pipeline?.name || ''}
                  searchValue={debouncedValue}
                  onClick={() => pipelineOnClick(pipeline.pipeline?.name || '')}
                />
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
  }, [debouncedValue, pipelineOnClick, searchResults]);

  const jobSetResult = useCallback(() => {
    if (searchResults && searchResults.jobSet) {
      const jobSet = searchResults.jobSet;

      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Jobs</SectionHeader>
          </div>
          <div className={styles.resultsContainer}>
            <SearchResultItem
              key={jobSet?.[0]?.job?.id}
              title={jobSet?.[0]?.job?.id || ''}
              searchValue={debouncedValue}
              onClick={() => jobSetOnClick(jobSet[0].job?.id || '')}
            />
          </div>
        </>
      );
    }
    return (
      <div className={styles.margin}>
        <NotFoundMessage>No matching pipelines found.</NotFoundMessage>
      </div>
    );
  }, [searchResults, jobSetOnClick, debouncedValue]);

  if (loading || debouncedValue !== searchValue) {
    return <></>;
  }

  if (
    searchResults &&
    searchResults.pipelines.length === 0 &&
    searchResults.repos.length === 0 &&
    !searchResults.jobSet
  ) {
    return (
      <div className={styles.noResults}>
        <NotFoundMessage>
          No matching pipelines, jobs or Global ID found.
        </NotFoundMessage>
      </div>
    );
  }

  if (searchResults?.jobSet) {
    return <div className={styles.base}>{jobSetResult()}</div>;
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

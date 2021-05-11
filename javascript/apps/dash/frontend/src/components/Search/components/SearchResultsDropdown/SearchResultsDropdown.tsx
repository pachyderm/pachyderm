import React from 'react';

import {useSearchResults} from '@dash-frontend/hooks/useSearchResults';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import {useSearch} from '../../hooks/useSearch';
import {NotFoundMessage, SectionHeader} from '../Messaging/Messaging';

import SearchResultItem from './components';
import styles from './SearchResultsDropdown.module.css';

const SearchResultsDropdown: React.FC = () => {
  const {
    closeDropdown,
    debouncedValue,
    clearSearch,
    addToSearchHistory,
    searchValue,
  } = useSearch();

  const onResultSelect = (value: string) => {
    addToSearchHistory(value);
    clearSearch();
    closeDropdown();
  };

  const {projectId} = useUrlState();
  const {loading, searchResults} = useSearchResults(projectId, debouncedValue);

  const repoResults = () => {
    if (searchResults && searchResults.repos.length > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Repos</SectionHeader>
          </div>
          {searchResults.repos.map((repo) => {
            return (
              <SearchResultItem
                key={repo.name}
                title={repo.name}
                searchValue={debouncedValue}
                linkText={'See Commits'}
                onClick={() => {
                  onResultSelect(repo.name);
                  //TODO: Will add routing in next PR
                }}
              />
            );
          })}
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
          {searchResults.pipelines.map((pipeline) => {
            return (
              <SearchResultItem
                key={pipeline.name}
                title={pipeline.name}
                searchValue={debouncedValue}
                linkText={'See Jobs'}
                onClick={() => {
                  onResultSelect(pipeline.name);
                  //TODO: Will add routing in next PR
                }}
              />
            );
          })}
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

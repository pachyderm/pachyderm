import classNames from 'classnames';
import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {NodeState} from '@dash-frontend/lib/types';
import {useSearchResults} from '@dash-frontend/views/Project/components/ProjectHeader/hooks/useSearchResults';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {CaptionTextSmall, LoadingDots} from '@pachyderm/components';

import {useSearch} from '../../hooks/useSearch';
import {NotFoundMessage} from '../Messaging/Messaging';

import {SearchResultItem, SecondaryAction} from './components/SearchResultItem';
import styles from './SearchResultsDropdown.module.css';

const SearchResultsDropdown: React.FC = () => {
  const {debouncedValue, addToSearchHistory} = useSearch();

  const onResultSelect = useCallback(
    (value: string) => {
      addToSearchHistory(value);
    },
    [addToSearchHistory],
  );

  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {loading, searchResults} = useSearchResults(projectId, debouncedValue);

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

  const pipelineResults = useCallback(() => {
    if (searchResults && searchResults.pipelines.length > 0) {
      return (
        <>
          <CaptionTextSmall
            className={classNames(
              styles.base,
              styles.sectionHeader,
              styles.smallPaddingLeft,
            )}
          >
            Pipelines
          </CaptionTextSmall>
          <div className={styles.resultsContainer}>
            {searchResults.pipelines.map((pipeline) => {
              return (
                <SearchResultItem
                  key={pipeline.pipeline?.name}
                  title={pipeline.pipeline?.name || ''}
                  searchValue={debouncedValue}
                  onClick={() => pipelineOnClick(pipeline.pipeline?.name || '')}
                  hasFailedPipeline={
                    restPipelineStateToNodeState(pipeline.state) ===
                    NodeState.ERROR
                  }
                  hasFailedSubjob={
                    restJobStateToNodeState(pipeline.lastJobState) ===
                    NodeState.ERROR
                  }
                />
              );
            })}
          </div>
        </>
      );
    }
    return (
      <>
        <CaptionTextSmall
          className={classNames(
            styles.base,
            styles.sectionHeader,
            styles.smallPaddingLeft,
          )}
        >
          Pipelines
        </CaptionTextSmall>
        <NotFoundMessage className={styles.smallPaddingLeft}>
          No matching pipelines found.
        </NotFoundMessage>
      </>
    );
  }, [debouncedValue, pipelineOnClick, searchResults]);

  const repoResults = useCallback(() => {
    if (searchResults && searchResults.repos.length > 0) {
      return (
        <>
          <CaptionTextSmall
            className={classNames(
              styles.base,
              styles.sectionHeader,
              styles.smallPaddingLeft,
            )}
          >
            Repos
          </CaptionTextSmall>
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

  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <LoadingDots />
      </div>
    );
  }

  if (
    searchResults &&
    searchResults.pipelines.length === 0 &&
    searchResults.repos.length === 0
  ) {
    return (
      <div className={styles.noResults}>
        <NotFoundMessage>No matching pipelines or repos.</NotFoundMessage>
      </div>
    );
  }

  return (
    <>
      <div>{pipelineResults()}</div>
      <div>{repoResults()}</div>
    </>
  );
};

export default SearchResultsDropdown;

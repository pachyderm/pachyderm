import React from 'react';

import {JobState} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import {NodeState} from '@dash-frontend/lib/types';
import {
  Icon,
  StatusWarningSVG,
  Tabs,
  Search,
  CloseSVG,
  PipelineColorlessSVG,
  LoadingDots,
} from '@pachyderm/components';

import {NotFoundMessage} from '../../../GlobalSearch/components/Messaging';
import {SearchResultItem} from '../../../GlobalSearch/components/SearchResultsDropdown/components';

import styles from './Dropdown.module.css';
import useDropdown from './hooks/useDropdown';

const Dropdown: React.FC = () => {
  const {
    search,
    setSearch,
    initialtab,
    setSelectedTab,
    subJobsHeading,
    pipelinesHeading,
    searchResults,
    lowercaseQuery,
    pipelineOnClick,
    loading,
    selectedTab,
  } = useDropdown();

  return (
    <div className={styles.base}>
      <div className={styles.searchContainer}>
        <div className={styles.search}>
          <Search
            value={search}
            placeholder="Search failed subjobs and pipelines"
            onSearch={setSearch}
            className={styles.searchInput}
          />
        </div>

        {search && (
          <Icon small color="black" className={styles.searchClose}>
            <CloseSVG aria-label="clear search" onClick={() => setSearch('')} />
          </Icon>
        )}
      </div>
      <Tabs
        initialActiveTabId={initialtab}
        onSwitch={(tabId) =>
          setSelectedTab(tabId === 'job' ? 'job' : 'pipeline')
        }
      >
        <Tabs.TabsHeader className={styles.tabsContainer}>
          <Tabs.Tab id="job">
            <div className={styles.tabContainer}>
              <Icon color="red" small>
                <StatusWarningSVG />
              </Icon>
              {subJobsHeading}
            </div>
          </Tabs.Tab>

          <Tabs.Tab id="pipeline">
            <div className={styles.tabContainer}>
              <Icon color="red" small>
                <PipelineColorlessSVG />
              </Icon>
              {pipelinesHeading}
            </div>
          </Tabs.Tab>
        </Tabs.TabsHeader>

        <div className={styles.resultsContainer}>
          {loading ? (
            <div className={styles.loadingContainer}>
              <LoadingDots />
            </div>
          ) : searchResults && searchResults.length > 0 ? (
            searchResults.map((pipeline) => {
              return (
                <SearchResultItem
                  key={pipeline.pipeline?.name}
                  title={pipeline.pipeline?.name || ''}
                  searchValue={lowercaseQuery}
                  onClick={() => pipelineOnClick(pipeline.pipeline?.name || '')}
                  wasKilled={pipeline.lastJobState === JobState.JOB_KILLED}
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
            })
          ) : (
            <div className={styles.noResults}>
              <NotFoundMessage>
                No matching {selectedTab === 'job' ? 'Subjobs' : 'Pipelines'}.
              </NotFoundMessage>
            </div>
          )}
        </div>
      </Tabs>
    </div>
  );
};

export default Dropdown;

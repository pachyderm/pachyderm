import {ButtonLink, Chip, ChipGroup} from '@pachyderm/components';
import React, {useCallback} from 'react';

import {jobStates} from '@dash-frontend/components/JobList/components/JobListStatusFilter/JobListStatusFilter';

import {useDefaultDropdown} from '../../hooks/useDefaultDropdown';
import {useSearch} from '../../hooks/useSearch';
import {NotFoundMessage, SectionHeader} from '../Messaging';

import styles from './DefaultDropdown.module.css';

const DefaultDropdown: React.FC = () => {
  const {
    setSearchValue,
    history,
    clearSearchHistory,
    closeDropdown,
  } = useSearch();
  const {stateCounts, allJobs, routeToJobs} = useDefaultDropdown();

  const handleChipClick = useCallback(() => {
    routeToJobs();
    closeDropdown();
  }, [closeDropdown, routeToJobs]);

  const recentSearch = () => {
    if (history.length > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Recent Searches</SectionHeader>
            <ButtonLink small onClick={clearSearchHistory}>
              Clear
            </ButtonLink>
          </div>
          <div className={styles.recentSearchGroup}>
            <ChipGroup>
              {history.map((item, i) => (
                <Chip
                  key={item}
                  onClick={() => {
                    setSearchValue(item);
                  }}
                >
                  {item}
                </Chip>
              ))}
            </ChipGroup>
          </div>
        </>
      );
    }

    return <NotFoundMessage>There are no recent searches.</NotFoundMessage>;
  };

  //TODO: Route with selected filter
  const renderJobStates = () => {
    if (allJobs > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Last 30 Jobs</SectionHeader>
          </div>
          <div className={styles.jobsGroup}>
            <ChipGroup>
              {jobStates.map((state) => {
                const stateCount = stateCounts[state.value];
                if (!stateCount) {
                  return null;
                }
                return (
                  <Chip key={state.value} onClick={handleChipClick}>
                    {state.label} ({stateCounts[state.value]})
                  </Chip>
                );
              })}
              <Chip name={'All'} onClick={handleChipClick}>
                All ({allJobs})
              </Chip>
            </ChipGroup>
          </div>
        </>
      );
    }

    return (
      <NotFoundMessage>There are no jobs on this project.</NotFoundMessage>
    );
  };
  return (
    <div className={styles.base}>
      {recentSearch()}
      <hr className={styles.searchHr} />
      {renderJobStates()}
    </div>
  );
};

export default DefaultDropdown;

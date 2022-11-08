import React from 'react';

import {jobStates} from '@dash-frontend/components/JobList/components/JobListStatusFilter/JobListStatusFilter';
import {Button, Chip, ChipGroup} from '@pachyderm/components';
import getListTitle from 'lib/getListTitle';

import {useDefaultDropdown} from '../../hooks/useDefaultDropdown';
import {useSearch} from '../../hooks/useSearch';
import {NotFoundMessage, SectionHeader} from '../Messaging';

import styles from './DefaultDropdown.module.css';

const DefaultDropdown: React.FC = () => {
  const {history, clearSearchHistory} = useSearch();
  const {stateCounts, allJobs, handleHistoryChipClick, handleJobChipClick} =
    useDefaultDropdown();

  const recentSearch = () => {
    if (history.length > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>Recent Searches</SectionHeader>
            <Button onClick={clearSearchHistory} buttonType="ghost">
              Clear
            </Button>
          </div>
          <div className={styles.recentSearchGroup}>
            <ChipGroup>
              {history.map((searchValue) => (
                <Chip
                  key={searchValue}
                  onClickValue={searchValue}
                  onClick={handleHistoryChipClick}
                >
                  {searchValue}
                </Chip>
              ))}
            </ChipGroup>
          </div>
        </>
      );
    }

    return <NotFoundMessage>There are no recent searches.</NotFoundMessage>;
  };

  const renderJobStates = () => {
    if (allJobs > 0) {
      return (
        <>
          <div className={styles.sectionHeader}>
            <SectionHeader>{getListTitle('Job', allJobs)}</SectionHeader>
          </div>
          <div className={styles.jobsGroup}>
            <ChipGroup>
              {jobStates.map((state) => {
                const stateCount = stateCounts[state.value];
                if (!stateCount) {
                  return (
                    <Chip key={state.value} disabled>
                      {state.label} (0)
                    </Chip>
                  );
                }
                return (
                  <Chip
                    key={state.value}
                    onClickValue={state.value}
                    onClick={handleJobChipClick}
                  >
                    {state.label} ({stateCounts[state.value]})
                  </Chip>
                );
              })}
              <Chip onClick={handleJobChipClick}>All ({allJobs})</Chip>
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

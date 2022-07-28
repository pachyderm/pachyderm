import {
  JobOverviewFragment,
  JobSetFieldsFragment,
  JobState,
} from '@graphqlTypes';
import {Chip, ChipGroup} from '@pachyderm/components';
import React from 'react';

import {readableJobState} from '@dash-frontend/lib/jobs';
import getListTitle from 'lib/getListTitle';

import {JobFilters} from '../../hooks/useJobFilters';

import useJobListStatusFilter from './hooks/useJobListStatusFilter';
import styles from './JobListStatusFilter.module.css';

export const jobStates = Object.values(JobState).map((state) => ({
  label: readableJobState(state),
  value: state,
}));

interface JobListStatusFilterProps {
  jobs: (JobOverviewFragment | JobSetFieldsFragment)[];
  selectedFilters: JobFilters;
}

const JobListStatusFilter: React.FC<JobListStatusFilterProps> = ({
  jobs,
  selectedFilters,
}) => {
  const {stateCounts, onChipClick} = useJobListStatusFilter(
    jobs,
    selectedFilters,
  );

  return (
    <div className={styles.base}>
      <p className={styles.label}>
        <strong>{getListTitle('Job', jobs.length)}</strong>
      </p>

      <ChipGroup>
        {jobStates.map((state) => {
          const stateCount = stateCounts[state.value];

          if (!stateCount) {
            return null;
          }

          return (
            <Chip
              selected={selectedFilters[state.value]}
              onClickValue={state.value}
              onClick={onChipClick}
              key={state.value}
            >
              {state.label} ({stateCounts[state.value]})
            </Chip>
          );
        })}
      </ChipGroup>
    </div>
  );
};

export default JobListStatusFilter;

import {Chip, ChipGroup} from '@pachyderm/components';
import React from 'react';

import readableJobState from '@dash-frontend/lib/readableJobState';
import {PipelineJobsQuery, PipelineJobState} from '@graphqlTypes';

import {PipelineJobFilters} from '../../hooks/useJobList';

import useJobListStatusFilter from './hooks/useJobListStatusFilter';
import styles from './JobListStatusFilter.module.css';

export const jobStates = Object.values(PipelineJobState).map((state) => ({
  label: readableJobState(state),
  value: state,
}));

interface JobListStatusFilterProps {
  pipelineJobs: PipelineJobsQuery['pipelineJobs'];
  selectedFilters: PipelineJobFilters;
}

const JobListStatusFilter: React.FC<JobListStatusFilterProps> = ({
  pipelineJobs,
  selectedFilters,
}) => {
  const {stateCounts, onChipClick} = useJobListStatusFilter(
    pipelineJobs,
    selectedFilters,
  );

  return (
    <div className={styles.base}>
      <p className={styles.label}>
        {pipelineJobs.length === 1
          ? `Last Job`
          : `Last ${pipelineJobs.length} Jobs`}
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

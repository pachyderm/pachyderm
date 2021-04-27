import {ChipGroup, ChipInput} from '@pachyderm/components';
import countBy from 'lodash/countBy';
import React, {useMemo} from 'react';

import readableJobState from '@dash-frontend/lib/readableJobState';
import {GetJobsQuery, JobState} from '@graphqlTypes';

import styles from './JobListStatusFilter.module.css';

export const jobStates = Object.values(JobState).map((state) => ({
  label: readableJobState(state),
  value: state,
}));

interface JobListStatusFilterProps {
  jobs: GetJobsQuery['jobs'];
}

const JobListStatusFilter: React.FC<JobListStatusFilterProps> = ({jobs}) => {
  const stateCounts = useMemo(() => countBy(jobs, (job) => job.state), [jobs]);

  return (
    <div className={styles.base}>
      <p className={styles.label}>Last {jobs.length} Jobs</p>

      <ChipGroup>
        {jobStates.map((state) => {
          const stateCount = stateCounts[state.value];

          if (!stateCount) {
            return null;
          }

          return (
            <ChipInput key={state.value} name={state.value}>
              {state.label} ({stateCounts[state.value]})
            </ChipInput>
          );
        })}
      </ChipGroup>
    </div>
  );
};

export default JobListStatusFilter;

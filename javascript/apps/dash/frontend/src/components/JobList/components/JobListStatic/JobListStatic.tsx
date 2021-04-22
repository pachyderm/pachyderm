import React from 'react';

import {Job} from '@graphqlTypes';

import JobListEmptyState from './components/JobListEmptyState';
import JobListItem from './components/JobListItem';
import JobListSkeleton from './components/JobListSkeleton';
import styles from './JobListStatic.module.css';

type JobListBaseProps = {
  jobs?: Pick<Job, 'id' | 'state' | 'createdAt'>[];
  loading?: boolean;
  projectId: string;
  expandActions?: boolean;
};

const JobListBase: React.FC<JobListBaseProps> = ({
  loading,
  jobs,
  projectId,
  expandActions = false,
}) => {
  if (loading) return <JobListSkeleton expandActions={expandActions} />;

  if (jobs?.length === 0) return <JobListEmptyState />;

  return (
    <ul className={styles.base} data-testid={`JobList__project${projectId}`}>
      {jobs?.map((job) => (
        <JobListItem job={job} key={job.id} expandActions={expandActions} />
      ))}
    </ul>
  );
};

export default JobListBase;

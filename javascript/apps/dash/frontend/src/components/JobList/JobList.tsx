import React from 'react';

import {useJobs} from '@dash-frontend/hooks/useJobs';

import JobListItem from './components/JobListItem';
import JobListSkeleton from './components/JobListSkeleton';
import styles from './JobList.module.css';

type JobListProps = {
  projectId: string;
  expandActions?: boolean;
};

const JobList: React.FC<JobListProps> = ({
  projectId,
  expandActions = false,
}) => {
  const {jobs, loading} = useJobs(projectId);

  if (loading) return <JobListSkeleton expandActions={expandActions} />;

  return (
    <ul className={styles.base} data-testid={`JobList__project${projectId}`}>
      {jobs.map((job) => (
        <JobListItem job={job} key={job.id} expandActions={expandActions} />
      ))}
    </ul>
  );
};

export default JobList;

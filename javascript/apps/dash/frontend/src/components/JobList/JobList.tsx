import {Form} from '@pachyderm/components';
import React from 'react';

import JobListItem from './components/JobListItem';
import JobListSkeleton from './components/JobListSkeleton';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobList from './hooks/useJobList';
import styles from './JobList.module.css';

type JobListProps = {
  projectId: string;
  expandActions?: boolean;
  showStatusFilter?: boolean;
};

const JobList: React.FC<JobListProps> = ({
  projectId,
  expandActions = false,
  showStatusFilter = false,
}) => {
  const {loading, formCtx, jobs, filteredJobs} = useJobList({projectId});

  if (loading) return <JobListSkeleton expandActions={expandActions} />;

  return (
    <Form formContext={formCtx}>
      {showStatusFilter && <JobListStatusFilter jobs={jobs} />}

      <ul className={styles.base} data-testid={`JobList__project${projectId}`}>
        {filteredJobs.map((job) => (
          <JobListItem job={job} key={job.id} expandActions={expandActions} />
        ))}
      </ul>
    </Form>
  );
};

export default JobList;

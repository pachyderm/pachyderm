import {Form} from '@pachyderm/components';
import React from 'react';

import {GetJobsQueryVariables} from '@graphqlTypes';

import JobListEmptyState from './components/JobListEmptyState';
import JobListItem from './components/JobListItem';
import JobListSkeleton from './components/JobListSkeleton';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobList from './hooks/useJobList';
import styles from './JobList.module.css';

type JobListProps = {
  expandActions?: boolean;
  showStatusFilter?: boolean;
} & GetJobsQueryVariables['args'];

const JobList: React.FC<JobListProps> = ({
  projectId,
  pipelineId,
  expandActions = false,
  showStatusFilter = false,
}) => {
  const {loading, formCtx, jobs, filteredJobs} = useJobList({
    projectId,
    pipelineId,
  });

  if (loading) return <JobListSkeleton expandActions={expandActions} />;

  if (jobs.length === 0) return <JobListEmptyState />;

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

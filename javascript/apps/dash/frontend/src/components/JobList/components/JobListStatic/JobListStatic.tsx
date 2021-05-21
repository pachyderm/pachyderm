import classnames from 'classnames';
import React from 'react';

import {PipelineJob} from '@graphqlTypes';

import JobListEmptyState from './components/JobListEmptyState';
import JobListItem from './components/JobListItem';
import JobListSkeleton from './components/JobListSkeleton';
import styles from './JobListStatic.module.css';

type JobListBaseProps = {
  pipelineJobs?: Pick<PipelineJob, 'id' | 'state' | 'createdAt'>[];
  loading?: boolean;
  projectId: string;
  expandActions?: boolean;
  listScroll?: boolean;
};

const JobListBase: React.FC<JobListBaseProps> = ({
  loading,
  pipelineJobs,
  projectId,
  expandActions = false,
  listScroll = false,
}) => {
  if (loading) return <JobListSkeleton expandActions={expandActions} />;

  if (pipelineJobs?.length === 0) return <JobListEmptyState />;

  return (
    <ul
      className={classnames(styles.base, {[styles.listScroll]: listScroll})}
      data-testid={`JobList__project${projectId}`}
    >
      {pipelineJobs?.map((job) => (
        <JobListItem
          pipelineJob={job}
          key={job.id}
          expandActions={expandActions}
        />
      ))}
    </ul>
  );
};

export default JobListBase;

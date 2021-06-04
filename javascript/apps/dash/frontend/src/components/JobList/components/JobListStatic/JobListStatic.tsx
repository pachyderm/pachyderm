import classnames from 'classnames';
import React from 'react';

import {PipelineJobOverviewFragment} from '@graphqlTypes';

import ListEmptyState from '../../../ListEmptyState';

import JobListItem from './components/JobListItem';
import JobListSkeleton from './components/JobListSkeleton';
import styles from './JobListStatic.module.css';

type JobListBaseProps = {
  pipelineJobs?: PipelineJobOverviewFragment[];
  loading?: boolean;
  projectId: string;
  expandActions?: boolean;
  listScroll?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
};

const JobListBase: React.FC<JobListBaseProps> = ({
  loading,
  pipelineJobs,
  projectId,
  expandActions = false,
  listScroll = false,
  emptyStateTitle,
  emptyStateMessage,
}) => {
  if (loading) return <JobListSkeleton expandActions={expandActions} />;

  if (pipelineJobs?.length === 0)
    return (
      <ListEmptyState title={emptyStateTitle} message={emptyStateMessage} />
    );

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

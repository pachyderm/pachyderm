import {LoadingDots} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import {JobOverviewFragment} from '@graphqlTypes';

import ListEmptyState from '../../../ListEmptyState';

import JobListItem from './components/JobListItem';
import styles from './JobListStatic.module.css';

type JobListBaseProps = {
  jobs?: JobOverviewFragment[];
  loading?: boolean;
  projectId: string;
  expandActions?: boolean;
  listScroll?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
};

const JobListBase: React.FC<JobListBaseProps> = ({
  loading,
  jobs,
  projectId,
  expandActions = false,
  listScroll = false,
  emptyStateTitle,
  emptyStateMessage,
}) => {
  if (loading)
    return (
      <div data-testid="JobListStatic__loadingdots">
        <LoadingDots />
      </div>
    );

  if (jobs?.length === 0)
    return (
      <ListEmptyState title={emptyStateTitle} message={emptyStateMessage} />
    );

  return (
    <ul
      className={classnames(styles.base, {[styles.listScroll]: listScroll})}
      data-testid={`JobList__project${projectId}`}
    >
      {jobs?.map((job) => (
        <JobListItem job={job} key={job.id} expandActions={expandActions} />
      ))}
    </ul>
  );
};

export default JobListBase;

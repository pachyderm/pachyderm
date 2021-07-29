import {JobOverviewFragment, JobSetFieldsFragment} from '@graphqlTypes';
import {LoadingDots} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import EmptyState from '../../../EmptyState';

import JobListItem from './components/JobListItem';
import styles from './JobListStatic.module.css';
import isPipelineJob from './utils/isPipelineJob';

type JobListBaseProps = {
  jobs?: (JobOverviewFragment | JobSetFieldsFragment)[];
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
    return <EmptyState title={emptyStateTitle} message={emptyStateMessage} />;

  // deriving the 'key' field from pipelineName and ID should no longer be
  // necessary once we are using jobSets and pipeline jobs. Currently,
  // JobList is being used to display "all" pipeline jobs in the project
  return (
    <ul
      className={classnames(styles.base, {[styles.listScroll]: listScroll})}
      data-testid={`JobList__project${projectId}`}
    >
      {jobs?.map((job) => (
        <JobListItem
          job={job}
          projectId={projectId}
          key={`${job.id}${isPipelineJob(job) ? `__${job.pipelineName}` : ''}`}
          expandActions={expandActions}
        />
      ))}
    </ul>
  );
};

export default JobListBase;

import {ApolloError} from '@apollo/client';
import {JobOverviewFragment, JobSetFieldsFragment} from '@graphqlTypes';
import classnames from 'classnames';
import React from 'react';

import {LoadingDots} from '@pachyderm/components';

import EmptyState from '../../../EmptyState';

import JobListItem from './components/JobListItem';
import styles from './JobListStatic.module.css';
import isPipelineJob from './utils/isPipelineJob';

const errorMessage = `Sorry! We're currently having trouble loading the jobs list.`;
const errorMessageAction = 'Please refresh the page';

type JobListBaseProps = {
  jobs?: (JobOverviewFragment | JobSetFieldsFragment)[];
  loading?: boolean;
  projectId: string;
  expandActions?: boolean;
  listScroll?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
  cardStyle?: boolean;
  error?: ApolloError;
};

const JobListBase: React.FC<JobListBaseProps> = ({
  loading,
  jobs,
  projectId,
  expandActions = false,
  listScroll = false,
  emptyStateTitle,
  emptyStateMessage,
  cardStyle,
  error,
}) => {
  if (loading)
    return (
      <div
        data-testid="JobListStatic__loadingdots"
        className={styles.loadingDots}
      >
        <LoadingDots />
      </div>
    );

  if (error)
    return (
      <EmptyState title={errorMessage} message={errorMessageAction} error />
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
          cardStyle={cardStyle}
        />
      ))}
    </ul>
  );
};

export default JobListBase;

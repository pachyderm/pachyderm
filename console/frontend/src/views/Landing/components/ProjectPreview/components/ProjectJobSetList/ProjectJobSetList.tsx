import React from 'react';

import {JobInfo} from '@dash-frontend/api/pps';
import EmptyState from '@dash-frontend/components/EmptyState';
import {LoadingDots} from '@pachyderm/components';

import JobListItem from './components/JobListItem';
import styles from './ProjectJobSetList.module.css';

const errorMessage = `Sorry! We're currently having trouble loading the list of jobs.`;
const errorMessageAction = 'Please refresh the page';

type ProjectJobSetListProps = {
  jobs?: JobInfo[];
  loading?: boolean;
  projectId: string;
  error?: string;
  emptyStateTitle: string;
  emptyStateMessage: string;
};

const ProjectJobSetList: React.FC<ProjectJobSetListProps> = ({
  loading,
  jobs,
  emptyStateTitle,
  emptyStateMessage,
  projectId,
  error,
}) => {
  if (loading)
    return (
      <div
        data-testid="ProjectJobSetList__loadingdots"
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

  return (
    <ul
      className={styles.base}
      data-testid={`ProjectJobSetList__project${projectId}`}
    >
      {jobs?.map((job) => (
        <JobListItem job={job} projectId={projectId} key={job.job?.id} />
      ))}
    </ul>
  );
};

export default ProjectJobSetList;

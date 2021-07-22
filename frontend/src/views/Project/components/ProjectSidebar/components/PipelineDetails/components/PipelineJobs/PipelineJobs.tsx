import React from 'react';
import {Redirect} from 'react-router';

import {LETS_START_TITLE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import JobList from '@dash-frontend/components/JobList';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import styles from './PipelineJobs.module.css';

const emptyJobListMessage = 'Create your first job on this pipeline!';

const PipelineJobs = () => {
  const {projectId, pipelineId} = useUrlState();
  const {isServiceOrSpout} = useCurrentPipeline();

  if (isServiceOrSpout) {
    return <Redirect to={pipelineRoute({pipelineId, projectId})} />;
  }

  return (
    <div className={styles.base}>
      <JobList
        projectId={projectId}
        pipelineId={pipelineId}
        emptyStateTitle={LETS_START_TITLE}
        emptyStateMessage={emptyJobListMessage}
      />
    </div>
  );
};

export default PipelineJobs;

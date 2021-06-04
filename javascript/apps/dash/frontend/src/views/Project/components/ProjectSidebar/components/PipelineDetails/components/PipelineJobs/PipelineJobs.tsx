import React from 'react';

import JobList from '@dash-frontend/components/JobList';
import {LETS_START_TITLE} from '@dash-frontend/components/ListEmptyState/constants/ListEmptyStateConstants';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import styles from './PipelineJobs.module.css';

const emptyJobListMessage = 'Create your first job on this pipeline!';

const PipelineJobs = () => {
  const {projectId, pipelineId} = useUrlState();

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

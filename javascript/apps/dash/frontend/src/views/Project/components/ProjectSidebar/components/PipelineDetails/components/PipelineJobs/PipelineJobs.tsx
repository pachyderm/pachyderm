import React from 'react';

import JobList from '@dash-frontend/components/JobList';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import styles from './PipelineJobs.module.css';

const emptyStateTitle = "Let's Start :)";
const emptyJobListMessage = 'Create your first job on this pipeline!';

const PipelineJobs = () => {
  const {projectId, pipelineId} = useUrlState();

  return (
    <div className={styles.base}>
      <JobList
        projectId={projectId}
        pipelineId={pipelineId}
        emptyStateTitle={emptyStateTitle}
        emptyStateMessage={emptyJobListMessage}
      />
    </div>
  );
};

export default PipelineJobs;

import React from 'react';

import JobList from '@dash-frontend/components/JobList';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import styles from './PipelineJobs.module.css';

const PipelineJobs = () => {
  const {projectId, pipelineId} = useUrlState();

  return (
    <div className={styles.base}>
      <JobList projectId={projectId} pipelineId={pipelineId} />
    </div>
  );
};

export default PipelineJobs;

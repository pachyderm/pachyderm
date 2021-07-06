import {FileDocSVG, Link} from '@pachyderm/components';
import React from 'react';
import {Redirect} from 'react-router';

import {LETS_START_TITLE} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import JobList from '@dash-frontend/components/JobList';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  logsViewerPipelneRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';

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
      <div className={styles.readLogsWrapper}>
        <Link
          small
          to={logsViewerPipelneRoute({
            projectId,
            pipelineId: pipelineId,
          })}
        >
          <span className={styles.readLogsText}>
            Read Full Logs{' '}
            <FileDocSVG className={styles.readLogsSvg} width={20} height={24} />
          </span>
        </Link>
      </div>
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

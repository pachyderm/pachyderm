import {Link, DocumentSVG, Icon} from '@pachyderm/components';
import React from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  logsViewerJobRoute,
  logsViewerPipelneRoute,
} from '@dash-frontend/views/Project/utils/routes';

import styles from './ReadLogsButton.module.css';

const ReadLogsButton: React.FC = () => {
  const {projectId, pipelineId, jobId} = useUrlState();

  const logsLink = jobId
    ? logsViewerJobRoute({
        projectId,
        jobId: jobId,
        pipelineId: pipelineId,
      })
    : logsViewerPipelneRoute({
        projectId,
        pipelineId: pipelineId,
      });

  return (
    <Link small to={logsLink} className={styles.base}>
      <Icon color="plum" className={styles.iconSvg}>
        <DocumentSVG />
      </Icon>
      Read Logs
    </Link>
  );
};

export default ReadLogsButton;

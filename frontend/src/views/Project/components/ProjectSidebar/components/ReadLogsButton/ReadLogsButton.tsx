import {Link, DocumentSVG, Icon, Button} from '@pachyderm/components';
import React from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  logsViewerJobRoute,
  logsViewerPipelneRoute,
} from '@dash-frontend/views/Project/utils/routes';

import styles from './ReadLogsButton.module.css';

interface ReadLogsButtonProps {
  isButton?: boolean;
}

const ReadLogsButton: React.FC<ReadLogsButtonProps> = ({isButton = false}) => {
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

  if (isButton) {
    return (
      <Button
        autoWidth
        className={styles.button}
        buttonType={'secondary'}
        to={logsLink}
      >
        Read Log
      </Button>
    );
  }

  return (
    <Link to={logsLink} className={styles.base}>
      <Icon color="plum" className={styles.iconSvg}>
        <DocumentSVG />
      </Icon>
      Read Logs
    </Link>
  );
};

export default ReadLogsButton;

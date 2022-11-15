import React from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  logsViewerJobRoute,
  logsViewerPipelneRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {DocumentSVG, Button} from '@pachyderm/components';

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

  return (
    <Button
      buttonType={isButton ? 'secondary' : 'ghost'}
      to={logsLink}
      IconSVG={!isButton ? DocumentSVG : undefined}
    >
      Read Logs
    </Button>
  );
};

export default ReadLogsButton;

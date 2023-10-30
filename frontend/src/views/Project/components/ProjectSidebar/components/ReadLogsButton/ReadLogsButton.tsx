import React from 'react';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {useJob} from '@dash-frontend/hooks/useJob';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Button, PipelineSVG} from '@pachyderm/components';

interface ReadLogsButtonProps {
  omitIcon?: boolean;
}

const ReadLogsButton: React.FC<ReadLogsButtonProps> = ({omitIcon = false}) => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const {searchParams} = useUrlQueryState();
  const {isServiceOrSpout} = useCurrentPipeline();
  const {getPathToJobLogs, getPathToLatestJobLogs} = useLogsNavigation();

  const buttonText = !isServiceOrSpout ? 'Inspect Jobs' : 'Read Logs';

  const {job} = useJob(
    {
      projectId,
      pipelineName: pipelineId,
      id: searchParams.globalIdFilter,
    },
    {
      skip: !pipelineId,
    },
  );

  let logsLink = '';

  if (isServiceOrSpout) {
    logsLink = getPathToLatestJobLogs({
      projectId,
      pipelineId: pipelineId,
    });
  } else if (jobId || job?.id) {
    logsLink = getPathToJobLogs({
      projectId,
      jobId: jobId || job?.id || '',
      pipelineId: pipelineId,
    });
  }

  return (
    <Button
      buttonType="secondary"
      disabled={!logsLink}
      IconSVG={!omitIcon ? PipelineSVG : undefined}
      to={logsLink}
      name={buttonText}
    >
      {buttonText}
    </Button>
  );
};

export default ReadLogsButton;

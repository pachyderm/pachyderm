import React from 'react';

import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {useJobOrJobs} from '@dash-frontend/hooks/useJobOrJobs';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Button, NavigationHistorySVG} from '@pachyderm/components';

interface ReadLogsButtonProps {
  omitIcon?: boolean;
}

const ReadLogsButton: React.FC<ReadLogsButtonProps> = ({omitIcon = false}) => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const {searchParams} = useUrlQueryState();
  const {isServiceOrSpout} = useCurrentPipeline();
  const {getPathToJobLogs, getPathToLatestJobLogs} = useLogsNavigation();

  const buttonText = !isServiceOrSpout ? 'Previous Subjobs' : 'Read Logs';

  const {job} = useJobOrJobs(
    {
      projectId,
      pipelineName: pipelineId,
      id: searchParams.globalIdFilter,
    },
    !!pipelineId,
  );

  let logsLink = '';

  if (isServiceOrSpout) {
    logsLink = getPathToLatestJobLogs({
      projectId,
      pipelineId: pipelineId,
    });
  } else if (jobId || job?.job?.id) {
    logsLink = getPathToJobLogs({
      projectId,
      jobId: jobId || job?.job?.id || '',
      pipelineId: pipelineId,
    });
  }

  return (
    <Button
      buttonType="primary"
      disabled={!logsLink}
      IconSVG={!omitIcon ? NavigationHistorySVG : undefined}
      to={logsLink}
      name={buttonText}
    >
      {buttonText}
    </Button>
  );
};

export default ReadLogsButton;

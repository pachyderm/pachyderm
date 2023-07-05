import React, {useCallback} from 'react';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  logsViewerJobRoute,
  logsViewerLatestRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Button, PipelineSVG} from '@pachyderm/components';

interface ReadLogsButtonProps {
  omitIcon?: boolean;
}

const ReadLogsButton: React.FC<ReadLogsButtonProps> = ({omitIcon = false}) => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const {getUpdatedSearchParams} = useUrlQueryState();
  const {isSpout} = useCurrentPipeline();

  const buttonText = !isSpout ? 'Inspect Jobs' : 'Read Logs';

  const {jobs} = useJobs(
    {
      projectId,
      pipelineId,
      limit: 1,
    },
    {
      skip: !pipelineId,
    },
  );

  let logsLink = '';

  if (isSpout) {
    logsLink = logsViewerLatestRoute(
      {
        projectId,
        pipelineId: pipelineId,
      },
      false,
    );
  } else if (jobId || jobs[0]?.id) {
    logsLink = logsViewerJobRoute(
      {
        projectId,
        jobId: jobId || jobs[0]?.id,
        pipelineId: pipelineId,
      },
      false,
    );
  }
  const addLogsQueryParams = useCallback(
    (path: string) => {
      return `${path}?${getUpdatedSearchParams({
        datumFilters: [],
      })}`;
    },
    [getUpdatedSearchParams],
  );

  return (
    <Button
      buttonType="secondary"
      disabled={!logsLink}
      IconSVG={!omitIcon ? PipelineSVG : undefined}
      to={addLogsQueryParams(logsLink)}
      name={buttonText}
    >
      {buttonText}
    </Button>
  );
};

export default ReadLogsButton;

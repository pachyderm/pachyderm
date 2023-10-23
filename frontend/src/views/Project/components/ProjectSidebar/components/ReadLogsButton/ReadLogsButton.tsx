import React, {useCallback} from 'react';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {useJob} from '@dash-frontend/hooks/useJob';
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
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();
  const {isServiceOrSpout} = useCurrentPipeline();

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
    logsLink = logsViewerLatestRoute(
      {
        projectId,
        pipelineId: pipelineId,
      },
      false,
    );
  } else if (jobId || job?.id) {
    logsLink = logsViewerJobRoute(
      {
        projectId,
        jobId: jobId || job?.id || '',
        pipelineId: pipelineId,
      },
      false,
    );
  }
  const addLogsQueryParams = useCallback(
    (path: string) => {
      return `${path}?${getUpdatedSearchParams(
        {
          datumFilters: [],
        },
        true,
      )}`;
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

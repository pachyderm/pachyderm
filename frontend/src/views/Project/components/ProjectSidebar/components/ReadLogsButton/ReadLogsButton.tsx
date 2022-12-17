import React, {useCallback} from 'react';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {logsViewerJobRoute} from '@dash-frontend/views/Project/utils/routes';
import {Button, PipelineSVG} from '@pachyderm/components';

interface ReadLogsButtonProps {
  omitIcon?: boolean;
  buttonText?: string;
}

const ReadLogsButton: React.FC<ReadLogsButtonProps> = ({
  omitIcon = false,
  buttonText = 'Read Logs',
}) => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const {getUpdatedSearchParams} = useUrlQueryState();

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
  if (jobId || jobs[0]?.id) {
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

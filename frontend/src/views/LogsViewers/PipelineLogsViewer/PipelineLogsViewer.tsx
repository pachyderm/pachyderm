import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import {PIPELINE_START_TIME_OPTION} from '../constants/logsViewersConstants';
import LogsViewer from '../LogsViewer';

const PipelineLogsViewer: React.FC = () => {
  const {projectId, pipelineId} = useUrlState();
  const browserHistory = useHistory();

  const {jobs, loading} = useJobs({
    limit: 1,
    pipelineId: pipelineId,
    projectId,
  });

  const onCloseCallback = useCallback(() => {
    setTimeout(
      () => browserHistory.push(pipelineRoute({projectId, pipelineId})),
      500,
    );
  }, [browserHistory, pipelineId, projectId]);

  const headerText = pipelineId;

  return (
    <LogsViewer
      headerText={headerText}
      startTime={jobs[0]?.createdAt}
      onCloseCallback={onCloseCallback}
      loading={loading}
      dropdownLabel={PIPELINE_START_TIME_OPTION}
    />
  );
};

export default PipelineLogsViewer;

import React, {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import {JOB_START_TIME_OPTION} from '../constants/logsViewersConstants';
import LogsViewer from '../LogsViewer';

const JobLogsViewer: React.FC = () => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const browserHistory = useHistory();

  const {job, loading} = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const onCloseCallback = useCallback(() => {
    setTimeout(
      () => browserHistory.push(jobRoute({projectId, jobId, pipelineId})),
      500,
    );
  }, [browserHistory, jobId, pipelineId, projectId]);

  const headerText = `${pipelineId}:${jobId}`;

  return (
    <LogsViewer
      headerText={headerText}
      startTime={job?.createdAt}
      onCloseCallback={onCloseCallback}
      loading={loading}
      dropdownLabel={JOB_START_TIME_OPTION}
    />
  );
};

export default JobLogsViewer;

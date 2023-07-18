import {Maybe} from '@graphqlTypes';
import {useEffect, useState} from 'react';
import {useForm} from 'react-hook-form';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useLogs from '@dash-frontend/hooks/useLogs';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const timeNowInSeconds = Math.floor(Date.now() / 1000);

export const defaultValues: {[key: string]: number} = {
  'Last 30 Minutes': timeNowInSeconds - 1800,
  'Last 24 Hours': timeNowInSeconds - 86400,
  'Last 3 Days': timeNowInSeconds - 86400 * 3,
};

export type LogsViewerFormValues = {
  selectedTime: string;
  displayRawLogs: boolean;
  highlightUserLogs: boolean;
};

const useLogsViewer = (startTime?: number | null) => {
  const {projectId, pipelineId, jobId, datumId} = useUrlState();
  const {isServiceOrSpout, pipelineType} = useCurrentPipeline();

  const [selectedLogsMap, setSelectedLogsMap] = useState<{
    [key: number]: boolean;
  }>({});

  const [rawLogs, handleRawLogsChange] = useLocalProjectSettings({
    projectId,
    key: 'raw_logs',
  });

  const formCtx = useForm<LogsViewerFormValues>({
    mode: 'onChange',
    defaultValues: {
      selectedTime: 'default',
      displayRawLogs: rawLogs,
      highlightUserLogs: false,
    },
  });

  const {getValues} = formCtx;
  const {selectedTime, displayRawLogs, highlightUserLogs} = getValues();

  const dropdownValues: {[key: string]: Maybe<number> | undefined} = {
    default: startTime,
    ...defaultValues,
  };

  const {logs, loading, error} = useLogs(
    {
      projectId: projectId,
      pipelineName: pipelineId,
      jobId: jobId,
      datumId: datumId,
      start: dropdownValues[selectedTime],
      master: isServiceOrSpout,
    },
    {skip: !pipelineType || !dropdownValues[selectedTime]},
  );

  useEffect(() => {
    setSelectedLogsMap({});
  }, [selectedTime, rawLogs, datumId, jobId]);

  useEffect(() => {
    handleRawLogsChange(displayRawLogs);
  }, [displayRawLogs, handleRawLogsChange]);

  return {
    loading,
    logs,
    selectedLogsMap,
    setSelectedLogsMap,
    highlightUserLogs,
    rawLogs: displayRawLogs,
    error,
    formCtx,
  };
};
export default useLogsViewer;

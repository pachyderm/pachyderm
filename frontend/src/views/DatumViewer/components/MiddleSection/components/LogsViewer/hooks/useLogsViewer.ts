import {LogInputCursor, Maybe} from '@graphqlTypes';
import {useEffect, useState} from 'react';
import {useForm} from 'react-hook-form';

import {LOGS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useLogs from '@dash-frontend/hooks/useLogs';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import {LOGS_PAGE_SIZE} from '../constants/logsViewersConstants';

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

const useLogsViewer = (isSkippedDatum: boolean, startTime?: number | null) => {
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
  const [shouldPoll, setShouldPoll] = useState(true);
  const [page, setPage] = useState(1);
  const [cursors, setCursors] = useState<(LogInputCursor | null)[]>([null]);
  const [currentCursor, setCurrentCursor] = useState<LogInputCursor | null>(
    null,
  );

  const {logs, cursor, loading, error, refetch} = useLogs({
    variables: {
      args: {
        projectId: projectId,
        pipelineName: pipelineId,
        jobId: jobId,
        datumId: datumId,
        start: dropdownValues[selectedTime],
        master: isServiceOrSpout,
        limit: LOGS_PAGE_SIZE,
        cursor: currentCursor,
      },
    },
    skip: isSkippedDatum || !pipelineType || !dropdownValues[selectedTime],
    notifyOnNetworkStatusChange: true,
    pollInterval: shouldPoll ? LOGS_POLL_INTERVAL_MS : 0,
  });

  useEffect(() => {
    if (shouldPoll && logs.length !== 0) {
      setShouldPoll(false);
    }
  }, [logs.length, shouldPoll]);

  useEffect(() => {
    if (cursor && cursors.length < page) {
      setCurrentCursor({
        timestamp: {
          seconds: cursor.timestamp.seconds,
          nanos: cursor.timestamp.nanos,
        },
        message: cursor.message,
      });
      setCursors((arr) => [
        ...arr,
        {
          timestamp: {
            seconds: cursor.timestamp.seconds,
            nanos: cursor.timestamp.nanos,
          },
          message: cursor.message,
        },
      ]);
    } else {
      setCurrentCursor(cursors[page - 1]);
    }
  }, [cursor, cursors, page]);

  useEffect(() => {
    setPage(1);
    setCursors([null]);
    setCurrentCursor(null);
  }, [jobId, datumId, projectId]);

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
    refetch,
    page,
    setPage,
  };
};
export default useLogsViewer;

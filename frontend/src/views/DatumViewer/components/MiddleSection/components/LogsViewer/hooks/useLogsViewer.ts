import {useEffect, useState} from 'react';
import {useForm} from 'react-hook-form';

import {LogMessage} from '@dash-frontend/api/pps';
import {LOGS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {useLogs} from '@dash-frontend/hooks/useLogs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getISOStringFromUnix} from '@dash-frontend/lib/dateTime';

const timeNowInSeconds = Math.floor(Date.now() / 1000);

export const defaultValues: {[key: string]: string} = {
  'Last 30 Minutes': getISOStringFromUnix(timeNowInSeconds - 1800) || '',
  'Last 24 Hours': getISOStringFromUnix(timeNowInSeconds - 86400) || '',
  'Last 3 Days': getISOStringFromUnix(timeNowInSeconds - 86400 * 3) || '',
};

export type LogsViewerFormValues = {
  selectedTime: string;
  displayRawLogs: boolean;
  highlightUserLogs: boolean;
};

const useLogsViewer = (
  isSkippedDatum: boolean,
  startTime?: string,
  jobId?: string,
) => {
  const {projectId, pipelineId, datumId} = useUrlState();
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

  const dropdownValues: {[key: string]: string} = {
    default: startTime || '', // TODO: Can we just use null and not pass anything to since? Is this an optimization?
    ...defaultValues,
  };

  const [shouldPoll, setShouldPoll] = useState(true);
  const [page, setPage] = useState(1);
  const [cursors, setCursors] = useState<(LogMessage | undefined)[]>([
    undefined,
  ]);
  const [currentCursor, setCurrentCursor] = useState<LogMessage | undefined>(
    undefined,
  );

  const {logs, loading, error, refetch, isPagingError, isOverLokiQueryLimit} =
    useLogs(
      {
        pipeline: {
          name: pipelineId,
          project: {
            name: projectId,
          },
        },
        job: !isServiceOrSpout
          ? {
              id: jobId,
              pipeline: {
                name: pipelineId,
                project: {
                  name: projectId,
                },
              },
            }
          : undefined,
        datum: {
          id: datumId,
        },
        master: isServiceOrSpout,
      },
      {
        since: dropdownValues[selectedTime],
        cursor: currentCursor,
        enabled:
          !isSkippedDatum && !!pipelineType && !!dropdownValues[selectedTime], // TODO: I think we do only !isSkippedDatum
        refetchInterval: shouldPoll ? LOGS_POLL_INTERVAL_MS : false,
      },
    );

  useEffect(() => {
    if (shouldPoll && logs?.length && logs.length !== 0) {
      setShouldPoll(false);
    }
  }, [logs?.length, shouldPoll]);

  const cursor = logs?.at(-1);

  useEffect(() => {
    if (cursor && cursors.length < page) {
      setCurrentCursor(cursor);
      setCursors((arr) => [...arr, cursor]);
    } else {
      setCurrentCursor(cursors[page - 1]);
    }
  }, [cursor, cursors, page]);

  useEffect(() => {
    setPage(1);
    setCursors([undefined]);
    setCurrentCursor(undefined);
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
    isPagingError,
    isOverLokiQueryLimit,
  };
};
export default useLogsViewer;

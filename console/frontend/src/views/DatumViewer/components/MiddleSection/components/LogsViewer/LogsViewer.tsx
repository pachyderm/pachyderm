import React from 'react';

import {LogMessage} from '@dash-frontend/api/pps';

import LogsBody from './components/LogsBody';
import LogsListHeader from './components/LogsListHeader';
import styles from './LogsViewer.module.css';

export type LogsViewerProps = {
  loading: boolean;
  logs?: LogMessage[];
  highlightUserLogs: boolean;
  selectedLogsMap: {[key: number]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: number]: boolean}>
  >;
  rawLogs: boolean;
  error?: string;
  isPagingError: boolean;
  isSkippedDatum?: boolean;
  page: number;
  isOverLokiQueryLimit?: boolean;
};

const LogsViewer: React.FC<LogsViewerProps> = ({
  loading,
  logs,
  highlightUserLogs,
  selectedLogsMap,
  setSelectedLogsMap,
  rawLogs,
  error,
  isSkippedDatum,
  page,
  isPagingError,
  isOverLokiQueryLimit,
}) => {
  return (
    <div className={styles.base}>
      <LogsListHeader
        selectedLogsMap={selectedLogsMap}
        setSelectedLogsMap={setSelectedLogsMap}
        rawLogs={rawLogs}
        logs={logs}
        loading={loading}
      />

      <LogsBody
        highlightUserLogs={highlightUserLogs}
        rawLogs={rawLogs}
        loading={loading}
        logs={logs}
        selectedLogsMap={selectedLogsMap}
        setSelectedLogsMap={setSelectedLogsMap}
        error={error}
        isSkippedDatum={isSkippedDatum}
        page={page}
        isPagingError={isPagingError}
        isOverLokiQueryLimit={isOverLokiQueryLimit}
      />
    </div>
  );
};

export default LogsViewer;

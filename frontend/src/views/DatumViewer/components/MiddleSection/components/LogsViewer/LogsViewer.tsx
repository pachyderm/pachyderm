import {ApolloError} from '@apollo/client';
import {GetLogsQuery} from '@graphqlTypes';
import React from 'react';

import LogsBody from './components/LogsBody';
import LogsListHeader from './components/LogsListHeader';
import styles from './LogsViewer.module.css';

export type LogsViewerProps = {
  loading: boolean;
  logs: GetLogsQuery['logs']['items'];
  highlightUserLogs: boolean;
  selectedLogsMap: {[key: number]: boolean};
  setSelectedLogsMap: React.Dispatch<
    React.SetStateAction<{[key: number]: boolean}>
  >;
  rawLogs: boolean;
  error?: ApolloError;
  isSkippedDatum?: boolean;
  page: number;
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
      />
    </div>
  );
};

export default LogsViewer;

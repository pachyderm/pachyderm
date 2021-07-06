import {FullPageModal} from '@pachyderm/components';
import React from 'react';

import LogsBody from './components/LogsBody';
import LogsListHeader from './components/LogsListHeader';
import LogsModalHeader from './components/LogsModalHeader';
import useLogsViewer from './hooks/useLogsViewer';
import styles from './LogsViewer.module.css';

export type LogsViewerProps = {
  headerText: string;
  startTime?: number;
  onCloseCallback: () => void;
  loading: boolean;
  dropdownLabel: string;
};

const LogsViewer: React.FC<LogsViewerProps> = ({
  headerText,
  startTime,
  onCloseCallback,
  loading,
  dropdownLabel,
}) => {
  const {
    onClose,
    isOpen,
    loading: loadingLogs,
    logs,
    highlightUserLogs,
    setHighlightUserLogs,
    selectedLogsMap,
    setSelectedLogsMap,
    selectedTime,
    dropdownOptions,
    setSelectedTime,
  } = useLogsViewer(onCloseCallback, dropdownLabel, startTime);

  return (
    <FullPageModal show={isOpen} onHide={onClose} hideType="exit">
      <div className={styles.base}>
        <LogsModalHeader
          setHighlightUserLogs={setHighlightUserLogs}
          selectedLogsMap={selectedLogsMap}
          logs={logs}
          headerText={headerText}
        />

        <LogsListHeader
          dropdownOptions={dropdownOptions}
          selectedTime={selectedTime}
          setSelectedTime={setSelectedTime}
          selectedLogsMap={selectedLogsMap}
          setSelectedLogsMap={setSelectedLogsMap}
          logs={logs}
          loading={loading || loadingLogs}
        />

        <LogsBody
          highlightUserLogs={highlightUserLogs}
          loading={loading || loadingLogs}
          logs={logs}
          selectedLogsMap={selectedLogsMap}
          setSelectedLogsMap={setSelectedLogsMap}
        />
      </div>
    </FullPageModal>
  );
};

export default LogsViewer;

import React from 'react';

import {CaptionText, FullPagePanelModal} from '@pachyderm/components';

import DatumHeaderBreadcrumbs from './components/DatumHeaderBreadcrumbs';
import LogsViewer from './components/LogsViewer';
import LogsControls from './components/LogsViewer/components/LogsControls';
import LogsFooter from './components/LogsViewer/components/LogsFooter';
import useLogsViewer from './components/LogsViewer/hooks/useLogsViewer';
import useMiddleSection from './hooks/useMiddleSection';
import styles from './MiddleSection.module.css';

const MiddleSection = () => {
  const {
    headerText,
    headerValue,
    job,
    jobId,
    startTime,
    loading: loadingData,
    isSkippedDatum,
    isSpout,
  } = useMiddleSection();

  const {
    loading: loadingLogs,
    logs,
    highlightUserLogs,
    selectedLogsMap,
    setSelectedLogsMap,
    rawLogs,
    error,
    formCtx,
    refetch,
    page,
    setPage,
  } = useLogsViewer(startTime);

  return (
    <FullPagePanelModal.Body>
      <div className={styles.base}>
        {!isSpout && <DatumHeaderBreadcrumbs jobId={jobId || job?.id} />}
        <div className={styles.header} data-testid="MiddleSection__title">
          <h6>{headerText}</h6>
          <CaptionText color="black" className={styles.headerId}>
            {headerValue}
          </CaptionText>
        </div>

        <div className={styles.tabs}>
          <h6>Logs</h6>

          <LogsControls
            selectedLogsMap={selectedLogsMap}
            logs={logs}
            formCtx={formCtx}
          />
        </div>
        <LogsViewer
          loading={loadingData || loadingLogs}
          logs={logs}
          highlightUserLogs={highlightUserLogs}
          selectedLogsMap={selectedLogsMap}
          setSelectedLogsMap={setSelectedLogsMap}
          rawLogs={rawLogs}
          error={error}
          isSkippedDatum={isSkippedDatum}
          page={page}
        />
        {!isSkippedDatum && (
          <LogsFooter
            logs={logs}
            refetch={refetch}
            page={page}
            setPage={setPage}
          />
        )}
      </div>
    </FullPagePanelModal.Body>
  );
};

export default MiddleSection;

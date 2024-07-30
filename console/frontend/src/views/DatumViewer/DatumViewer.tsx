import React from 'react';

import InfoPanel from '@dash-frontend/components/InfoPanel/';
import PipelineActionsMenu from '@dash-frontend/components/PipelineActionsMenu';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import PipelineInfo from '@dash-frontend/views/Project/components/ProjectSidebar/components/PipelineDetails/components/PipelineInfo';
import {FullPageResizablePanelModal, Group} from '@pachyderm/components';

import {pipelineRoute, useSelectedRunRoute} from '../Project/utils/routes';

import DatumDetails from './components/DatumDetails';
import LeftPanel from './components/LeftPanel';
import MiddleSection from './components/MiddleSection';
import styles from './DatumViewer.module.css';
import useDatumViewer from './hooks/useDatumViewer';

type DatumViewerProps = {
  onCloseRoute: string;
};
const DatumViewer: React.FC<DatumViewerProps> = ({onCloseRoute}) => {
  const {
    isOpen,
    onClose,
    pipelineId,
    jobId,
    datumId,
    job,
    isServiceOrSpout,
    leftPanelOpen,
    setLeftPanelOpen,
  } = useDatumViewer(onCloseRoute);

  const rightPanelProps = {
    defaultSize: 25,
    minSize: 25,
    showClose: true,
    onClose,
    headerContent: (
      <Group spacing={8} className={styles.actionsGroup}>
        <PipelineActionsMenu pipelineId={pipelineId} />
      </Group>
    ),
    'data-testid': 'SidePanel__right',
  };

  return (
    <>
      {isOpen && (
        <FullPageResizablePanelModal
          show={isOpen}
          onHide={onClose}
          autoSaveId="DatumViewer"
        >
          {isServiceOrSpout && (
            <>
              <FullPageResizablePanelModal.Body defaultSize={50} minSize={40}>
                <MiddleSection key={jobId} />
              </FullPageResizablePanelModal.Body>
              <FullPageResizablePanelModal.PanelResizeHandle />
              <FullPageResizablePanelModal.Panel {...rightPanelProps}>
                <PipelineInfo />
              </FullPageResizablePanelModal.Panel>
            </>
          )}

          {!isServiceOrSpout && (
            <>
              <LeftPanel
                job={job}
                setIsOpen={setLeftPanelOpen}
                isOpen={leftPanelOpen}
              />
              <FullPageResizablePanelModal.Body
                defaultSize={50}
                minSize={40}
                className={
                  leftPanelOpen
                    ? styles.middleSectionLeftPanelOpen
                    : styles.middleSection
                }
              >
                <MiddleSection key={`${jobId}-${datumId}`} />
              </FullPageResizablePanelModal.Body>
              <FullPageResizablePanelModal.PanelResizeHandle />
              <FullPageResizablePanelModal.Panel {...rightPanelProps}>
                {datumId ? (
                  <DatumDetails className={styles.overflowYScroll} />
                ) : (
                  <InfoPanel
                    hideReadLogs={true}
                    className={styles.overflowYScroll}
                  />
                )}
              </FullPageResizablePanelModal.Panel>
            </>
          )}
        </FullPageResizablePanelModal>
      )}
    </>
  );
};

export const PipelineDatumViewer: React.FC = () => {
  const {projectId, pipelineId} = useUrlState();
  const {getPathFromLogs} = useLogsNavigation();
  const onCloseRoute = getPathFromLogs(
    pipelineRoute({projectId, pipelineId}, false),
  );
  return <DatumViewer onCloseRoute={onCloseRoute} />;
};

export const JobDatumViewer: React.FC = () => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const {getPathFromLogs} = useLogsNavigation();
  const jobRoute = useSelectedRunRoute({
    projectId,
    jobId: jobId,
    pipelineId: pipelineId,
  });
  const onCloseRoute = getPathFromLogs(jobRoute);
  return <DatumViewer onCloseRoute={onCloseRoute} />;
};

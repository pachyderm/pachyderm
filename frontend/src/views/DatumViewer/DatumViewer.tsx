import {PipelineType} from '@graphqlTypes';
import React from 'react';

import InfoPanel from '@dash-frontend/components/InfoPanel/';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import PipelineInfo from '@dash-frontend/views/Project/components/ProjectSidebar/components/PipelineDetails/components/PipelineInfo';
import {FullPagePanelModal} from '@pachyderm/components';

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
  const {isOpen, onClose, jobId, datumId, job, isServiceOrSpout} =
    useDatumViewer(onCloseRoute);
  return (
    <>
      {isOpen && (
        <FullPagePanelModal
          show={isOpen}
          onHide={onClose}
          hideType="exit"
          hideLeftPanel={isServiceOrSpout}
        >
          {isServiceOrSpout && (
            <>
              <MiddleSection key={jobId} />
              <FullPagePanelModal.RightPanel>
                <PipelineInfo />
              </FullPagePanelModal.RightPanel>
            </>
          )}

          {!isServiceOrSpout && (
            <>
              <LeftPanel job={job} />
              <MiddleSection key={`${jobId}-${datumId}`} />
              <FullPagePanelModal.RightPanel>
                {datumId ? (
                  <DatumDetails className={styles.overflowYScroll} />
                ) : (
                  <InfoPanel
                    hideReadLogs={true}
                    className={styles.overflowYScroll}
                  />
                )}
              </FullPagePanelModal.RightPanel>
            </>
          )}
        </FullPagePanelModal>
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

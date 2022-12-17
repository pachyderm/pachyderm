import React from 'react';

import InfoPanel from '@dash-frontend/components/InfoPanel/';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {FullPagePanelModal} from '@pachyderm/components';

import {pipelineRoute, jobRoute} from '../Project/utils/routes';

import DatumDetails from './components/DatumDetails';
import LeftPanel from './components/LeftPanel';
import MiddleSection from './components/MiddleSection';
import styles from './DatumViewer.module.css';
import useDatumViewer from './hooks/useDatumViewer';
type DatumViewerProps = {
  onCloseRoute: string;
};
const DatumViewer: React.FC<DatumViewerProps> = ({onCloseRoute}) => {
  const {isOpen, onClose, datumId} = useDatumViewer(onCloseRoute);

  return (
    <FullPagePanelModal show={isOpen} onHide={onClose} hideType="exit">
      <LeftPanel />
      <MiddleSection />
      <FullPagePanelModal.RightPanel>
        {datumId ? (
          <DatumDetails className={styles.overflowYScroll} />
        ) : (
          <InfoPanel className={styles.overflowYScroll} disableGlobalFilter />
        )}
      </FullPagePanelModal.RightPanel>
    </FullPagePanelModal>
  );
};

export const PipelineDatumViewer: React.FC = () => {
  const {projectId, pipelineId} = useUrlState();
  const onCloseRoute = pipelineRoute({projectId, pipelineId});
  return <DatumViewer onCloseRoute={onCloseRoute} />;
};

export const JobDatumViewer: React.FC = () => {
  const {projectId, pipelineId, jobId} = useUrlState();
  const onCloseRoute = jobRoute({projectId, pipelineId, jobId});
  return <DatumViewer onCloseRoute={onCloseRoute} />;
};

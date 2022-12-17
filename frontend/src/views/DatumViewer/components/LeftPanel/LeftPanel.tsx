import React from 'react';
import {Route} from 'react-router-dom';

import {
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {FullPagePanelModal} from '@pachyderm/components';

import {BreadCrumbs} from './components/BreadCrumbs';
import DatumList from './components/DatumList';
import {Filter} from './components/Filter';
import JobList from './components/JobList';
import useLeftPanel from './hooks/useLeftPanel';
import styles from './LeftPanel.module.css';

const LeftPanel: React.FC = () => {
  const {isExpanded, setIsExpanded, jobs, loading, formCtx} = useLeftPanel();
  return (
    <FullPagePanelModal.LeftPanel
      isExpanded={isExpanded}
      setIsExpanded={setIsExpanded}
    >
      <div className={styles.base}>
        <Filter formCtx={formCtx} />
        <BreadCrumbs isExpanded={isExpanded} setIsExpanded={setIsExpanded} />
        <div className={styles.data}>
          <Route
            path={[
              LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
              PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
              LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
              PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
            ]}
            exact
          >
            <JobList
              jobs={jobs}
              loading={loading}
              isExpanded={isExpanded}
              setIsExpanded={setIsExpanded}
            />

            {isExpanded && <DatumList setIsExpanded={setIsExpanded} />}
          </Route>

          <Route
            path={[
              PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
              LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
              PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
              LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
            ]}
          >
            <DatumList setIsExpanded={setIsExpanded} />
          </Route>
        </div>
      </div>
    </FullPagePanelModal.LeftPanel>
  );
};

export default LeftPanel;

import React from 'react';
import {Route} from 'react-router-dom';

import {JobInfo} from '@dash-frontend/api/pps';
import {
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {FullPagePanelModal} from '@pachyderm/components';

import {BreadCrumbs} from './components/BreadCrumbs';
import DatumList from './components/DatumList';
import {Filter} from './components/Filter';
import JobList from './components/JobList';
import useLeftPanel from './hooks/useLeftPanel';
import styles from './LeftPanel.module.css';

type LeftPanelProps = {
  job?: JobInfo;
};

const LeftPanel: React.FC<LeftPanelProps> = ({job}) => {
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
              LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
              PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
            ]}
            exact
          >
            <JobList
              currentJob={job}
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

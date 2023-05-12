import classNames from 'classnames';
import React from 'react';
import {Route} from 'react-router-dom';

import useDatumPath from '@dash-frontend/views/DatumViewer/hooks/useDatumPath';
import {
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {ArrowLeftSVG, Button, CaptionTextSmall} from '@pachyderm/components';

import styles from './BreadCrumbs.module.css';

export interface BreadCrumbsProps {
  isExpanded: boolean;
  setIsExpanded: React.Dispatch<React.SetStateAction<boolean>>;
}

export const BreadCrumbs: React.FC<BreadCrumbsProps> = ({
  isExpanded,
  setIsExpanded,
}) => {
  const {
    urlJobId: currentJobId,
    currentDatumId,
    updateSelectedJob,
  } = useDatumPath();

  const back = () => {
    updateSelectedJob(currentJobId);
    setIsExpanded(false);
  };

  if (!currentJobId && !currentDatumId) {
    return (
      <div className={styles.base}>
        <CaptionTextSmall>Jobs</CaptionTextSmall>
      </div>
    );
  }

  return (
    <div className={styles.base}>
      <div className={styles.crumbWrapper} data-testid="BreadCrumbs__base">
        {currentJobId && (
          <CaptionTextSmall
            className={classNames(styles.id, {
              [styles.expanded]: isExpanded,
            })}
          >
            {!currentDatumId || isExpanded ? `Job: ${currentJobId}` : '...'}
          </CaptionTextSmall>
        )}

        {currentDatumId && (
          <>
            <CaptionTextSmall className={styles.divider}>/</CaptionTextSmall>
            <CaptionTextSmall
              className={classNames(styles.id, {
                [styles.expanded]: isExpanded,
              })}
            >
              Datum: {currentDatumId}
            </CaptionTextSmall>
          </>
        )}
      </div>
      <Route
        path={[
          PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
          LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
          PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
          LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
        ]}
      >
        <Button
          buttonType="ghost"
          IconSVG={ArrowLeftSVG}
          iconPosition="end"
          color="black"
          onClick={back}
        >
          Back
        </Button>
      </Route>
    </div>
  );
};
export default BreadCrumbs;

import classnames from 'classnames';
import React from 'react';

import GlobalIdCopy from '@dash-frontend/components/GlobalIdCopy';
import JobStateBadges from '@dash-frontend/components/JobStateBadges';
import {ArrowRightSVG, Button, CloseSVG} from '@pachyderm/components';

import styles from './GlobalIDStatusBar.module.css';
import {useGlobalIDStatusBar} from './hooks/useGlobalIDStatusBar';

const GlobalIDStatusBar: React.FC = () => {
  const {
    globalIdFilter,
    getPathToJobsTable,
    internalJobSet,
    totalRuntime,
    totalSubJobs,
    closeStatusBar,
    showErrorBorder,
  } = useGlobalIDStatusBar();

  if (!globalIdFilter) {
    return <></>;
  }
  return (
    <div className={styles.base}>
      <div
        className={classnames(styles.container, {
          [styles.error]: showErrorBorder,
        })}
        aria-label={
          showErrorBorder ? 'Global ID contains subjobs with errors' : undefined
        }
      >
        <div className={styles.leftGroup}>
          <div>
            Global ID - <GlobalIdCopy id={globalIdFilter} shortenId />
          </div>

          <div className={styles.center}>
            <Button
              to={() => getPathToJobsTable('subjobs')}
              buttonType="ghost"
              color="black"
            >
              {totalSubJobs}
              {internalJobSet && (
                <JobStateBadges jobSet={internalJobSet} showToolTips={false} />
              )}
            </Button>
          </div>

          <Button
            to={() => getPathToJobsTable('runtimes')}
            IconSVG={ArrowRightSVG}
            buttonType="ghost"
            iconPosition="end"
            color="black"
          >
            Runtime {totalRuntime}
          </Button>
        </div>
        <Button
          onClick={closeStatusBar}
          IconSVG={CloseSVG}
          buttonType="ghost"
          iconPosition="end"
          color="black"
        >
          Clear Filter
        </Button>
      </div>
    </div>
  );
};

export default GlobalIDStatusBar;

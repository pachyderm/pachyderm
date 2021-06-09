import {SkeletonDisplayText, Tooltip} from '@pachyderm/components';
import classnames from 'classnames';
import React, {useCallback, useState} from 'react';
import {Link} from 'react-router-dom';

import Badge from '@dash-frontend/components/Badge';
import Header from '@dash-frontend/components/Header';
import Search from '@dash-frontend/components/Search';

import {ReactComponent as BackArrowSvg} from './BackArrow.svg';
import useProjectHeader from './hooks/useProjectHeader';
import styles from './ProjectHeader.module.css';

const ProjectHeader = () => {
  const {projectName, numOfFailedJobs, seeJobsUrl, loading, seeJobsOpen} =
    useProjectHeader();
  const [showTooltip, setShowTooltip] = useState(false);
  const setProjectNameRef = useCallback((element: HTMLHeadingElement) => {
    setShowTooltip(element && element.clientWidth < element.scrollWidth);
  }, []);

  return (
    <Header>
      <Link to="/" className={styles.goBack}>
        <BackArrowSvg
          aria-label="Go back to landing page."
          className={styles.goBackSvg}
        />
      </Link>

      <div className={styles.projectNameWrapper}>
        {loading ? (
          <SkeletonDisplayText
            data-testid="ProjectHeader__projectNameLoader"
            className={styles.projectNameLoader}
            color="grey"
          />
        ) : (
          <Tooltip
            tooltipKey="Project Name"
            tooltipText={projectName}
            placement="bottom"
            size="large"
            disabled={!showTooltip}
            className={styles.projectNameTooltip}
          >
            <h1 ref={setProjectNameRef} className={styles.projectName}>
              {projectName}
            </h1>
          </Tooltip>
        )}
      </div>
      <Search />

      <Link
        className={classnames(styles.seeJobs, {
          [styles.active]: seeJobsOpen,
        })}
        to={seeJobsUrl}
      >
        <div className={styles.seeJobsContent}>
          {numOfFailedJobs > 0 && (
            <Badge
              className={styles.seeJobsBadge}
              aria-label="Number of failed jobs"
            >
              {numOfFailedJobs}
            </Badge>
          )}
          <span className={styles.seeJobsText}>Show Jobs</span>
        </div>
      </Link>
    </Header>
  );
};

export default ProjectHeader;

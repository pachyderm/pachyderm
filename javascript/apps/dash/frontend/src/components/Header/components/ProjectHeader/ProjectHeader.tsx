import {SkeletonDisplayText} from '@pachyderm/components';
import React from 'react';
import {Link} from 'react-router-dom';

import Badge from '@dash-frontend/components/Badge';

import {ReactComponent as BackArrowSvg} from './BackArrow.svg';
import useProjectHeader from './hooks/useProjectHeader';
import styles from './ProjectHeader.module.css';

const ProjectHeader = () => {
  const {
    projectName,
    numOfFailedJobs,
    seeJobsUrl,
    loading,
  } = useProjectHeader();

  return (
    <>
      <Link to="/" className={styles.goBack}>
        <BackArrowSvg
          aria-label={'Go back to landing page.'}
          className={styles.goBackSvg}
        />
      </Link>

      <h1 className={styles.projectName}>
        {loading ? (
          <SkeletonDisplayText
            data-testid="ProjectHeader__projectNameLoader"
            className={styles.projectNameLoader}
            blueShimmer
          />
        ) : (
          projectName
        )}
      </h1>

      <Link className={styles.seeJobs} to={seeJobsUrl}>
        <div className={styles.seeJobsContent}>
          {numOfFailedJobs > 0 && (
            <Badge
              className={styles.seeJobsBadge}
              aria-label="Number of failed jobs"
            >
              {numOfFailedJobs}
            </Badge>
          )}
          <span className={styles.seeJobsText}>See Jobs</span>
        </div>
      </Link>
    </>
  );
};

export default ProjectHeader;

import {SkeletonDisplayText} from '@pachyderm/components';
import React from 'react';
import {Link} from 'react-router-dom';

import Badge from '@dash-frontend/components/Badge';
import Header from '@dash-frontend/components/Header';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {projectRoute} from '@dash-frontend/views/Project/utils/routes';

import {ReactComponent as BackArrowSvg} from './BackArrow.svg';
import useProjectHeader from './hooks/useProjectHeader';
import styles from './ProjectHeader.module.css';

const ProjectHeader = ({totalDags}: {totalDags: number}) => {
  const {
    projectName,
    numOfFailedJobs,
    seeJobsUrl,
    loading,
  } = useProjectHeader();
  const {projectId, dagId} = useUrlState();
  const goesBackToProject = dagId && totalDags > 1;

  return (
    <Header>
      <Link
        to={
          goesBackToProject ? projectRoute({projectId, withSearch: false}) : '/'
        }
        className={styles.goBack}
      >
        <BackArrowSvg
          aria-label={
            goesBackToProject
              ? 'Go back to project page'
              : 'Go back to landing page.'
          }
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
    </Header>
  );
};

export default ProjectHeader;

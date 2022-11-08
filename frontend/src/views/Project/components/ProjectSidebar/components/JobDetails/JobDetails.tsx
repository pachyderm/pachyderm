import {JobSet} from '@graphqlTypes';
import React, {useState} from 'react';
import {Helmet} from 'react-helmet';
import {Redirect, Route} from 'react-router-dom';

import CommitIdCopy from '@dash-frontend/components/CommitIdCopy';
import {useJobSet} from '@dash-frontend/hooks/useJobSet';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  LINEAGE_PIPELINE_JOB_PATH,
  PROJECT_PIPELINE_JOB_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  jobsRoute,
  jobRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Link, LoadingDots} from '@pachyderm/components';

import InfoPanel from './components/InfoPanel';
import PipelineList from './components/PipelineList';
import styles from './JobDetails.module.css';

const JobDetails = () => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {jobSet, loading} = useJobSet({id: jobId, projectId});

  const [containerRef, setContainerRef] = useState<HTMLDivElement | null>(null);

  if (
    jobSet &&
    jobSet.jobs.length > 0 &&
    (!pipelineId || !jobSet.jobs.find((job) => job.pipelineName === pipelineId))
  ) {
    return (
      <Redirect
        to={jobRoute({
          projectId,
          jobId,
          pipelineId: jobSet.jobs[0].pipelineName,
        })}
      />
    );
  }

  return (
    <div className={styles.base} ref={setContainerRef}>
      <Helmet>
        <title>Job - Pachyderm Console</title>
      </Helmet>
      <section className={styles.headerSection}>
        <Route path={LINEAGE_PIPELINE_JOB_PATH}>
          <nav>
            <Link to={jobsRoute({projectId})} className={styles.seeMoreJobs}>
              All jobs
            </Link>
            {pipelineId ? '/ ' : ''}
            {pipelineId && (
              <Link
                className={styles.seeMoreJobs}
                to={pipelineRoute({
                  projectId,
                  pipelineId,
                  tabId: 'jobs',
                })}
              >
                Pipeline: {pipelineId}
              </Link>
            )}
          </nav>
        </Route>

        <div className={styles.heading}>
          <h5 className={styles.jobLabel}>Job</h5>
          <CommitIdCopy commit={jobId} longId />
        </div>
      </section>

      <section className={styles.pipelineSection}>
        {loading && (
          <div
            className={styles.pipelineList}
            data-testid="JobDetails__loading"
          >
            <LoadingDots />
          </div>
        )}

        {!loading && (
          <PipelineList
            containerRef={containerRef}
            jobSet={jobSet as JobSet}
            projectId={projectId}
            jobId={jobId}
          />
        )}

        <div className={styles.pipelineSpec}>
          <Route path={[PROJECT_PIPELINE_JOB_PATH, LINEAGE_PIPELINE_JOB_PATH]}>
            <InfoPanel showReadLogs />
          </Route>
        </div>
      </section>
    </div>
  );
};

export default JobDetails;

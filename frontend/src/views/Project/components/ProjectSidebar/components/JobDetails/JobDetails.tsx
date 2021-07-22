import {Link, Tooltip, LoadingDots} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {NavLink, Redirect, Route} from 'react-router-dom';

import {useJobSet} from '@dash-frontend/hooks/useJobSet';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import readableJobState from '@dash-frontend/lib/readableJobState';
import {PIPELINE_JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  jobsRoute,
  jobRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {JobState} from '@graphqlTypes';

import InfoPanel from './components/InfoPanel';
import styles from './JobDetails.module.css';

type JobVisualState = 'BUSY' | 'ERROR' | 'SUCCESS';

const getVisualJobState = (state: JobState): JobVisualState => {
  switch (state) {
    case JobState.JOB_CREATED:
    case JobState.JOB_EGRESSING:
    case JobState.JOB_RUNNING:
    case JobState.JOB_STARTING:
      return 'BUSY';
    case JobState.JOB_FAILURE:
    case JobState.JOB_KILLED:
      return 'ERROR';
    default:
      return 'SUCCESS';
  }
};

const getJobStateHref = (state: JobVisualState) => {
  switch (state) {
    case 'BUSY':
      return '/dag_busy.svg';
    case 'ERROR':
      return '/dag_error.svg';
    default:
      return '/dag_success.svg';
  }
};

const JobDetails = () => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {jobSet, loading} = useJobSet({id: jobId, projectId});

  if (!pipelineId && jobSet && jobSet.jobs.length > 0) {
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
    <div className={styles.base}>
      <section className={styles.headerSection}>
        <nav>
          <Link to={jobsRoute({projectId})} className={styles.seeMoreJobs}>
            All jobs{pipelineId ? ' >' : ''}
          </Link>
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

        <h2 className={styles.heading}>Job {jobId}</h2>
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
          <nav>
            <ol className={styles.pipelineList}>
              {jobSet?.jobs.map((job) => {
                const jobVisualState = getVisualJobState(job.state);

                return (
                  <li key={job.pipelineName}>
                    <Tooltip
                      tooltipText={job.pipelineName}
                      tooltipKey={job.pipelineName}
                      size="large"
                      placement="right"
                    >
                      <NavLink
                        aria-current="true"
                        activeClassName={styles.active}
                        exact={true}
                        to={jobRoute({
                          projectId,
                          jobId,
                          pipelineId: job.pipelineName,
                        })}
                        className={classnames(styles.pipelineLink, {
                          [styles.error]: jobVisualState === 'ERROR',
                          [styles.success]: jobVisualState === 'SUCCESS',
                          [styles.busy]: jobVisualState === 'BUSY',
                        })}
                      >
                        <img
                          alt={`Pipeline job ${job.id} ${readableJobState(
                            job.state,
                          )}:`}
                          className={styles.pipelineIcon}
                          src={getJobStateHref(jobVisualState)}
                        />
                        <div className={styles.pipelineLinkText}>
                          {job.pipelineName}
                        </div>
                      </NavLink>
                    </Tooltip>
                  </li>
                );
              })}
            </ol>
          </nav>
        )}

        <div className={styles.pipelineSpec}>
          <Route path={PIPELINE_JOB_PATH}>
            <InfoPanel />
          </Route>
        </div>
      </section>
    </div>
  );
};

export default JobDetails;

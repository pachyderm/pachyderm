import {JobState} from '@graphqlTypes';
import {Link, Tooltip, LoadingDots, Group} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Helmet} from 'react-helmet';
import {NavLink, Redirect, Route} from 'react-router-dom';

import CommitIdCopy from '@dash-frontend/components/CommitIdCopy';
import {useJobSet} from '@dash-frontend/hooks/useJobSet';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import readableJobState from '@dash-frontend/lib/readableJobState';
import {
  LINEAGE_PIPELINE_JOB_PATH,
  PROJECT_PIPELINE_JOB_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  jobsRoute,
  jobRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';

import InfoPanel from './components/InfoPanel';
import styles from './JobDetails.module.css';

type JobVisualState = 'BUSY' | 'ERROR' | 'SUCCESS';

const getVisualJobState = (state: JobState): JobVisualState => {
  switch (state) {
    case JobState.JOB_CREATED:
    case JobState.JOB_EGRESSING:
    case JobState.JOB_RUNNING:
    case JobState.JOB_STARTING:
    case JobState.JOB_FINISHING:
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
    <div className={styles.base}>
      <Helmet>
        <title>Job - Pachyderm Console</title>
      </Helmet>
      <section className={styles.headerSection}>
        <Route path={LINEAGE_PIPELINE_JOB_PATH}>
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
        </Route>

        <Group spacing={8} className={styles.heading}>
          Job
          <CommitIdCopy commit={jobId} longId />
        </Group>
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
          <Route path={[PROJECT_PIPELINE_JOB_PATH, LINEAGE_PIPELINE_JOB_PATH]}>
            <InfoPanel showReadLogs />
          </Route>
        </div>
      </section>
    </div>
  );
};

export default JobDetails;

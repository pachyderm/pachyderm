import {Link, ArrowSVG, Tooltip} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {NavLink, Redirect, Route} from 'react-router-dom';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {PIPELINE_JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  pipelineJobRoute,
  jobsRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {PipelineJobState} from '@graphqlTypes';

import InfoPanel from './components/InfoPanel';
import styles from './JobDetails.module.css';
import pipelineJobs from './mock/pipelineJobs';

const JobDetails = () => {
  const {jobId, projectId, pipelineJobId} = useUrlState();

  // Replace pipelineJob mocks with data from useJob (which should contain a jobSet)

  if (!pipelineJobId) {
    return (
      <Redirect
        to={pipelineJobRoute({
          projectId,
          jobId,
          pipelineJobId: pipelineJobs[0].id,
          pipelineId: pipelineJobs[0].pipelineName,
        })}
      />
    );
  }

  return (
    <div className={styles.base}>
      <section className={styles.headerSection}>
        <nav>
          <Link to={jobsRoute({projectId})} className={styles.seeMoreJobs}>
            <ArrowSVG aria-hidden className={styles.backArrow} /> See more jobs
          </Link>
        </nav>

        <h2 className={styles.heading}>Job {jobId}</h2>
      </section>

      <section className={styles.pipelineSection}>
        <nav>
          <ol className={styles.pipelineList}>
            {pipelineJobs.map((job) => (
              <li key={job.id}>
                <Tooltip
                  tooltipText={job.pipelineName}
                  tooltipKey={job.id}
                  size="large"
                  placement="right"
                >
                  <NavLink
                    aria-current="true"
                    activeClassName={styles.active}
                    exact={true}
                    to={pipelineJobRoute({
                      projectId,
                      jobId,
                      pipelineJobId: job.id,
                      pipelineId: job.pipelineName,
                    })}
                    className={classnames(styles.pipelineLink, {
                      [styles.error]:
                        job.state === PipelineJobState.JOB_FAILURE,
                      [styles.success]:
                        job.state === PipelineJobState.JOB_SUCCESS,
                    })}
                  >
                    <img
                      alt={`Job ${job.id} ${
                        job.state === PipelineJobState.JOB_FAILURE
                          ? 'failed'
                          : 'succeeded'
                      }`}
                      className={styles.pipelineIcon}
                      src={
                        job.state === PipelineJobState.JOB_SUCCESS
                          ? '/dag_success.svg'
                          : '/dag_error.svg'
                      }
                    />
                    {job.pipelineName}
                  </NavLink>
                </Tooltip>
              </li>
            ))}
          </ol>
        </nav>

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

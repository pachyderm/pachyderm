import {JobSet} from '@graphqlTypes';
import {Tooltip, Icon} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {NavLink} from 'react-router-dom';

import {getJobStateIcon, getVisualJobState} from '@dash-frontend/lib/jobs';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import styles from './PipelineList.module.css';

interface PipelineListProps {
  containerRef: HTMLDivElement | null;
  jobSet?: JobSet;
  projectId: string;
  jobId: string;
}

const PipelineList: React.FC<PipelineListProps> = ({
  containerRef,
  jobSet,
  projectId,
  jobId,
}) => {
  const parentWidth = containerRef?.getBoundingClientRect().width;
  return (
    <nav
      className={classnames(styles.nav, {
        [styles.collapsed]: parentWidth ? parentWidth < 340 : false,
      })}
    >
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
                    [styles.busy]: jobVisualState === 'RUNNING',
                  })}
                >
                  <Icon small>{getJobStateIcon(jobVisualState)}</Icon>
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
  );
};

export default PipelineList;

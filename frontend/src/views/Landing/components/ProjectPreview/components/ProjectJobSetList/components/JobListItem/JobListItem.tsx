import classNames from 'classnames';
import React from 'react';
import {Link} from 'react-router-dom';

import {JobInfo, JobState} from '@dash-frontend/api/pps';
import {getDurationToNowFromISOString} from '@dash-frontend/lib/dateTime';
import {readableJobState} from '@dash-frontend/lib/jobs';
import {useSelectedRunRoute} from '@dash-frontend/views/Project/utils/routes';
import {ArrowRightSVG, Group, Icon} from '@pachyderm/components';

import styles from './JobListItem.module.css';

type JobListItemProps = {
  job: JobInfo;
  projectId: string;
};

const JobListItem: React.FC<JobListItemProps> = ({job, projectId}) => {
  const jobRoute = useSelectedRunRoute({
    projectId,
    jobId: job.job?.id,
    pipelineId: '',
  });

  return (
    <li data-testid="JobListItem__job">
      <Link
        to={jobRoute}
        className={classNames(styles.base, {
          [styles.error]: job.state === JobState.JOB_FAILURE,
        })}
      >
        <Group className={styles.innerContent} spacing={8}>
          <span
            className={classNames(styles.jobStatus, styles[job.state || ''])}
          >
            {job.state ? readableJobState(job.state) : null}
          </span>
          <span className={styles.timestamp}>
            {job.created
              ? `Created ${getDurationToNowFromISOString(job.created, true)}`
              : 'Creating...'}
          </span>

          <Icon>
            <ArrowRightSVG />
          </Icon>
        </Group>
      </Link>
    </li>
  );
};

export default JobListItem;

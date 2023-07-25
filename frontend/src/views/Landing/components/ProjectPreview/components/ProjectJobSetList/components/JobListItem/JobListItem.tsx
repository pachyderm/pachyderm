import {JobState, ProjectDetailsQuery} from '@graphqlTypes';
import classNames from 'classnames';
import React from 'react';
import {Link} from 'react-router-dom';

import {getDurationToNow} from '@dash-frontend/lib/dateTime';
import {readableJobState} from '@dash-frontend/lib/jobs';
import {useSelectedRunRoute} from '@dash-frontend/views/Project/utils/routes';
import {ArrowRightSVG, Group, Icon} from '@pachyderm/components';

import styles from './JobListItem.module.css';

type JobListItemProps = {
  job: ProjectDetailsQuery['projectDetails']['jobSets'][0];
  projectId: string;
};

const JobListItem: React.FC<JobListItemProps> = ({job, projectId}) => {
  const jobRoute = useSelectedRunRoute({
    projectId,
    jobId: job.id,
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
          <span className={classNames(styles.jobStatus, styles[job.state])}>
            {readableJobState(job.state)}
          </span>
          <span className={styles.timestamp}>
            {job.createdAt
              ? `Created ${getDurationToNow(job.createdAt, true)}`
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

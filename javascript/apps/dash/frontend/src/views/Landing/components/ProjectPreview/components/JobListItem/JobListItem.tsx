import {ArrowSVG, Tooltip} from '@pachyderm/components';
import classNames from 'classnames';
import {formatDistanceToNowStrict, fromUnixTime, format} from 'date-fns';
import React from 'react';
import {Link} from 'react-router-dom';

import readableJobState from '@dash-frontend/lib/readableJobState';
import {Job} from '@graphqlTypes';

import styles from './JobListItem.module.css';

type JobListItemProps = {
  job: Job;
};

const JobListItem: React.FC<JobListItemProps> = ({job}) => {
  return (
    <Link to="/" className={styles.base}>
      <Tooltip
        tooltipKey="Job Details"
        tooltipText={`See details for Job ID: ${job.id} Created: ${format(
          fromUnixTime(job.createdAt),
          'MM/dd/yyyy h:mmaaa',
        )}`}
      >
        <div className={styles.innerContent}>
          <span className={classNames(styles.jobStatus, styles[job.state])}>
            {readableJobState(job.state)}
          </span>
          <span className={styles.timestamp}>
            {`Created ${formatDistanceToNowStrict(fromUnixTime(job.createdAt), {
              addSuffix: true,
            })}`}
          </span>
          <ArrowSVG aria-hidden className={styles.icon} />
        </div>
      </Tooltip>
    </Link>
  );
};

export default JobListItem;

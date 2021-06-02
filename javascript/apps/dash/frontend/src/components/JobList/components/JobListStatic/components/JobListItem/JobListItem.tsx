import {ArrowSVG, ButtonLink, Tooltip} from '@pachyderm/components';
import classNames from 'classnames';
import {formatDistanceToNowStrict, fromUnixTime, format} from 'date-fns';
import React from 'react';
import {Link} from 'react-router-dom';

import readableJobState from '@dash-frontend/lib/readableJobState';
import {PipelineJobOverviewFragment} from '@graphqlTypes';

import styles from './JobListItem.module.css';

type JobListItemProps = {
  pipelineJob: PipelineJobOverviewFragment;
  expandActions?: boolean;
};

const JobListItem: React.FC<JobListItemProps> = ({
  pipelineJob,
  expandActions = false,
}) => {
  return (
    <li>
      <Link
        to="/"
        className={classNames(styles.base, {
          [styles.expandActions]: expandActions,
        })}
      >
        <Tooltip
          tooltipKey="Job Details"
          tooltipText={`See details for Job ID: ${
            pipelineJob.id
          } Created: ${format(
            fromUnixTime(pipelineJob.createdAt),
            'MM/dd/yyyy h:mmaaa',
          )}`}
        >
          <div className={styles.innerContent}>
            <span
              className={classNames(
                styles.jobStatus,
                styles.innerContentItem,
                styles[pipelineJob.state],
              )}
            >
              {readableJobState(pipelineJob.state)}
            </span>
            <span
              className={classNames(styles.timestamp, styles.innerContentItem)}
            >
              {`Created ${formatDistanceToNowStrict(
                fromUnixTime(pipelineJob.createdAt),
                {
                  addSuffix: true,
                },
              )}`}
            </span>

            {expandActions ? (
              <>
                <ButtonLink className={styles.innerContentItem}>
                  Read logs
                </ButtonLink>
                <ButtonLink className={styles.innerContentItem}>
                  See Details
                </ButtonLink>
              </>
            ) : (
              <ArrowSVG
                aria-hidden
                className={classNames(styles.icon, styles.innerContentItem)}
              />
            )}
          </div>
        </Tooltip>
      </Link>
    </li>
  );
};

export default JobListItem;

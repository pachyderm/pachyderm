import {
  JobOverviewFragment,
  JobSetFieldsFragment,
  JobState,
} from '@graphqlTypes';
import {ArrowRightSVG, ButtonLink, Tooltip} from '@pachyderm/components';
import classNames from 'classnames';
import {formatDistanceToNowStrict, fromUnixTime, format} from 'date-fns';
import React from 'react';
import {Link} from 'react-router-dom';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import readableJobState from '@dash-frontend/lib/readableJobState';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import isPipelineJob from '../../utils/isPipelineJob';

import styles from './JobListItem.module.css';

type JobListItemProps = {
  job: JobOverviewFragment | JobSetFieldsFragment;
  projectId: string;
  expandActions?: boolean;
  cardStyle?: boolean;
};

const JobListItem: React.FC<JobListItemProps> = ({
  job,
  projectId,
  expandActions = false,
  cardStyle,
}) => {
  const {jobId} = useUrlState();

  return (
    <li>
      <Link
        to={jobRoute({
          projectId,
          jobId: job.id,
          pipelineId: isPipelineJob(job) ? job.pipelineName : undefined,
        })}
        className={classNames(styles.base, {
          [styles.expandActions]: expandActions,
          [styles.cardStyle]: cardStyle,
          [styles.selected]: job.id === jobId,
          [styles.error]: job.state === JobState.JOB_FAILURE,
        })}
      >
        <Tooltip
          disabled={cardStyle}
          tooltipKey="Job Details"
          tooltipText={`See details for Job ID: ${job.id} \n ${
            job.createdAt
              ? `Created: ${format(
                  fromUnixTime(job.createdAt),
                  'MM/dd/yyyy h:mmaaa',
                )}`
              : 'Creating...'
          }`}
        >
          <div className={styles.innerContent}>
            <span
              className={classNames(
                styles.jobStatus,
                styles.innerContentItem,
                styles[job.state],
              )}
            >
              {readableJobState(job.state)}
            </span>
            <span
              className={classNames(styles.timestamp, styles.innerContentItem)}
            >
              {cardStyle && (
                <b className={styles.jobId}>{job.id.slice(0, 8)}</b>
              )}
              {job.createdAt
                ? `Created ${formatDistanceToNowStrict(
                    fromUnixTime(job.createdAt),
                    {
                      addSuffix: true,
                    },
                  )}`
                : 'Creating...'}
            </span>

            {expandActions ? (
              <>
                <ButtonLink className={styles.innerContentItem}>
                  See Details
                </ButtonLink>
              </>
            ) : (
              !cardStyle && (
                <ArrowRightSVG
                  aria-hidden
                  className={classNames(styles.icon, styles.innerContentItem)}
                />
              )
            )}
          </div>
        </Tooltip>
      </Link>
    </li>
  );
};

export default JobListItem;

import {ArrowSVG, ButtonLink, Tooltip} from '@pachyderm/components';
import classNames from 'classnames';
import {formatDistanceToNowStrict, fromUnixTime, format} from 'date-fns';
import React from 'react';
import {Link} from 'react-router-dom';

import readableJobState from '@dash-frontend/lib/readableJobState';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';
import {JobOverviewFragment, JobSetFieldsFragment} from '@graphqlTypes';

import isPipelineJob from '../../utils/isPipelineJob';

import styles from './JobListItem.module.css';

type JobListItemProps = {
  job: JobOverviewFragment | JobSetFieldsFragment;
  projectId: string;
  expandActions?: boolean;
};

const JobListItem: React.FC<JobListItemProps> = ({
  job,
  projectId,
  expandActions = false,
}) => {
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
        })}
      >
        <Tooltip
          tooltipKey="Job Details"
          tooltipText={`See details for Job ID: ${job.id} Created: ${format(
            fromUnixTime(job.createdAt),
            'MM/dd/yyyy h:mmaaa',
          )}`}
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
              {`Created ${formatDistanceToNowStrict(
                fromUnixTime(job.createdAt),
                {
                  addSuffix: true,
                },
              )}`}
            </span>

            {expandActions ? (
              <>
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

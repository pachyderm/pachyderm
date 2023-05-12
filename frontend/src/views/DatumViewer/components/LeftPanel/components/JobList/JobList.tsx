import {JobOverviewFragment, JobSetFieldsFragment, Job} from '@graphqlTypes';
import classNames from 'classnames';
import React, {useCallback} from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ListItem from '@dash-frontend/components/ListItem';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  getJobStateColor,
  getJobStateSVG,
  getVisualJobState,
} from '@dash-frontend/lib/jobs';
import useDatumPath from '@dash-frontend/views/DatumViewer/hooks/useDatumPath';
import {CaretRightSVG, LoadingDots} from '@pachyderm/components';

import styles from './JobList.module.css';

export type jobListProps = {
  jobs?: (JobOverviewFragment | JobSetFieldsFragment)[];
  loading: boolean;
  isExpanded: boolean;
  setIsExpanded: React.Dispatch<React.SetStateAction<boolean>>;
  currentJob?: Job;
};

const JobList: React.FC<jobListProps> = ({
  jobs,
  loading,
  isExpanded,
  setIsExpanded,
  currentJob,
}) => {
  const {urlJobId, currentDatumId, updateSelectedJob} = useDatumPath();

  const deriveState = useCallback(
    (jobId: string) => {
      if ([urlJobId, currentJob?.id].includes(jobId) && !currentDatumId) {
        return 'selected';
      } else if ([urlJobId, currentJob?.id].includes(jobId) && currentDatumId) {
        return 'highlighted';
      }
      return 'default';
    },
    [currentDatumId, currentJob?.id, urlJobId],
  );

  const onClick = useCallback(
    (jobId: string) => {
      updateSelectedJob(jobId);
      setIsExpanded(true);
    },
    [setIsExpanded, updateSelectedJob],
  );

  if (!loading && !jobs) {
    return <LoadingDots />;
  }

  if (!loading && jobs?.length === 0) {
    return <EmptyState title="No jobs found for this pipeline." />;
  }

  return (
    <div
      data-testid="JobList__list"
      className={classNames(styles.base, {
        [styles.expanded]: isExpanded,
      })}
    >
      {jobs &&
        jobs.map((job) => (
          <ListItem
            data-testid="JobList__listItem"
            key={job.id}
            state={deriveState(job.id)}
            text={job.createdAt ? getStandardDate(job.createdAt) : 'N/A'}
            LeftIconSVG={
              getJobStateSVG(getVisualJobState(job.state)) || undefined
            }
            leftIconColor={
              getJobStateColor(getVisualJobState(job.state)) || undefined
            }
            RightIconSVG={CaretRightSVG}
            captionText={job.id}
            onClick={() => onClick(job.id)}
          />
        ))}
    </div>
  );
};

export default JobList;

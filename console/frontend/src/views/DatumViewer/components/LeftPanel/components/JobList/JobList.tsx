import classNames from 'classnames';
import React, {useCallback} from 'react';

import {JobInfo} from '@dash-frontend/api/pps';
import EmptyState from '@dash-frontend/components/EmptyState';
import ListItem from '@dash-frontend/components/ListItem';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {
  getJobStateColor,
  getJobStateSVG,
  getVisualJobState,
} from '@dash-frontend/lib/jobs';
import useDatumPath from '@dash-frontend/views/DatumViewer/hooks/useDatumPath';
import {CaretRightSVG, LoadingDots} from '@pachyderm/components';

import styles from './JobList.module.css';

export type jobListProps = {
  jobs?: JobInfo[];
  loading: boolean;
  isExpanded: boolean;
  setIsExpanded: React.Dispatch<React.SetStateAction<boolean>>;
  currentJob?: JobInfo;
};

const JobList: React.FC<jobListProps> = ({
  jobs,
  loading,
  isExpanded,
  setIsExpanded,
  currentJob,
}) => {
  const {urlJobId, currentDatumId, updateSelectedJob} = useDatumPath();
  const {isServiceOrSpout} = useCurrentPipeline();

  const deriveState = useCallback(
    (jobId: string) => {
      if ([urlJobId, currentJob?.job?.id].includes(jobId) && !currentDatumId) {
        return 'selected';
      } else if (
        [urlJobId, currentJob?.job?.id].includes(jobId) &&
        currentDatumId
      ) {
        return 'highlighted';
      }
      return 'default';
    },
    [currentDatumId, currentJob?.job?.id, urlJobId],
  );

  const onClick = useCallback(
    (jobId: string) => {
      updateSelectedJob(jobId);
      if (!isServiceOrSpout) {
        setIsExpanded(true);
      }
    },
    [isServiceOrSpout, setIsExpanded, updateSelectedJob],
  );

  if (loading || !jobs) {
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
            key={job?.job?.id}
            state={deriveState(job?.job?.id || '')}
            text={getStandardDateFromISOString(job.created)}
            LeftIconSVG={
              getJobStateSVG(getVisualJobState(job.state)) || undefined
            }
            leftIconColor={
              getJobStateColor(getVisualJobState(job.state)) || undefined
            }
            RightIconSVG={!isServiceOrSpout ? CaretRightSVG : undefined}
            captionText={job?.job?.id}
            onClick={() => onClick(job?.job?.id || '')}
          />
        ))}
    </div>
  );
};

export default JobList;

import classnames from 'classnames';
import React, {useCallback} from 'react';
import {useHistory} from 'react-router-dom';

import {JobInfo} from '@dash-frontend/api/pps';
import Sidebar from '@dash-frontend/components/Sidebar';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {calculateJobTotalRuntime} from '@dash-frontend/lib/dateTime';
import {CaptionTextSmall, Button} from '@pachyderm/components';

import styles from './DatumsSidebar.module.css';

type DatumsSidebarProps = {
  jobs?: JobInfo[];
  onClose: () => void;
  selectedJob: string;
  selectFailedDatumJob: React.Dispatch<React.SetStateAction<string>>;
};

const DatumsSidebar: React.FC<DatumsSidebarProps> = ({
  jobs,
  onClose,
  selectedJob,
  selectFailedDatumJob,
}) => {
  const browserHistory = useHistory();
  const {projectId} = useUrlState();
  const {getPathToDatumLogs} = useLogsNavigation();

  const inpectDatumsLink = useCallback(
    (job: JobInfo) => {
      const logsLink = getPathToDatumLogs(
        {
          projectId,
          jobId: job?.job?.id || '',
          pipelineId: job?.job?.pipeline?.name || '',
        },
        [],
      );
      browserHistory.push(logsLink);
    },
    [browserHistory, getPathToDatumLogs, projectId],
  );

  const handleSelection = (jobId: string) => {
    if (jobId === selectedJob) {
      selectFailedDatumJob('');
    } else {
      selectFailedDatumJob(jobId);
    }
  };

  return (
    <Sidebar className={styles.base} onClose={onClose} fixed>
      <h5 className={styles.title}>Jobs with failed datums</h5>
      {jobs?.map((job) => (
        <div
          key={job?.job?.id}
          className={classnames(styles.listItem, {
            [styles.selected]: job?.job?.id === selectedJob,
          })}
          onClick={() => handleSelection(job?.job?.id || '')}
          aria-label={`${job?.job?.id} failed datums`}
        >
          <CaptionTextSmall className={styles.listItemText}>
            {job?.job?.id?.slice(0, 6)}; @{job?.job?.pipeline?.name}
          </CaptionTextSmall>
          <CaptionTextSmall className={styles.listItemText}>
            {job.dataFailed} Failed{' '}
            {Number(job?.dataFailed) > 1 ? 'Datums' : 'Datum'};{' '}
            {calculateJobTotalRuntime(job, 'In Progress')} Runtime
          </CaptionTextSmall>
          <Button
            className={styles.button}
            buttonType="ghost"
            onClick={() => inpectDatumsLink(job)}
          >
            Inspect datums
          </Button>
        </div>
      ))}
    </Sidebar>
  );
};

export default DatumsSidebar;

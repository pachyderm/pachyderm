import {JobsQuery} from '@graphqlTypes';
import classnames from 'classnames';
import React, {useCallback} from 'react';
import {useHistory} from 'react-router-dom';

import Sidebar from '@dash-frontend/components/Sidebar';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getJobRuntime} from '@dash-frontend/lib/jobs';
import {CaptionTextSmall, Button} from '@pachyderm/components';

import styles from './DatumsSidebar.module.css';

type DatumsSidebarProps = {
  jobs?: JobsQuery['jobs']['items'];
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
    (job: JobsQuery['jobs']['items'][number]) => {
      const logsLink = getPathToDatumLogs(
        {
          projectId,
          jobId: job.id,
          pipelineId: job.pipelineName,
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
          key={job.id}
          className={classnames(styles.listItem, {
            [styles.selected]: job.id === selectedJob,
          })}
          onClick={() => handleSelection(job.id)}
          aria-label={`${job.id} failed datums`}
        >
          <CaptionTextSmall className={styles.listItemText}>
            {job.id.slice(0, 6)}; @{job.pipelineName}
          </CaptionTextSmall>
          <CaptionTextSmall className={styles.listItemText}>
            {job.dataFailed} Failed {job.dataFailed > 1 ? 'Datums' : 'Datum'};{' '}
            {getJobRuntime(job.createdAt, job.finishedAt)} Runtime
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

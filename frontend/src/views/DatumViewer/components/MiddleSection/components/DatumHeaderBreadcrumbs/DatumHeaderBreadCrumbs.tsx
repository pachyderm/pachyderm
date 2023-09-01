import React, {useCallback} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {logsViewerJobRoute} from '@dash-frontend/views/Project/utils/routes';
import {
  Button,
  CaptionText,
  CopySVG,
  Tooltip,
  useClipboardCopy,
  Link,
} from '@pachyderm/components';

import styles from './DatumHeaderBreadCrumbs.module.css';

type DatumHeaderBreadcrumbsProps = {
  jobId?: string;
};

const DatumHeaderBreadCrumbs: React.FC<DatumHeaderBreadcrumbsProps> = ({
  jobId,
}) => {
  const {projectId, pipelineId, datumId} = useUrlState();

  const jobPath = jobId
    ? logsViewerJobRoute({
        projectId,
        jobId,
        pipelineId,
      })
    : '';

  const inputString = datumId
    ? `${pipelineId}@${jobId} ${datumId}`
    : `${pipelineId}@${jobId}`;

  const tooltipText = datumId ? 'Copy path with Datum id' : 'Copy path';

  const {copy, reset} = useClipboardCopy(inputString);
  const handleCopy = useCallback(() => {
    copy();
    setTimeout(reset, 1000);
  }, [copy, reset]);

  return (
    <div className={styles.base}>
      <div className={styles.path} data-testid="DatumHeaderBreadCrumbs__path">
        <CaptionText color="black">Pipeline...</CaptionText>
        <CaptionText color="black">/</CaptionText>
        {jobId && datumId ? (
          <Link inline to={jobPath}>
            <CaptionText className={styles.link}>Job...</CaptionText>
          </Link>
        ) : (
          <CaptionText color="black">Job: {jobId}</CaptionText>
        )}
        {datumId && (
          <>
            <CaptionText color="black">/</CaptionText>
            <CaptionText color="black">Datum: {datumId}</CaptionText>
          </>
        )}
      </div>
      <Tooltip tooltipText={tooltipText}>
        <Button
          IconSVG={CopySVG}
          buttonType="ghost"
          color="black"
          onClick={handleCopy}
        />
      </Tooltip>
    </div>
  );
};

export default DatumHeaderBreadCrumbs;

import {JobQuery} from '@graphqlTypes';
import classnames from 'classnames';
import React from 'react';

import ConfigFilePreview from '@dash-frontend/components/ConfigFilePreview';
import Description from '@dash-frontend/components/Description';
import {
  CommitLink,
  PipelineLink,
  RepoLink,
} from '@dash-frontend/components/ResourceLink';
import RuntimeStats from '@dash-frontend/components/RuntimeStats';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';
import {
  getJobStateIcon,
  getVisualJobState,
  readableJobState,
} from '@dash-frontend/lib/jobs';
import ReadLogsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/ReadLogsButton';
import {
  ArrowRightSVG,
  CaptionTextSmall,
  Icon,
  Link,
  LoadingDots,
  SpinnerSVG,
  Tooltip,
} from '@pachyderm/components';

import {Chip} from '../Chip';

import useInfoPanel from './hooks/useInfoPanel';
import styles from './InfoPanel.module.css';

type InfoPanelProps = {
  showReadLogs?: boolean;
  lastPipelineJob?: JobQuery['job'];
  pipelineLoading?: boolean;
  disableGlobalFilter?: boolean;
  className?: string;
};

const InfoPanel: React.FC<InfoPanelProps> = ({
  showReadLogs = false,
  lastPipelineJob,
  pipelineLoading,
  disableGlobalFilter = false,
  className,
}) => {
  const {
    job,
    duration,
    jobDetails,
    loading,
    started,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
    logsDatumRoute,
    addLogsQueryParams,
  } = useInfoPanel({
    pipelineJob: lastPipelineJob,
    pipelineLoading: !!pipelineLoading,
    useGlobalFilter: !disableGlobalFilter,
  });

  return (
    <div
      className={classnames(styles.base, className)}
      data-testid="InfoPanel__description"
    >
      <div className={styles.section}>
        <div className={styles.stateHeader}>
          <div className={styles.state}>
            {job ? (
              <Icon>{getJobStateIcon(getVisualJobState(job.state))}</Icon>
            ) : null}
            <span data-testid="InfoPanel__state" className={styles.stateText}>
              <h6>{job?.state ? readableJobState(job?.state) : ''}</h6>
            </span>
          </div>
          {showReadLogs ? <ReadLogsButton omitIcon={true} /> : null}
        </div>
        {loading ? (
          <div data-testid="InfoPanel__loading" className={styles.datumCard}>
            <LoadingDots />
          </div>
        ) : (
          <div className={styles.datumCard}>
            <div className={styles.datumCardHeader}>
              <div className={styles.datumCardRow}>
                <div>
                  <span
                    data-testid="InfoPanel__total"
                    className={styles.number}
                  >
                    {job?.dataTotal}
                  </span>
                  <CaptionTextSmall>Total Datums</CaptionTextSmall>
                </div>
                <Tooltip
                  tooltipText="Pipeline Version"
                  tooltipKey="Pipeline Version"
                >
                  <div>v: {jobDetails.pipelineVersion}</div>
                </Tooltip>
              </div>
              {!job?.finishedAt && (
                <Chip
                  LeftIconSVG={SpinnerSVG}
                  LeftIconSmall={false}
                  isButton={false}
                  className={styles.datumsProcessing}
                >
                  Processing datums
                </Chip>
              )}
            </div>
            <div className={styles.datumCardBody}>
              {datumMetrics.map(({label, value, filter}) => (
                <Link
                  className={styles.datumCardMetric}
                  key={label}
                  to={addLogsQueryParams(logsDatumRoute, filter)}
                >
                  <div
                    data-testid={`InfoPanel__${label.toLowerCase()}`}
                    className={styles.number}
                  >
                    {value}
                    <Icon className={styles.arrow}>
                      <ArrowRightSVG />
                    </Icon>
                  </div>
                  <CaptionTextSmall className={styles.metricLabel}>
                    {label}
                  </CaptionTextSmall>
                </Link>
              ))}
            </div>
          </div>
        )}
        {job?.reason && (
          <Description
            term="Failure Reason"
            loading={loading}
            className={styles.inputs}
            data-testid="InfoPanel__reason"
          >
            {extractAndShortenIds(job.reason)}
          </Description>
        )}
      </div>
      <div className={styles.section}>
        <RuntimeStats
          loading={loading}
          started={started}
          duration={duration}
          runtimeMetrics={runtimeMetrics}
        />
      </div>
      <div className={styles.section}>
        <div className={styles.inputSection}>
          {inputs.length ? (
            <Description
              term="Input Repos"
              loading={loading}
              data-testid="InfoPanel__repos"
              className={styles.inputs}
            >
              {inputs.map(({projectId, name}) => (
                <RepoLink
                  key={`${projectId}_${name}`}
                  name={name}
                  projectId={projectId}
                />
              ))}
            </Description>
          ) : null}
        </div>
        <div className={styles.inputSection}>
          <Description
            term="Pipeline"
            loading={loading}
            className={styles.inputs}
          >
            {job?.pipelineName ? (
              <PipelineLink
                data-testid="InfoPanel__pipeline"
                name={job.pipelineName}
              />
            ) : null}
          </Description>
        </div>
        <div className={styles.inputSection}>
          {job?.outputCommit && job?.outputBranch ? (
            <Description
              term="Output Commit"
              loading={loading}
              data-testid="InfoPanel__output"
              className={styles.inputs}
            >
              <CommitLink
                data-testid="InfoPanel__commitLink"
                name={job.outputCommit}
                branch={job.outputBranch}
                repoName={job.pipelineName}
              />
            </Description>
          ) : null}
        </div>
      </div>
      {!loading && jobConfig ? (
        <ConfigFilePreview
          title="Job Definition"
          data-testid="InfoPanel__details"
          header={
            jobConfig?.transform?.debug ? (
              <Tooltip
                tooltipText="Debug logs return more detailed information and can be enabled by setting debug: true before the next run of this pipeline."
                tooltipKey="debugLogs"
              >
                <span className={styles.detailedTooltip}>Detailed</span>
              </Tooltip>
            ) : undefined
          }
          config={jobConfig}
        />
      ) : (
        <LoadingDots />
      )}
    </div>
  );
};

export default InfoPanel;

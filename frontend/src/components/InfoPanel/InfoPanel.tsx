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
import {
  ArrowRightSVG,
  CaptionTextSmall,
  Icon,
  Link,
  LoadingDots,
  SpinnerSVG,
  Tooltip,
  Button,
  SkeletonBodyText,
  ButtonGroup,
  StatusStopSVG,
  JobsSVG,
} from '@pachyderm/components';

import {Chip} from '../Chip';

import useInfoPanel from './hooks/useInfoPanel';
import styles from './InfoPanel.module.css';

type InfoPanelProps = {
  hideReadLogs?: boolean;
  className?: string;
};

const InfoPanel: React.FC<InfoPanelProps> = ({hideReadLogs, className}) => {
  const {
    job,
    totalRuntime,
    cumulativeTime,
    jobLoading,
    started,
    created,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
    logsDatumRoute,
    addLogsQueryParams,
    stopJob,
    stopJobLoading,
    canStopJob,
  } = useInfoPanel();

  return (
    <div
      className={classnames(styles.base, className)}
      data-testid="InfoPanel__description"
    >
      <section className={styles.section} aria-label="Job Details">
        {jobLoading ? (
          <SkeletonBodyText />
        ) : (
          <div className={styles.stateHeader}>
            <div className={styles.state}>
              {job && (
                <Icon>{getJobStateIcon(getVisualJobState(job.state))}</Icon>
              )}
              {job?.state && (
                <span
                  data-testid="InfoPanel__state"
                  className={styles.stateText}
                >
                  <h6>{readableJobState(job.state)}</h6>
                </span>
              )}
            </div>
            <ButtonGroup>
              {!hideReadLogs && (
                <Button
                  buttonType="ghost"
                  to={logsDatumRoute}
                  disabled={!logsDatumRoute}
                  IconSVG={JobsSVG}
                >
                  Inspect Job
                </Button>
              )}
              {canStopJob && (
                <Button
                  buttonType="ghost"
                  onClick={() =>
                    stopJob({
                      job: {
                        id: job?.job?.id,
                        pipeline: job?.job?.pipeline,
                      },
                      reason: 'job manually stopped from console',
                    })
                  }
                  disabled={stopJobLoading}
                  IconSVG={!stopJobLoading ? StatusStopSVG : SpinnerSVG}
                >
                  Stop Job
                </Button>
              )}
            </ButtonGroup>
          </div>
        )}
        {jobLoading ? (
          <aside data-testid="InfoPanel__loading" className={styles.datumCard}>
            <LoadingDots />
          </aside>
        ) : (
          <aside className={styles.datumCard}>
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
                <Tooltip tooltipText="Pipeline Version">
                  <div>v: {job?.pipelineVersion}</div>
                </Tooltip>
              </div>
              {!job?.finished && (
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
          </aside>
        )}
        {job?.reason && (
          <Description
            term="Failure Reason"
            loading={jobLoading}
            className={styles.inputs}
            data-testid="InfoPanel__reason"
          >
            {extractAndShortenIds(job.reason)}
          </Description>
        )}
      </section>
      <section className={styles.section} aria-label="Job Runtime Statistics">
        <RuntimeStats
          loading={jobLoading}
          started={started}
          created={created}
          totalRuntime={totalRuntime}
          cumulativeTime={cumulativeTime}
          runtimeMetrics={runtimeMetrics}
        />
      </section>
      <section
        className={styles.section}
        aria-label="Related Pipeline and Repo information"
      >
        <div className={styles.inputSection}>
          {inputs.length !== 0 && (
            <Description
              term="Input Repos"
              loading={jobLoading}
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
          )}
        </div>
        <div className={styles.inputSection}>
          <Description
            term="Pipeline"
            loading={jobLoading}
            className={styles.inputs}
          >
            {job?.job?.pipeline?.name && (
              <PipelineLink
                data-testid="InfoPanel__pipeline"
                name={job?.job?.pipeline?.name}
              />
            )}
          </Description>
        </div>
        <div className={styles.inputSection}>
          {job?.outputCommit?.id && job?.outputCommit?.branch?.name && (
            <Description
              term="Output Commit"
              loading={jobLoading}
              data-testid="InfoPanel__output"
              className={styles.inputs}
            >
              <CommitLink
                data-testid="InfoPanel__commitLink"
                name={job?.outputCommit?.id}
                repoName={job?.job?.pipeline?.name || ''}
              />
            </Description>
          )}
        </div>
      </section>
      {!jobLoading && jobConfig ? (
        <ConfigFilePreview
          title="Job Definition"
          data-testid="InfoPanel__details"
          header={
            jobConfig?.transform?.debug ? (
              <Tooltip tooltipText="Debug logs return more detailed information and can be enabled by setting debug: true before the next run of this pipeline.">
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

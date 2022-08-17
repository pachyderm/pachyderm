import formatBytes from '@dash-backend/lib/formatBytes';
import {
  ChevronDownSVG,
  ChevronUpSVG,
  Icon,
  LoadingDots,
  Tooltip,
  CaptionTextSmall,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import ConfigFilePreview from '@dash-frontend/components/ConfigFilePreview';
import Description from '@dash-frontend/components/Description';
import {
  RepoLink,
  PipelineLink,
  CommitLink,
} from '@dash-frontend/components/ResourceLink';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';
import {
  readableJobState,
  getJobStateHref,
  getVisualJobState,
} from '@dash-frontend/lib/jobs';
import ReadLogsButton from '@dash-frontend/views/Project/components/ProjectSidebar/components/ReadLogsButton';

import useInfoPanel from './hooks/useInfoPanel';
import styles from './InfoPanel.module.css';

type InfoPanelProps = {
  showReadLogs?: boolean;
};

const InfoPanel: React.FC<InfoPanelProps> = ({showReadLogs = false}) => {
  const {
    job,
    duration,
    jobDetails,
    loading,
    started,
    runtimeDetailsOpen,
    runtimeDetailsClosing,
    toggleRunTimeDetailsOpen,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
  } = useInfoPanel();

  return (
    <div className={styles.base} data-testid="InfoPanel__description">
      <div className={styles.section}>
        <div className={styles.stateHeader}>
          <div className={styles.state}>
            {job ? (
              <img
                alt={`Pipeline job ${job.id} ${readableJobState(job.state)}:`}
                className={styles.stateIcon}
                src={getJobStateHref(getVisualJobState(job.state))}
              />
            ) : null}
            <span data-testid="InfoPanel__state">
              <h6>{job?.state ? readableJobState(job?.state) : ''}</h6>
            </span>
          </div>
          {showReadLogs ? <ReadLogsButton isButton={true} /> : null}
        </div>
        {loading ? (
          <div data-testid="InfoPanel__loading" className={styles.datumCard}>
            <LoadingDots />
          </div>
        ) : (
          <div className={styles.datumCard}>
            <div className={styles.datumCardHeader}>
              <div>
                <span data-testid="InfoPanel__total" className={styles.number}>
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
            <div className={styles.datumCardBody}>
              {datumMetrics.map(({label, value}) => (
                <div className={styles.datumCardMetric} key={label}>
                  <div
                    data-testid={`InfoPanel__${label.toLowerCase()}`}
                    className={styles.number}
                  >
                    {value}
                  </div>
                  <CaptionTextSmall className={styles.metricLabel}>
                    {label}
                  </CaptionTextSmall>
                </div>
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
        <div className={styles.metricSection}>
          <Description
            term="Start"
            loading={loading}
            className={styles.duration}
            data-testid="InfoPanel__started"
          >
            {started ? started : 'N/A'}
          </Description>
        </div>
        <div className={styles.metricSection}>
          <Description
            term="Runtime"
            loading={loading}
            data-testid="InfoPanel__duration"
            className={classnames(styles.duration, {
              [styles.runtime]: !loading,
            })}
            onClick={toggleRunTimeDetailsOpen}
          >
            {duration ? duration : 'N/A'}
            <Icon>
              {runtimeDetailsOpen ? <ChevronUpSVG /> : <ChevronDownSVG />}
            </Icon>
          </Description>
        </div>
        {runtimeDetailsOpen ? (
          <div
            className={classnames(
              {[styles.closing]: runtimeDetailsClosing},
              styles.runtimeDetails,
            )}
            data-testid="InfoPanel__durationDetails"
          >
            {runtimeMetrics.map(({duration, bytes, label}) => (
              <div className={styles.runtimeDetail} key={label}>
                <div>
                  <div className={styles.duration}>{duration}</div>
                  {bytes ? (
                    <CaptionTextSmall className={styles.bytes}>
                      {formatBytes(bytes)}
                    </CaptionTextSmall>
                  ) : null}
                </div>
                <div>
                  <div className={styles.metricLabel}>{label}</div>
                </div>
              </div>
            ))}
          </div>
        ) : null}
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
              {inputs.map((input) => (
                <RepoLink key={input} name={input} />
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

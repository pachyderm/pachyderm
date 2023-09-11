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
    jobDetails,
    jobLoading,
    started,
    created,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
    logsDatumRoute,
    addLogsQueryParams,
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
            {!hideReadLogs && (
              <Button
                buttonType="ghost"
                to={logsDatumRoute}
                disabled={!logsDatumRoute}
              >
                Inspect Job
              </Button>
            )}
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
            {job?.pipelineName && (
              <PipelineLink
                data-testid="InfoPanel__pipeline"
                name={job.pipelineName}
              />
            )}
          </Description>
        </div>
        <div className={styles.inputSection}>
          {job?.outputCommit && job?.outputBranch && (
            <Description
              term="Output Commit"
              loading={jobLoading}
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

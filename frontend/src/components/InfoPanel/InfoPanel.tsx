import classnames from 'classnames';
import parse from 'parse-duration';
import React from 'react';
import {Route, Switch} from 'react-router';

import ConfigFilePreview from '@dash-frontend/components/ConfigFilePreview';
import Description from '@dash-frontend/components/Description';
import {
  CommitLink,
  PipelineLink,
  RepoLink,
} from '@dash-frontend/components/ResourceLink';
import RuntimeStats from '@dash-frontend/components/RuntimeStats';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';
import {
  getJobStateIcon,
  getVisualJobState,
  readableJobState,
} from '@dash-frontend/lib/jobs';
import {LINEAGE_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
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
} from '@pachyderm/components';

import {Chip} from '../Chip';
import ContentBox from '../ContentBox';
import GlobalIdCopy from '../GlobalIdCopy';
import GlobalIdDisclaimer from '../GlobalIdDisclaimer';

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
    globalId,
  } = useInfoPanel();

  const {pipeline} = useCurrentPipeline();

  return (
    <div
      className={classnames(styles.base, className)}
      data-testid="InfoPanel__description"
    >
      <section className={styles.section} aria-label="Job Details">
        {jobLoading ? (
          <SkeletonBodyText />
        ) : (
          <>
            <div className={styles.stateHeader}>
              Subjob Summary
              <ButtonGroup>
                {!hideReadLogs && (
                  <Button
                    buttonType="ghost"
                    to={logsDatumRoute}
                    disabled={!logsDatumRoute}
                  >
                    Inspect Subjob
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
            <Switch>
              <Route path={LINEAGE_PIPELINE_PATH} exact>
                <GlobalIdDisclaimer
                  globalId={globalId}
                  artifact="subjob"
                  startTime={job?.created}
                />
              </Route>
            </Switch>
          </>
        )}
        {job?.reason && (
          <ContentBox
            loading={jobLoading}
            content={[
              {
                label: 'Failure Reason',
                value: extractAndShortenIds(job.reason),
              },
            ]}
          />
        )}
        <ContentBox
          loading={jobLoading}
          content={[
            {
              label: 'Job ID',
              value: job ? <GlobalIdCopy id={job.job?.id || ''} /> : 'N/A',
              responsiveValue: job ? (
                <GlobalIdCopy id={job.job?.id || ''} shortenId />
              ) : (
                'N/A'
              ),
              dataTestId: 'InfoPanel__subjob_id',
            },
            {
              label: 'Max Restarts Allowed',
              value: pipeline?.details?.datumTries,
              dataTestId: 'InfoPanel__restarts_allowed',
            },
            {
              label: 'Job Timeout',
              value: pipeline?.details?.jobTimeout
                ? `${parse(pipeline.details?.jobTimeout, 's')} seconds`
                : 'N/A',
              dataTestId: 'InfoPanel__job_timeout',
            },
          ]}
        />
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
          {job?.outputCommit?.id && (
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

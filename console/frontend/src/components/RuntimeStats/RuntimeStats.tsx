import classNames from 'classnames';
import React from 'react';

import {DatumInfo, DatumState} from '@dash-frontend/api/pps';
import Description from '@dash-frontend/components/Description';
import GlobalIdCopy from '@dash-frontend/components/GlobalIdCopy';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {
  CaptionTextSmall,
  ChevronDownSVG,
  ChevronUpSVG,
  Icon,
} from '@pachyderm/components';

import {useRuntimeStats} from './hooks/useRuntimeStats';
import styles from './RuntimeStats.module.css';

type RuntimeStatsProps = {
  loading: boolean;
  started?: string;
  created?: string;
  totalRuntime?: string;
  cumulativeTime: string;
  runtimeMetrics: {
    label: string;
    duration: string;
    bytes: number | null | undefined;
  }[];
  datum?: DatumInfo | null;
};

const RuntimeStats = ({
  loading,
  started,
  created,
  totalRuntime,
  cumulativeTime,
  runtimeMetrics,
  datum,
}: RuntimeStatsProps) => {
  const {toggleRunTimeDetailsOpen, runtimeDetailsOpen, runtimeDetailsClosing} =
    useRuntimeStats();

  return (
    <>
      {datum?.state === DatumState.SKIPPED && (
        <>
          <h6 className={styles.header}>Previous Job</h6>
          <div className={styles.metricSection}>
            <Description
              term="Job ID"
              loading={loading}
              className={styles.jobIdContainer}
              data-testid="RuntimeStats__jobId"
            >
              <GlobalIdCopy id={datum.datum?.job?.id || ''} />
            </Description>
          </div>
        </>
      )}
      {created && (
        <div className={styles.metricSection}>
          <Description
            term="Created"
            loading={loading}
            className={styles.duration}
            data-testid="RuntimeStats__created"
          >
            {created}
          </Description>
        </div>
      )}
      {started && (
        <div className={styles.metricSection}>
          <Description
            term="Start"
            loading={loading}
            className={styles.duration}
            data-testid="RuntimeStats__started"
          >
            {started}
          </Description>
        </div>
      )}
      {totalRuntime && (
        <div className={styles.metricSection}>
          <Description
            loading={loading}
            className={classNames(styles.duration, {
              [styles.runtime]: !loading,
            })}
            aria-label="Runtime"
          >
            Runtime: {totalRuntime}
          </Description>
          <CaptionTextSmall>Finished time minus start time</CaptionTextSmall>
        </div>
      )}
      <div className={styles.metricSection}>
        <Description
          loading={loading}
          className={classNames(styles.duration, {
            [styles.runtime]: !loading,
          })}
          onClick={toggleRunTimeDetailsOpen}
          role="button"
          aria-label="Cumulative Time"
        >
          Cumulative Time: {cumulativeTime}
          <Icon>
            {runtimeDetailsOpen ? <ChevronUpSVG /> : <ChevronDownSVG />}
          </Icon>
        </Description>
        <CaptionTextSmall>Time spent actively working</CaptionTextSmall>
      </div>
      {runtimeDetailsOpen ? (
        <div
          className={classNames(
            {[styles.closing]: runtimeDetailsClosing},
            styles.runtimeDetails,
          )}
        >
          {runtimeMetrics.map(({duration, bytes, label}) => (
            <div className={styles.runtimeDetail} key={label}>
              <div>
                <dd className={styles.duration} aria-labelledby={label}>
                  {duration}
                </dd>
                {bytes ? (
                  <CaptionTextSmall className={styles.bytes}>
                    {formatBytes(bytes)}
                  </CaptionTextSmall>
                ) : null}
              </div>
              <dt id={label}>
                <div className={styles.metricLabel}>{label}</div>
              </dt>
            </div>
          ))}
        </div>
      ) : null}
    </>
  );
};

export default RuntimeStats;

import formatBytes from '@dash-backend/lib/formatBytes';
import classNames from 'classnames';
import React from 'react';

import Description from '@dash-frontend/components/Description';
import {
  CaptionTextSmall,
  ChevronDownSVG,
  ChevronUpSVG,
  Icon,
} from '@pachyderm/components';

import {CopiableField} from '../ShortId';

import {useRuntimeStats} from './hooks/useRuntimeStats';
import styles from './RuntimeStats.module.css';

const RuntimeStats = ({
  loading,
  started,
  created,
  totalRuntime,
  runtimeMetrics,
  previousJobId,
}: {
  loading: boolean;
  started?: string;
  created?: string;
  totalRuntime: string;
  runtimeMetrics: {
    label: string;
    duration: string;
    bytes: number | null | undefined;
  }[];
  previousJobId?: string | null;
}) => {
  const {toggleRunTimeDetailsOpen, runtimeDetailsOpen, runtimeDetailsClosing} =
    useRuntimeStats();

  return (
    <>
      {previousJobId && (
        <>
          <h6 className={styles.header}>Previous Job</h6>
          <div className={styles.metricSection}>
            <Description
              term="Job ID"
              loading={loading}
              className={styles.jobIdContainer}
              data-testid="RuntimeStats__jobId"
            >
              <CopiableField
                inputString={previousJobId}
                successCheckmarkAriaLabel="You have successfully copied the id"
              >
                <p className={styles.overflow}>{previousJobId}</p>
              </CopiableField>
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
      <div className={styles.metricSection}>
        <Description
          term="Runtime"
          loading={loading}
          className={classNames(styles.duration, {
            [styles.runtime]: !loading,
          })}
          onClick={toggleRunTimeDetailsOpen}
        >
          {totalRuntime}
          <Icon>
            {runtimeDetailsOpen ? <ChevronUpSVG /> : <ChevronDownSVG />}
          </Icon>
        </Description>
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

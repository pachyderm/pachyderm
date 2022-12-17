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
  duration,
  runtimeMetrics,
  previousJobId,
}: {
  loading: boolean;
  started?: string;
  duration: string;
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
          data-testid="RuntimeStats__duration"
          className={classNames(styles.duration, {
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
          className={classNames(
            {[styles.closing]: runtimeDetailsClosing},
            styles.runtimeDetails,
          )}
          data-testid="RuntimeStats__durationDetails"
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
    </>
  );
};

export default RuntimeStats;

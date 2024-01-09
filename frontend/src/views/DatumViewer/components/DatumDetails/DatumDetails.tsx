import classnames from 'classnames';
import React from 'react';

import {DatumState} from '@dash-frontend/api/pps';
import ConfigFilePreview from '@dash-frontend/components/ConfigFilePreview';
import RuntimeStats from '@dash-frontend/components/RuntimeStats';
import {getDatumStateIcon, readableDatumState} from '@dash-frontend/lib/datums';
import {Icon} from '@pachyderm/components';

import styles from './DatumDetails.module.css';
import {useDatumDetails} from './hooks/useDatumDetails';

type DatumDetailsProps = {
  className?: string;
};

const DatumDetails: React.FC<DatumDetailsProps> = ({className}) => {
  const {datum, cumulativeTime, loading, started, runtimeMetrics, inputSpec} =
    useDatumDetails();

  return (
    <div
      className={classnames(styles.base, className)}
      data-testid="DatumDetails__panel"
    >
      <div className={styles.stateHeader}>
        <div className={styles.state}>
          {datum && <Icon>{getDatumStateIcon(datum.state)}</Icon>}
          <span data-testid="DatumDetails__state" className={styles.stateText}>
            <h6>{readableDatumState(datum?.state ?? '')}</h6>
          </span>
        </div>
      </div>

      {datum?.state === DatumState.SKIPPED && (
        <div className={styles.section}>
          <p>
            This datum has been successfully processed in a previous job, has
            not changed since then, and therefore, it was skipped in the current
            job.
          </p>
        </div>
      )}

      <div className={styles.section}>
        <RuntimeStats
          datum={datum}
          loading={loading}
          started={started}
          cumulativeTime={cumulativeTime}
          runtimeMetrics={runtimeMetrics}
        />
      </div>
      <ConfigFilePreview title="Input Spec" config={inputSpec} />
    </div>
  );
};

export default DatumDetails;

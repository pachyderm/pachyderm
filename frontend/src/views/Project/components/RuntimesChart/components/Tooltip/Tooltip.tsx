import {Chart as ChartJS, Tooltip as TooltipChartJS} from 'chart.js';
import React from 'react';

import {CaptionTextSmall} from '@pachyderm/components';

import styles from './Tooltip.module.css';

ChartJS.register(TooltipChartJS);

type TooltipProps = {
  tooltipState: {
    top: number;
    left: number;
    opacity: number;
    runtime: string;
  };
  useHoursAsUnit: boolean;
};

const Tooltip: React.FC<TooltipProps> = ({tooltipState, useHoursAsUnit}) => {
  return (
    <div
      className={styles.base}
      style={{
        top: tooltipState.top,
        left: tooltipState.left,
        opacity: tooltipState.opacity,
      }}
    >
      <div className={styles.tooltipLine}>
        <CaptionTextSmall>
          {parseFloat(tooltipState.runtime).toFixed(1)}{' '}
          {useHoursAsUnit ? 'hours' : 'seconds'}
        </CaptionTextSmall>
        <CaptionTextSmall>Runtime</CaptionTextSmall>
      </div>
      <CaptionTextSmall className={styles.tooltipText}>
        Click to view datums
      </CaptionTextSmall>
    </div>
  );
};

export default Tooltip;
